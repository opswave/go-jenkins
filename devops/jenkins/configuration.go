/*
Copyright 2022 The KubeSphere Authors.

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

package jenkins

import (
	"errors"
	"fmt"
	"net/http"
)

// According to: https://github.com/jenkinsci/configuration-as-code-plugin/blob/master/docs/features/configurationReload.md
const (
	reloadEndpoint         = "/configuration-as-code/reload/"
	checkNewSourceEndpoint = "/configuration-as-code/checkNewSource"
	replaceEndpoint        = "/configuration-as-code/replace"
)

// ReloadConfiguration reloads the Jenkins Configuration as Code YAML file
func (j *Jenkins) ReloadConfiguration() (err error) {
	var response *http.Response
	if response, err = j.Requester.Post(reloadEndpoint, nil, nil, nil); err != nil {
		err = fmt.Errorf("failed to send the POST request to reload Jenkins CasC, error: %v", err)
	} else if response.StatusCode != http.StatusFound {
		err = errors.New("failed to reload Jenkins CasC")
	}
	return
}

// CheckNewSource checks source of Configuration as Code from new location.
func (j *Jenkins) CheckNewSource(source string) (err error) {
	var response *http.Response
	if response, err = j.Requester.PostForm(checkNewSourceEndpoint, nil, nil, map[string]string{
		"newSource:": source,
	}); err == nil && response.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to check the new CasC source: %s, status code is: %d", source, response.StatusCode)
	}
	return
}

// ApplyNewSource applies a new config file
func (j *Jenkins) ApplyNewSource(source string) (err error) {
	if err = j.CheckNewSource(source); err != nil {
		err = fmt.Errorf("failed to check the new source: %s, error: %v", source, err)
		return
	}
	var response *http.Response
	if response, err = j.Requester.PostForm(replaceEndpoint, nil, nil, map[string]string{
		"json":        fmt.Sprintf(`{"newSource": "%s"}`, source),
		"_.newSource": source,
	}); err == nil && response.StatusCode != http.StatusFound {
		// Jenkins does not have a standard API. This is a form submit, so the expected code is not 200
		err = fmt.Errorf("failed to replace the new CasC source: %s, expected status code is: %d, got: %d",
			source, http.StatusFound, response.StatusCode)
	}
	return
}
