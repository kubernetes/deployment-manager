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

/*Package cli describes the operating environment for the Helm CLI.

Helm's environment encapsulates all of the service dependencies Helm has.
These dependencies are expressed as interfaces so that alternate implementations
(mocks, etc.) can be easily generated.
*/
package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"helm.sh/helm/v3/pkg/helmpath"
)

// defaultMaxHistory sets the maximum number of releases to 0: unlimited
const defaultMaxHistory = 10

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	namespace string
	config    *genericclioptions.ConfigFlags

	// KubeConfig is the path to the kubeconfig file
	KubeConfig string
	// KubeContext is the name of the kubeconfig context.
	KubeContext string
	// Bearer KubeToken used for authentication
	KubeToken string
	// Username to impersonate for the operation
	KubeAsUser string
	// Groups to impersonate for the operation, multiple groups parsed from a comma delimited list
	KubeAsGroups []string
	// Kubernetes API Server Endpoint for authentication
	KubeAPIServer string
	// Custom certificate authority file.
	KubeCaFile string
	// Debug indicates whether or not Helm is running in Debug mode.
	Debug bool
	// RegistryConfig is the path to the registry config file.
	RegistryConfig string
	// RepositoryConfig is the path to the repositories file.
	RepositoryConfig string
	// RepositoryCache is the path to the repository cache directory.
	RepositoryCache string
	// PluginsDirectory is the path to the plugins directory.
	PluginsDirectory string
	// MaxHistory is the max release history maintained.
	MaxHistory int
}

func New() *EnvSettings {
	env := &EnvSettings{
		namespace:        os.Getenv("HELM_NAMESPACE"),
		MaxHistory:       envIntOr("HELM_MAX_HISTORY", defaultMaxHistory),
		KubeContext:      os.Getenv("HELM_KUBECONTEXT"),
		KubeToken:        os.Getenv("HELM_KUBETOKEN"),
		KubeAsUser:       os.Getenv("HELM_KUBEASUSER"),
		KubeAsGroups:     envCSV("HELM_KUBEASGROUPS"),
		KubeAPIServer:    os.Getenv("HELM_KUBEAPISERVER"),
		KubeCaFile:       os.Getenv("HELM_KUBECAFILE"),
		PluginsDirectory: envOr("HELM_PLUGINS", helmpath.DataPath("plugins")),
		RegistryConfig:   envOr("HELM_REGISTRY_CONFIG", helmpath.ConfigPath("registry.json")),
		RepositoryConfig: envOr("HELM_REPOSITORY_CONFIG", helmpath.ConfigPath("repositories.yaml")),
		RepositoryCache:  envOr("HELM_REPOSITORY_CACHE", helmpath.CachePath("repository")),
	}
	env.Debug, _ = strconv.ParseBool(os.Getenv("HELM_DEBUG"))

	// bind to kubernetes config flags
	env.config = &genericclioptions.ConfigFlags{
		Namespace:        &env.namespace,
		Context:          &env.KubeContext,
		BearerToken:      &env.KubeToken,
		APIServer:        &env.KubeAPIServer,
		CAFile:           &env.KubeCaFile,
		KubeConfig:       &env.KubeConfig,
		Impersonate:      &env.KubeAsUser,
		ImpersonateGroup: &env.KubeAsGroups,
	}
	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&s.namespace, "namespace", "n", s.namespace, "namespace scope for this request")
	fs.StringVar(&s.KubeConfig, "kubeconfig", "", "path to the kubeconfig file")
	fs.StringVar(&s.KubeContext, "kube-context", s.KubeContext, "name of the kubeconfig context to use")
	fs.StringVar(&s.KubeToken, "kube-token", s.KubeToken, "bearer token used for authentication")
	fs.StringVar(&s.KubeAsUser, "kube-as-user", s.KubeAsUser, "username to impersonate for the operation")
	fs.StringArrayVar(&s.KubeAsGroups, "kube-as-group", s.KubeAsGroups, "group to impersonate for the operation, this flag can be repeated to specify multiple groups.")
	fs.StringVar(&s.KubeAPIServer, "kube-apiserver", s.KubeAPIServer, "the address and the port for the Kubernetes API server")
	fs.StringVar(&s.KubeCaFile, "kube-ca-file", s.KubeCaFile, "the certificate authority file for the Kubernetes API server connection")
	fs.BoolVar(&s.Debug, "debug", s.Debug, "enable verbose output")
	fs.StringVar(&s.RegistryConfig, "registry-config", s.RegistryConfig, "path to the registry config file")
	fs.StringVar(&s.RepositoryConfig, "repository-config", s.RepositoryConfig, "path to the file containing repository names and URLs")
	fs.StringVar(&s.RepositoryCache, "repository-cache", s.RepositoryCache, "path to the file containing cached repository indexes")
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}

func envIntOr(name string, def int) int {
	if name == "" {
		return def
	}
	envVal := envOr(name, strconv.Itoa(def))
	ret, err := strconv.Atoi(envVal)
	if err != nil {
		return def
	}
	return ret
}

func envCSV(name string) (ls []string) {
	trimmed := strings.Trim(os.Getenv(name), ", ")
	if trimmed != "" {
		ls = strings.Split(trimmed, ",")
	}
	return
}

// EnvironmentVariable represents information about helm env vars
type EnvironmentVariable struct {
	Name  string
	Value string
	Desc  string
}

// EnvVarsWithDesc return an array of helm environment variables
func (s *EnvSettings) EnvVarsWithDesc() []EnvironmentVariable {
	return []EnvironmentVariable{
		{Name: "HELM_BIN", Value: os.Args[0], Desc: "Path of the helm binary"},
		{Name: "HELM_CACHE_HOME", Value: helmpath.CachePath(""), Desc: "Path of helm's cache files"},
		{Name: "HELM_CONFIG_HOME", Value: helmpath.ConfigPath(""), Desc: "Path of helm's configuration files"},
		{Name: "HELM_DATA_HOME", Value: helmpath.DataPath(""), Desc: "Path of helm's data files"},
		{Name: "HELM_DEBUG", Value: fmt.Sprint(s.Debug), Desc: "Indicates if debug statements should be printed (true/false)"},
		{Name: "HELM_PLUGINS", Value: s.PluginsDirectory, Desc: "Path of helm's plugin files"},
		{Name: "HELM_REGISTRY_CONFIG", Value: s.RegistryConfig, Desc: "Path to the registry configuration file"},
		{Name: "HELM_REPOSITORY_CACHE", Value: s.RepositoryCache, Desc: "Path to the file containing cached repository indexes"},
		{Name: "HELM_REPOSITORY_CONFIG", Value: s.RepositoryConfig, Desc: "Path to the file containing repository names and URLs"},
		{Name: "HELM_NAMESPACE", Value: s.Namespace(), Desc: "Kubernetes namespace to which this helm command applies"},
		{Name: "HELM_MAX_HISTORY", Value: strconv.Itoa(s.MaxHistory), Desc: "Maximum number of revisions to configure for release history"},

		// broken, these are populated from helm flags and not kubeconfig.
		{Name: "HELM_KUBECONTEXT", Value: s.KubeContext, Desc: "Kubernetes context to which this helm command applies"},
		{Name: "HELM_KUBETOKEN", Value: s.KubeToken, Desc: "Path to the token file helm should use to contact the Kubernetes API server"},
		{Name: "HELM_KUBEASUSER", Value: s.KubeAsUser, Desc: "Username to impersonate for the operation"},
		{Name: "HELM_KUBEASGROUPS", Value: strings.Join(s.KubeAsGroups, ","), Desc: "Comma-separated list of groups to impersonate for the operation"},
		{Name: "HELM_KUBEAPISERVER", Value: s.KubeAPIServer, Desc: "<IP>:<port> helm should use to reach the Kubernetes API server"},
		{Name: "HELM_KUBECAFILE", Value: s.KubeCaFile, Desc: "Path to the certificate authority file helm should use to contact the Kubernetes API server"},
	}
}

// EnvVars returns a map of env var with their current values
func (s *EnvSettings) EnvVars() map[string]string {
	rawvars := s.EnvVarsWithDesc()
	envvars := make(map[string]string, len(rawvars))
	for _, v := range rawvars {
		envvars[v.Name] = v.Value
	}

	if s.KubeConfig != "" {
		envvars["KUBECONFIG"] = s.KubeConfig
	}
	return envvars
}

// Namespace gets the namespace from the configuration
func (s *EnvSettings) Namespace() string {
	if ns, _, err := s.config.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "default"
}

// RESTClientGetter gets the kubeconfig from EnvSettings
func (s *EnvSettings) RESTClientGetter() genericclioptions.RESTClientGetter {
	return s.config
}
