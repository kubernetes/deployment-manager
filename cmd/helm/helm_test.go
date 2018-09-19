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

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/repo"
)

// releaseCmd is a command that works with a FakeClient
type releaseCmd func(c *helm.FakeClient, out io.Writer) *cobra.Command

// runReleaseCases runs a set of release cases through the given releaseCmd.
func runReleaseCases(t *testing.T, tests []releaseCase, rcmd releaseCmd) {
	var buf bytes.Buffer
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &helm.FakeClient{
				Rels:      tt.rels,
				Responses: tt.responses,
			}
			cmd := rcmd(c, &buf)
			cmd.ParseFlags(tt.flags)
			err := cmd.RunE(cmd, tt.args)
			if (err == nil) && tt.err {
				t.Errorf("expected error, got no error")
			}
			re := regexp.MustCompile(tt.expected)
			if !re.Match(buf.Bytes()) {
				t.Errorf("expected\n%q\ngot\n%q", tt.expected, buf.String())
			}
			buf.Reset()
		})
	}
}

// releaseCase describes a test case that works with releases.
type releaseCase struct {
	name  string
	args  []string
	flags []string
	// expected is the string to be matched. This supports regular expressions.
	expected string
	err      bool
	resp     *release.Release
	// Rels are the available releases at the start of the test.
	rels      []*release.Release
	responses map[string]release.TestRun_Status
}

// tempHelmHome sets up a Helm Home in a temp dir.
//
// This does not clean up the directory. You must do that yourself.
// You  must also set helmHome yourself.
func tempHelmHome(t *testing.T) (helmpath.Home, error) {
	oldhome := settings.Home
	dir, err := ioutil.TempDir("", "helm_home-")
	if err != nil {
		return helmpath.Home("n/"), err
	}

	settings.Home = helmpath.Home(dir)
	if err := ensureTestHome(settings.Home, t); err != nil {
		return helmpath.Home("n/"), err
	}
	settings.Home = oldhome
	return helmpath.Home(dir), nil
}

// ensureTestHome creates a home directory like ensureHome, but without remote references.
//
// t is used only for logging.
func ensureTestHome(home helmpath.Home, t *testing.T) error {
	configDirectories := []string{home.String(), home.Repository(), home.Cache(), home.LocalRepository(), home.Plugins(), home.Starters()}
	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			if err := os.MkdirAll(p, 0755); err != nil {
				return fmt.Errorf("Could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", p)
		}
	}

	repoFile := home.RepositoryFile()
	if fi, err := os.Stat(repoFile); err != nil {
		rf := repo.NewRepoFile()
		rf.Add(&repo.Entry{
			Name:  "charts",
			URL:   "http://example.com/foo",
			Cache: "charts-index.yaml",
		}, &repo.Entry{
			Name:  "local",
			URL:   "http://localhost.com:7743/foo",
			Cache: "local-index.yaml",
		})
		if err := rf.WriteFile(repoFile, 0644); err != nil {
			return err
		}
	} else if fi.IsDir() {
		return fmt.Errorf("%s must be a file, not a directory", repoFile)
	}
	if r, err := repo.LoadRepositoriesFile(repoFile); err == repo.ErrRepoOutOfDate {
		t.Log("Updating repository file format...")
		if err := r.WriteFile(repoFile, 0644); err != nil {
			return err
		}
	}

	localRepoIndexFile := home.LocalRepository(localRepositoryIndexFile)
	if fi, err := os.Stat(localRepoIndexFile); err != nil {
		i := repo.NewIndexFile()
		if err := i.WriteFile(localRepoIndexFile, 0644); err != nil {
			return err
		}

		//TODO: take this out and replace with helm update functionality
		os.Symlink(localRepoIndexFile, home.CacheIndex("local"))
	} else if fi.IsDir() {
		return fmt.Errorf("%s must be a file, not a directory", localRepoIndexFile)
	}

	t.Logf("$HELM_HOME has been configured at %s.\n", settings.Home.String())
	return nil

}

func TestRootCmd(t *testing.T) {
	cleanup := resetEnv()
	defer cleanup()

	tests := []struct {
		name   string
		args   []string
		envars map[string]string
		home   string
	}{
		{
			name: "defaults",
			args: []string{"home"},
			home: filepath.Join(os.Getenv("HOME"), "/.helm"),
		},
		{
			name: "with --home set",
			args: []string{"--home", "/foo"},
			home: "/foo",
		},
		{
			name: "subcommands with --home set",
			args: []string{"home", "--home", "/foo"},
			home: "/foo",
		},
		{
			name:   "with $HELM_HOME set",
			args:   []string{"home"},
			envars: map[string]string{"HELM_HOME": "/bar"},
			home:   "/bar",
		},
		{
			name:   "subcommands with $HELM_HOME set",
			args:   []string{"home"},
			envars: map[string]string{"HELM_HOME": "/bar"},
			home:   "/bar",
		},
		{
			name:   "with $HELM_HOME and --home set",
			args:   []string{"home", "--home", "/foo"},
			envars: map[string]string{"HELM_HOME": "/bar"},
			home:   "/foo",
		},
	}

	// ensure not set locally
	os.Unsetenv("HELM_HOME")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.Unsetenv("HELM_HOME")

			for k, v := range tt.envars {
				os.Setenv(k, v)
			}

			cmd := newRootCmd(tt.args)
			cmd.SetOutput(ioutil.Discard)
			cmd.SetArgs(tt.args)
			cmd.Run = func(*cobra.Command, []string) {}
			if err := cmd.Execute(); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if settings.Home.String() != tt.home {
				t.Errorf("expected home %q, got %q", tt.home, settings.Home)
			}
			homeFlag := cmd.Flag("home").Value.String()
			homeFlag = os.ExpandEnv(homeFlag)
			if homeFlag != tt.home {
				t.Errorf("expected home %q, got %q", tt.home, homeFlag)
			}
		})
	}
}

func TestTLSFlags(t *testing.T) {
	cleanup := resetEnv()
	defer cleanup()

	homePath := os.Getenv("HELM_HOME")
	if homePath == "" {
		homePath = filepath.Join(os.Getenv("HOME"), ".helm")
	}

	home := helmpath.Home(homePath)

	tests := []struct {
		name     string
		args     []string
		envars   map[string]string
		settings environment.EnvSettings
	}{
		{
			name: "defaults",
			args: []string{"version", "-c"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name: "tls enable",
			args: []string{"version", "-c", "--tls"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               true,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name: "tls verify",
			args: []string{"version", "-c", "--tls-verify"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               true,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name: "tls servername",
			args: []string{"version", "-c", "--tls-hostname=foo"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "foo",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name: "tls cacert",
			args: []string{"version", "-c", "--tls-ca-cert=/foo"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           "/foo",
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name: "tls cert",
			args: []string{"version", "-c", "--tls-cert=/foo"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             "/foo",
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name: "tls key",
			args: []string{"version", "-c", "--tls-key=/foo"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              "/foo",
			},
		},
		{
			name:   "tls enable envvar",
			args:   []string{"version", "-c"},
			envars: map[string]string{"HELM_TLS_ENABLE": "true"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               true,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name:   "tls verify envvar",
			args:   []string{"version", "-c"},
			envars: map[string]string{"HELM_TLS_VERIFY": "true"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               true,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name:   "tls servername envvar",
			args:   []string{"version", "-c"},
			envars: map[string]string{"HELM_TLS_HOSTNAME": "foo"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "foo",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name:   "tls cacert envvar",
			args:   []string{"version", "-c"},
			envars: map[string]string{"HELM_TLS_CA_CERT": "/foo"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           "/foo",
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name:   "tls cert envvar",
			args:   []string{"version", "-c"},
			envars: map[string]string{"HELM_TLS_CERT": "/foo"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             "/foo",
				TLSKeyFile:              home.TLSKey(),
			},
		},
		{
			name:   "tls key envvar",
			args:   []string{"version", "-c"},
			envars: map[string]string{"HELM_TLS_KEY": "/foo"},
			settings: environment.EnvSettings{
				TillerHost:              "",
				TillerConnectionTimeout: 300,
				TillerNamespace:         "kube-system",
				Home:                    home,
				Debug:                   false,
				KubeContext:             "",
				KubeConfig:              "",
				TLSEnable:               false,
				TLSVerify:               false,
				TLSServerName:           "",
				TLSCaCertFile:           home.TLSCaCert(),
				TLSCertFile:             home.TLSCert(),
				TLSKeyFile:              "/foo",
			},
		},
	}

	// ensure not set locally
	tlsEnvvars := []string{
		"HELM_TLS_HOSTNAME",
		"HELM_TLS_CA_CERT",
		"HELM_TLS_CERT",
		"HELM_TLS_KEY",
		"HELM_TLS_VERIFY",
		"HELM_TLS_ENABLE",
	}

	for i := range tlsEnvvars {
		os.Unsetenv(tlsEnvvars[i])
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for k, v := range tt.envars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cmd := newRootCmd(tt.args)
			cmd.SetOutput(ioutil.Discard)
			cmd.SetArgs(tt.args)
			cmd.Run = func(*cobra.Command, []string) {}
			if err := cmd.Execute(); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if settings != tt.settings {
				t.Errorf("expected settings %v, got %v", tt.settings, settings)
			}
		})
	}
}

func resetEnv() func() {
	origSettings := settings
	origEnv := os.Environ()
	return func() {
		settings = origSettings
		for _, pair := range origEnv {
			kv := strings.SplitN(pair, "=", 2)
			os.Setenv(kv[0], kv[1])
		}
	}
}
