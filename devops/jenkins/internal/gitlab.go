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
	"strings"

	"github.com/beevik/etree"
	"k8s.io/klog/v2"

	devopsv1alpha3 "github.com/opswave/go-jenkins/devops/v1alpha3"
)

func AppendGitlabSourceToEtree(source *etree.Element, gitSource *devopsv1alpha3.GitlabSource) {
	if gitSource == nil {
		klog.Warning("please provide Gitlab source when the sourceType is Gitlab")
		return
	}
	source.CreateAttr("class", "io.jenkins.plugins.gitlabbranchsource.GitLabSCMSource")
	source.CreateAttr("plugin", "gitlab-branch-source")
	source.CreateElement("id").SetText(gitSource.ScmId)
	source.CreateElement("serverName").SetText(gitSource.ServerName)
	source.CreateElement("credentialsId").SetText(gitSource.CredentialId)
	source.CreateElement("projectOwner").SetText(gitSource.Owner)
	source.CreateElement("projectPath").SetText(gitSource.Repo)
	traits := source.CreateElement("traits")
	if gitSource.DiscoverBranches != 0 {
		traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.BranchDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverBranches))
	}
	if gitSource.DiscoverTags {
		traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.TagDiscoveryTrait")
	}
	if gitSource.DiscoverPRFromOrigin != 0 {
		traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.OriginMergeRequestDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverPRFromOrigin))
	}
	if gitSource.DiscoverPRFromForks != nil {
		forkTrait := traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait")
		forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverPRFromForks.Strategy))
		trustClass := "io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait$"

		if prTrust := PRDiscoverTrust(gitSource.DiscoverPRFromForks.Trust); prTrust.IsValid() {
			trustClass += prTrust.String()
		} else {
			klog.Warningf("invalid Gitlab discover PR trust value: %d", prTrust.Value())
		}
		forkTrait.CreateElement("trust").CreateAttr("class", trustClass)
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
	if !gitSource.AcceptJenkinsNotification {
		traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.GitLabSkipNotificationsTrait")
	}

	return
}

func GetGitlabSourceFromEtree(source *etree.Element) (gitSource *devopsv1alpha3.GitlabSource) {
	gitSource = &devopsv1alpha3.GitlabSource{}
	if credential := source.SelectElement("credentialsId"); credential != nil {
		gitSource.CredentialId = credential.Text()
	}
	if serverName := source.SelectElement("serverName"); serverName != nil {
		gitSource.ServerName = serverName.Text()
	}
	if repoOwner := source.SelectElement("projectOwner"); repoOwner != nil {
		gitSource.Owner = repoOwner.Text()
	}
	if repository := source.SelectElement("projectPath"); repository != nil {
		gitSource.Repo = repository.Text()
	}
	traits := source.SelectElement("traits")
	if branchDiscoverTrait := traits.SelectElement(
		"io.jenkins.plugins.gitlabbranchsource.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(branchDiscoverTrait.SelectElement("strategyId").Text())
		gitSource.DiscoverBranches = strategyId
	}
	if tagDiscoverTrait := traits.SelectElement(
		"io.jenkins.plugins.gitlabbranchsource.TagDiscoveryTrait"); tagDiscoverTrait != nil {
		gitSource.DiscoverTags = true
	}
	if originPRDiscoverTrait := traits.SelectElement(
		"io.jenkins.plugins.gitlabbranchsource.OriginMergeRequestDiscoveryTrait"); originPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(originPRDiscoverTrait.SelectElement("strategyId").Text())
		gitSource.DiscoverPRFromOrigin = strategyId
	}
	if forkPRDiscoverTrait := traits.SelectElement(
		"io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait"); forkPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(forkPRDiscoverTrait.SelectElement("strategyId").Text())
		if trustEle := forkPRDiscoverTrait.SelectElement("trust"); trustEle != nil {
			trustClass := trustEle.SelectAttr("class").Value
			trust := strings.Split(trustClass, "$")
			if prTrust := PRDiscoverTrust(1).ParseFromString(trust[1]); prTrust.IsValid() {
				gitSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: strategyId,
					Trust:    prTrust.Value(),
				}
			} else {
				klog.Warningf("invalid Gitlab discover PR trust value: %s", trust[1])
			}
		}

		gitSource.CloneOption = parseFromCloneTrait(traits.SelectElement("jenkins.plugins.git.traits.CloneOptionTrait"))

		if regexTrait := traits.SelectElement(
			"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
			if regex := regexTrait.SelectElement("regex"); regex != nil {
				gitSource.RegexFilter = regex.Text()
			}
		}

		if skipNotificationTrait := traits.SelectElement(
			"io.jenkins.plugins.gitlabbranchsource.GitLabSkipNotificationsTrait"); skipNotificationTrait == nil {
			gitSource.AcceptJenkinsNotification = true
		}
	}
	return
}
