/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package rules

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/lint/support"
	"k8s.io/helm/pkg/timeconv"
	tversion "k8s.io/helm/pkg/version"
)

// Templates lints the templates in the Linter.
func Templates(linter *support.Linter) {
	path := "templates/"
	templatesPath := filepath.Join(linter.ChartDir, path)

	templatesDirExist := linter.RunLinterRule(support.WarningSev, path, validateTemplatesDir(templatesPath))

	// Templates directory is optional for now
	if !templatesDirExist {
		return
	}

	// Load chart and parse templates, based on tiller/release_server
	chart, err := chartutil.Load(linter.ChartDir)

	chartLoaded := linter.RunLinterRule(support.ErrorSev, path, err)

	if !chartLoaded {
		return
	}

	options := chartutil.ReleaseOptions{Name: "testRelease", Time: timeconv.Now(), Namespace: "testNamespace"}
	caps := &chartutil.Capabilities{
		APIVersions: chartutil.DefaultVersionSet,
		KubeVersion: &version.Info{
			Major:     "1",
			Minor:     "6",
			GoVersion: runtime.Version(),
			Compiler:  runtime.Compiler,
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
		TillerVersion: tversion.GetVersionProto(),
	}
	valuesToRender, err := chartutil.ToRenderValuesCaps(chart, chart.Values, options, caps)
	if err != nil {
		// FIXME: This seems to generate a duplicate, but I can't find where the first
		// error is coming from.
		//linter.RunLinterRule(support.ErrorSev, err)
		return
	}
	renderedContentMap, err := engine.New().Render(chart, valuesToRender)

	renderOk := linter.RunLinterRule(support.ErrorSev, path, err)

	if !renderOk {
		return
	}

	serverAvailable := true

	/* Iterate over all the templates to check:
	- It is a .yaml file
	- All the values in the template file is defined
	- {{}} include | quote
	- Generated content is a valid Yaml file
	- Metadata.Namespace is not set
	*/

	for _, template := range chart.Templates {
		fileName, _ := template.Name, template.Data
		path = fileName

		linter.RunLinterRule(support.ErrorSev, path, validateAllowedExtension(fileName))

		// We only apply the following lint rules to yaml files
		if filepath.Ext(fileName) != ".yaml" {
			continue
		}

		// NOTE: disabled for now, Refs https://github.com/kubernetes/helm/issues/1463
		// Check that all the templates have a matching value
		//linter.RunLinterRule(support.WarningSev, path, validateNoMissingValues(templatesPath, valuesToRender, preExecutedTemplate))

		// NOTE: disabled for now, Refs https://github.com/kubernetes/helm/issues/1037
		// linter.RunLinterRule(support.WarningSev, path, validateQuotes(string(preExecutedTemplate)))

		renderedContent := renderedContentMap[filepath.Join(chart.GetMetadata().Name, fileName)]
		var yamlStruct K8sYamlStruct
		// Even though K8sYamlStruct only defines Metadata namespace, an error in any other
		// key will be raised as well
		err := yaml.Unmarshal([]byte(renderedContent), &yamlStruct)

		validYaml := linter.RunLinterRule(support.ErrorSev, path, validateYamlContent(err))

		if !validYaml {
			continue
		}

		if serverAvailable {
			// access kubernetes URL from the kubectl client
			kubeConfig := kube.GetConfig("")
			clientConfig, _ := kubeConfig.ClientConfig()
			u, _ := url.Parse(clientConfig.Host)

			// if kubernetes server is unavailable print a warning
			// and don't try again this run.
			timeout := time.Duration(5 * time.Second)
			_, err = net.DialTimeout("tcp" , u.Host , timeout)
			if err != nil {
				e := fmt.Errorf("%s, skipping schema validation\n", err)
				linter.RunLinterRule(support.WarningSev, path, e)
				serverAvailable = false
				continue
			}

			kubeClient := kube.New(kubeConfig)
			f := kubeClient.Factory

			// get the schema validator
			schema, err := f.Validator(true, kubeClient.SchemaCacheDir)
			validSchemaAccess := linter.RunLinterRule(support.ErrorSev, path, validateSchemaAccess(err))

			if !validSchemaAccess {
				continue
			}

			// convert to YAML to JSON, validated above so should be ok
			j, _ := yaml.YAMLToJSON([]byte(renderedContent))
			err = schema.ValidateBytes(j)

			validSchema := linter.RunLinterRule(support.ErrorSev, path, validateSchema(err))

			if !validSchema {
				continue
			}
		}
	}
}

// Validation functions
func validateTemplatesDir(templatesPath string) error {
	if fi, err := os.Stat(templatesPath); err != nil {
		return errors.New("directory not found")
	} else if err == nil && !fi.IsDir() {
		return errors.New("not a directory")
	}
	return nil
}

func validateAllowedExtension(fileName string) error {
	ext := filepath.Ext(fileName)
	validExtensions := []string{".yaml", ".tpl", ".txt"}

	for _, b := range validExtensions {
		if b == ext {
			return nil
		}
	}

	return fmt.Errorf("file extension '%s' not valid. Valid extensions are .yaml, .tpl, or .txt", ext)
}

func validateYamlContent(err error) error {
	if err != nil {
		return fmt.Errorf("unable to parse YAML\n\t%s", err)
	}
	return nil
}

func validateSchemaAccess(err error) error {
	if err != nil {
		return fmt.Errorf("can not access schema\n\t%s", err)
	}
	return nil
}

func validateSchema(err error) error {
	if err != nil {
		return fmt.Errorf("schema validation failure\n\t%s", err)
	}
	return nil
}

// K8sYamlStruct stubs a Kubernetes YAML file.
// Need to access for now to Namespace only
type K8sYamlStruct struct {
	Metadata struct {
		Namespace string
	}
}
