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
	"testing"
)

func TestEnv(t *testing.T) {
	tests := []cmdTestCase{{
		name: "completion for env",
		// We disable descriptions by using __completeNoDesc because they would contain
		// OS-specific paths which will change depending on where the tests are run
		cmd:    "__completeNoDesc env ''",
		golden: "output/env-comp.txt",
	}, {
		name: "completion for env",
		cmd: "__complete env HELM_K	",
		golden: "output/env-desc-comp.txt",
	}}
	runTestCmd(t, tests)
}

func TestEnvFileCompletion(t *testing.T) {
	checkFileCompletion(t, "env", false)
	checkFileCompletion(t, "env HELM_BIN", false)
}
