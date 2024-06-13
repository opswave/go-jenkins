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

package internal

import (
	"strconv"

	"github.com/beevik/etree"
	"k8s.io/klog/v2"

	devopsv1alpha3 "github.com/opswave/go-jenkins/devops/v1alpha3"
)

func AppendGitSourceToEtree(source *etree.Element, gitSource *devopsv1alpha3.GitSource) {
	if gitSource == nil {
		klog.Warning("please provide Git source when the sourceType is Git")
		return
	}
	source.CreateAttr("class", "jenkins.plugins.git.GitSCMSource")
	source.CreateAttr("plugin", "git")
	source.CreateElement("id").SetText(gitSource.ScmId)
	source.CreateElement("remote").SetText(gitSource.Url)
	if gitSource.CredentialId != "" {
		source.CreateElement("credentialsId").SetText(gitSource.CredentialId)
	}
	traits := source.CreateElement("traits")
	if gitSource.DiscoverBranches {
		traits.CreateElement("jenkins.plugins.git.traits.BranchDiscoveryTrait")
	}
	if gitSource.DiscoverTags {
		traits.CreateElement("jenkins.plugins.git.traits.TagDiscoveryTrait")
	}
	if gitSource.CloneOption != nil {
		cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
		cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
		cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(gitSource.CloneOption.Shallow))
		cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
		cloneExtension.CreateElement("honorRefspec").SetText(strconv.FormatBool(true))
		cloneExtension.CreateElement("reference")
		if gitSource.CloneOption.Timeout >= 0 {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(gitSource.CloneOption.Timeout))
		} else {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
		}

		if gitSource.CloneOption.Depth >= 0 {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(gitSource.CloneOption.Depth))
		} else {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
		}
	}

	if gitSource.RegexFilter != "" {
		regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
		regexTraits.CreateAttr("plugin", "scm-api")
		regexTraits.CreateElement("regex").SetText(gitSource.RegexFilter)
	}
	return
}

func GetGitSourcefromEtree(source *etree.Element) *devopsv1alpha3.GitSource {
	var gitSource devopsv1alpha3.GitSource
	if credential := source.SelectElement("credentialsId"); credential != nil {
		gitSource.CredentialId = credential.Text()
	}
	if remote := source.SelectElement("remote"); remote != nil {
		gitSource.Url = remote.Text()
	}

	traits := source.SelectElement("traits")
	if branchDiscoverTrait := traits.SelectElement(
		"jenkins.plugins.git.traits.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
		gitSource.DiscoverBranches = true
	}
	if tagDiscoverTrait := traits.SelectElement(
		"jenkins.plugins.git.traits.TagDiscoveryTrait"); tagDiscoverTrait != nil {
		gitSource.DiscoverTags = true
	}

	gitSource.CloneOption = parseFromCloneTrait(traits.SelectElement("jenkins.plugins.git.traits.CloneOptionTrait"))

	if regexTrait := traits.SelectElement(
		"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
		if regex := regexTrait.SelectElement("regex"); regex != nil {
			gitSource.RegexFilter = regex.Text()
		}
	}
	return &gitSource
}
