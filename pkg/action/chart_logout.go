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

package action

import (
	"io"
)

// ChartLogout performs a chart login operation.
type ChartLogout struct {
	cfg *Configuration
}

// NewChartLogout creates a new ChartLogout object with the given configuration.
func NewChartLogout(cfg *Configuration) *ChartLogout {
	return &ChartLogout{
		cfg: cfg,
	}
}

// Run executes the chart logout operation
func (a *ChartLogout) Run(out io.Writer, hostname string) error {
	return a.cfg.RegistryClient.Logout(hostname)
}
