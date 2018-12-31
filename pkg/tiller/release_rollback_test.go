/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tiller

import (
	"strings"
	"testing"

	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
)

func TestRollbackRelease(t *testing.T) {
	c := helm.NewContext()
	rs := rsFixture()
	rel := releaseStub()
	rs.env.Releases.Create(rel)
	upgradedRel := upgradeReleaseVersion(rel)
	upgradedRel.Hooks = []*release.Hook{
		{
			Name:     "test-cm",
			Kind:     "ConfigMap",
			Path:     "test-cm",
			Manifest: manifestWithRollbackHooks,
			Events: []release.Hook_Event{
				release.Hook_PRE_ROLLBACK,
				release.Hook_POST_ROLLBACK,
			},
		},
	}

	upgradedRel.Manifest = "hello world"
	rs.env.Releases.Update(rel)
	rs.env.Releases.Create(upgradedRel)

	req := &services.RollbackReleaseRequest{
		Name: rel.Name,
	}
	res, err := rs.RollbackRelease(c, req)
	if err != nil {
		t.Fatalf("Failed rollback: %s", err)
	}

	if res.Release.Name == "" {
		t.Errorf("Expected release name.")
	}

	if res.Release.Name != rel.Name {
		t.Errorf("Updated release name does not match previous release name. Expected %s, got %s", rel.Name, res.Release.Name)
	}

	if res.Release.Namespace != rel.Namespace {
		t.Errorf("Expected release namespace '%s', got '%s'.", rel.Namespace, res.Release.Namespace)
	}

	if res.Release.Version != 3 {
		t.Errorf("Expected release version to be %v, got %v", 3, res.Release.Version)
	}

	updated, err := rs.env.Releases.Get(res.Release.Name, res.Release.Version)
	if err != nil {
		t.Errorf("Expected release for %s (%v).", res.Release.Name, rs.env.Releases)
	}

	if len(updated.Hooks) != 2 {
		t.Fatalf("Expected 2 hooks, got %d", len(updated.Hooks))
	}

	if updated.Hooks[0].Manifest != manifestWithHook {
		t.Errorf("Unexpected manifest: %v", updated.Hooks[0].Manifest)
	}

	anotherUpgradedRelease := upgradeReleaseVersion(upgradedRel)
	rs.env.Releases.Update(upgradedRel)
	rs.env.Releases.Create(anotherUpgradedRelease)

	res, err = rs.RollbackRelease(c, req)
	if err != nil {
		t.Fatalf("Failed rollback: %s", err)
	}

	updated, err = rs.env.Releases.Get(res.Release.Name, res.Release.Version)
	if err != nil {
		t.Errorf("Expected release for %s (%v).", res.Release.Name, rs.env.Releases)
	}

	if len(updated.Hooks) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(updated.Hooks))
	}

	if updated.Hooks[0].Manifest != manifestWithRollbackHooks {
		t.Errorf("Unexpected manifest: %v", updated.Hooks[0].Manifest)
	}

	if res.Release.Version != 4 {
		t.Errorf("Expected release version to be %v, got %v", 3, res.Release.Version)
	}

	if updated.Hooks[0].Events[0] != release.Hook_PRE_ROLLBACK {
		t.Errorf("Expected event 0 to be pre rollback")
	}

	if updated.Hooks[0].Events[1] != release.Hook_POST_ROLLBACK {
		t.Errorf("Expected event 1 to be post rollback")
	}

	if len(res.Release.Manifest) == 0 {
		t.Errorf("No manifest returned: %v", res.Release)
	}

	if len(updated.Manifest) == 0 {
		t.Errorf("Expected manifest in %v", res)
	}

	if !strings.Contains(updated.Manifest, "hello world") {
		t.Errorf("unexpected output: %s", rel.Manifest)
	}

	if res.Release.Info.Description != "Rollback to 2" {
		t.Errorf("Expected rollback to 2, got %q", res.Release.Info.Description)
	}
}

func TestRollbackWithReleaseVersion(t *testing.T) {
	c := helm.NewContext()
	rs := rsFixture()
	rs.Log = t.Logf
	rs.env.Releases.Log = t.Logf
	rel2 := releaseStub()
	rel2.Name = "other"
	rs.env.Releases.Create(rel2)
	rel := releaseStub()
	rs.env.Releases.Create(rel)
	v2 := upgradeReleaseVersion(rel)
	rs.env.Releases.Update(rel)
	rs.env.Releases.Create(v2)
	v3 := upgradeReleaseVersion(v2)
	// retain the original release as DEPLOYED while the update should fail
	v2.Info.Status.Code = release.Status_DEPLOYED
	v3.Info.Status.Code = release.Status_FAILED
	rs.env.Releases.Update(v2)
	rs.env.Releases.Create(v3)

	req := &services.RollbackReleaseRequest{
		Name:         rel.Name,
		DisableHooks: true,
		Version:      1,
	}

	_, err := rs.RollbackRelease(c, req)
	if err != nil {
		t.Fatalf("Failed rollback: %s", err)
	}
	// check that v2 is now in a SUPERSEDED state
	oldRel, err := rs.env.Releases.Get(rel.Name, 2)
	if err != nil {
		t.Fatalf("Failed to retrieve v2: %s", err)
	}
	if oldRel.Info.Status.Code != release.Status_SUPERSEDED {
		t.Errorf("Expected v2 to be in a SUPERSEDED state, got %q", oldRel.Info.Status.Code)
	}
	// make sure we didn't update some other deployments.
	otherRel, err := rs.env.Releases.Get(rel2.Name, 1)
	if err != nil {
		t.Fatalf("Failed to retrieve other v1: %s", err)
	}
	if otherRel.Info.Status.Code != release.Status_DEPLOYED {
		t.Errorf("Expected other deployed release to stay untouched, got %q", otherRel.Info.Status.Code)
	}
}

func TestRollbackLatestSuperseded(t *testing.T) {
	c := helm.NewContext()
	rs := rsFixture()

	rel := releaseStub()
	rel.Version = 1
	rel.Info.Status.Code = release.Status_SUPERSEDED
	rs.env.Releases.Create(rel)
	v2 := upgradeReleaseVersion(rel)
	v2.Version = 2
	v2.Info.Status.Code = release.Status_FAILED
	rs.env.Releases.Create(v2)
	v3 := upgradeReleaseVersion(v2)
	v3.Info.Status.Code = release.Status_DEPLOYED
	rs.env.Releases.Create(v3)

	req := &services.RollbackReleaseRequest{
		Name:         rel.Name,
		DisableHooks: true,
		Version:      0,
	}

	_, err := rs.RollbackRelease(c, req)
	if err != nil {
		t.Fatalf("Failed rollback: %s", err)
	}

	// check that v3 is now in a SUPERSEDED state
	oldRel, err := rs.env.Releases.Get(rel.Name, 3)
	if err != nil {
		t.Fatalf("Failed to retrieve v3: %s", err)
	}
	if oldRel.Info.Status.Code != release.Status_SUPERSEDED {
		t.Errorf("Expected v3 to be in a SUPERSEDED state, got %q", oldRel.Info.Status.Code)
	}

	lastRelease, _ := rs.env.Releases.Get(rel.Name, 4)
	statusCode := lastRelease.Info.Status.Code
	description := lastRelease.Info.Description

	if statusCode != release.Status_DEPLOYED {
		t.Errorf("Expected a new release created on deployed state, got: %s", statusCode)
	}
	if description != "Rollback to 1" {
		t.Errorf("Expected to rollback for release 1, got: %s", description)
	}
}

func TestRollbackDeleted(t *testing.T) {
	c := helm.NewContext()
	rs := rsFixture()
	rs.Log = t.Logf
	rs.env.Releases.Log = t.Logf
	rel2 := releaseStub()
	rel2.Name = "other"
	rs.env.Releases.Create(rel2)
	rel := releaseStub()
	rs.env.Releases.Create(rel)
	v2 := upgradeReleaseVersion(rel)
	rs.env.Releases.Update(rel)
	rs.env.Releases.Create(v2)
	v3 := upgradeReleaseVersion(v2)
	// retain the original release as DEPLOYED while the update should fail
	v2.Info.Status.Code = release.Status_DEPLOYED
	v3.Info.Status.Code = release.Status_FAILED
	rs.env.Releases.Update(v2)
	rs.env.Releases.Create(v3)

	req1 := &services.UninstallReleaseRequest{
		Name:         rel.Name,
		DisableHooks: true,
	}

	_, err := rs.UninstallRelease(c, req1)
	if err != nil {
		t.Fatalf("Failed uninstall: %s", err)
	}

	oldRel, err := rs.env.Releases.Get(rel.Name, 3)
	if err != nil {
		t.Fatalf("Failed to retrieve v3: %s", err)
	}
	if oldRel.Info.Status.Code != release.Status_DELETED {
		t.Errorf("Expected v3 to be in a DELETED state, got %q", oldRel.Info.Status.Code)
	}

	req2 := &services.RollbackReleaseRequest{
		Name:         rel.Name,
		DisableHooks: true,
		Version:      2,
	}

	_, err = rs.RollbackRelease(c, req2)
	if err != nil {
		t.Fatalf("Failed rollback: %s", err)
	}
	// check that v3 is now in a SUPERSEDED state
	oldRel, err = rs.env.Releases.Get(rel.Name, 3)
	if err != nil {
		t.Fatalf("Failed to retrieve v3: %s", err)
	}
	if oldRel.Info.Status.Code != release.Status_SUPERSEDED {
		t.Errorf("Expected v3 to be in a SUPERSEDED state, got %q", oldRel.Info.Status.Code)
	}
	// make sure we didn't update some other deployments.
	otherRel, err := rs.env.Releases.Get(rel2.Name, 1)
	if err != nil {
		t.Fatalf("Failed to retrieve other v1: %s", err)
	}
	if otherRel.Info.Status.Code != release.Status_DEPLOYED {
		t.Errorf("Expected other deployed release to stay untouched, got %q", otherRel.Info.Status.Code)
	}
}

func TestRollbackReleaseNoHooks(t *testing.T) {
	c := helm.NewContext()
	rs := rsFixture()
	rel := releaseStub()
	rel.Hooks = []*release.Hook{
		{
			Name:     "test-cm",
			Kind:     "ConfigMap",
			Path:     "test-cm",
			Manifest: manifestWithRollbackHooks,
			Events: []release.Hook_Event{
				release.Hook_PRE_ROLLBACK,
				release.Hook_POST_ROLLBACK,
			},
		},
	}
	rs.env.Releases.Create(rel)
	upgradedRel := upgradeReleaseVersion(rel)
	rs.env.Releases.Update(rel)
	rs.env.Releases.Create(upgradedRel)

	req := &services.RollbackReleaseRequest{
		Name:         rel.Name,
		DisableHooks: true,
	}

	res, err := rs.RollbackRelease(c, req)
	if err != nil {
		t.Fatalf("Failed rollback: %s", err)
	}

	if hl := res.Release.Hooks[0].LastRun; hl != nil {
		t.Errorf("Expected that no hooks were run. Got %d", hl)
	}
}

func TestRollbackReleaseFailure(t *testing.T) {
	c := helm.NewContext()
	rs := rsFixture()
	rel := releaseStub()
	rs.env.Releases.Create(rel)
	upgradedRel := upgradeReleaseVersion(rel)
	rs.env.Releases.Update(rel)
	rs.env.Releases.Create(upgradedRel)

	req := &services.RollbackReleaseRequest{
		Name:         rel.Name,
		DisableHooks: true,
	}

	rs.env.KubeClient = newUpdateFailingKubeClient()
	res, err := rs.RollbackRelease(c, req)
	if err == nil {
		t.Error("Expected failed rollback")
	}

	if targetStatus := res.Release.Info.Status.Code; targetStatus != release.Status_FAILED {
		t.Errorf("Expected FAILED release. Got %v", targetStatus)
	}

	oldRelease, err := rs.env.Releases.Get(rel.Name, rel.Version)
	if err != nil {
		t.Errorf("Expected to be able to get previous release")
	}
	if oldStatus := oldRelease.Info.Status.Code; oldStatus != release.Status_SUPERSEDED {
		t.Errorf("Expected SUPERSEDED status on previous Release version. Got %v", oldStatus)
	}
}

func TestRollbackReleaseWithCustomDescription(t *testing.T) {
	c := helm.NewContext()
	rs := rsFixture()
	rel := releaseStub()
	rs.env.Releases.Create(rel)
	upgradedRel := upgradeReleaseVersion(rel)
	rs.env.Releases.Update(rel)
	rs.env.Releases.Create(upgradedRel)

	customDescription := "foo"
	req := &services.RollbackReleaseRequest{
		Name:        rel.Name,
		Description: customDescription,
	}
	res, err := rs.RollbackRelease(c, req)
	if err != nil {
		t.Fatalf("Failed rollback: %s", err)
	}

	if res.Release.Name == "" {
		t.Errorf("Expected release name.")
	}

	if res.Release.Name != rel.Name {
		t.Errorf("Updated release name does not match previous release name. Expected %s, got %s", rel.Name, res.Release.Name)
	}

	if res.Release.Info.Description != customDescription {
		t.Errorf("Expected Description to be %q, got %q", customDescription, res.Release.Info.Description)
	}
}
