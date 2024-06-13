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

func AppendGithubSourceToEtree(source *etree.Element, githubSource *devopsv1alpha3.GithubSource) {
	if githubSource == nil {
		klog.Warning("please provide GitHub source when the sourceType is GitHub")
		return
	}
	source.CreateAttr("class", "org.jenkinsci.plugins.github_branch_source.GitHubSCMSource")
	source.CreateAttr("plugin", "github-branch-source")
	source.CreateElement("id").SetText(githubSource.ScmId)
	source.CreateElement("credentialsId").SetText(githubSource.CredentialId)
	source.CreateElement("repoOwner").SetText(githubSource.Owner)
	source.CreateElement("repository").SetText(githubSource.Repo)
	if githubSource.ApiUri != "" {
		source.CreateElement("apiUri").SetText(githubSource.ApiUri)
	}
	traits := source.CreateElement("traits")
	if githubSource.DiscoverBranches != 0 {
		traits.CreateElement("org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(githubSource.DiscoverBranches))
	}
	if githubSource.DiscoverPRFromOrigin != 0 {
		traits.CreateElement("org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(githubSource.DiscoverPRFromOrigin))
	}
	if githubSource.DiscoverPRFromForks != nil {
		forkTrait := traits.CreateElement("org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait")
		forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(githubSource.DiscoverPRFromForks.Strategy))
		trustClass := "org.jenkinsci.plugins.github_branch_source.ForkPullRequestDiscoveryTrait$"
		if prTrust := GitHubPRDiscoverTrust(githubSource.DiscoverPRFromForks.Trust); prTrust.IsValid() {
			trustClass += prTrust.String()
		} else {
			klog.Warningf("invalid GitHub discover PR trust value: %d", prTrust.Value())
		}
		forkTrait.CreateElement("trust").CreateAttr("class", trustClass)
	}
	if githubSource.DiscoverTags {
		traits.CreateElement("org.jenkinsci.plugins.github__branch__source.TagDiscoveryTrait")
	}
	if githubSource.CloneOption != nil {
		cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
		cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
		cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(githubSource.CloneOption.Shallow))
		cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
		cloneExtension.CreateElement("honorRefspec").SetText(strconv.FormatBool(true))
		cloneExtension.CreateElement("reference")
		if githubSource.CloneOption.Timeout >= 0 {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(githubSource.CloneOption.Timeout))
		} else {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
		}

		if githubSource.CloneOption.Depth >= 0 {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(githubSource.CloneOption.Depth))
		} else {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
		}
	}
	if githubSource.RegexFilter != "" {
		regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
		regexTraits.CreateAttr("plugin", "scm-api")
		regexTraits.CreateElement("regex").SetText(githubSource.RegexFilter)
	}
	if !githubSource.AcceptJenkinsNotification {
		skipNotifications := traits.CreateElement("org.jenkinsci.plugins.github.notifications.NotificationsSkipTrait")
		skipNotifications.CreateAttr("plugin", "skip-notifications-trait")
	}
	return
}

func GetGithubSourcefromEtree(source *etree.Element) *devopsv1alpha3.GithubSource {
	var githubSource devopsv1alpha3.GithubSource
	if credential := source.SelectElement("credentialsId"); credential != nil {
		githubSource.CredentialId = credential.Text()
	}
	if repoOwner := source.SelectElement("repoOwner"); repoOwner != nil {
		githubSource.Owner = repoOwner.Text()
	}
	if repository := source.SelectElement("repository"); repository != nil {
		githubSource.Repo = repository.Text()
	}
	if apiUri := source.SelectElement("apiUri"); apiUri != nil {
		githubSource.ApiUri = apiUri.Text()
	}
	traits := source.SelectElement("traits")
	if branchDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(branchDiscoverTrait.SelectElement("strategyId").Text())
		githubSource.DiscoverBranches = strategyId
	}
	if tagDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.TagDiscoveryTrait"); tagDiscoverTrait != nil {
		githubSource.DiscoverTags = true
	}
	if originPRDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait"); originPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(originPRDiscoverTrait.SelectElement("strategyId").Text())
		githubSource.DiscoverPRFromOrigin = strategyId
	}
	if forkPRDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait"); forkPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(forkPRDiscoverTrait.SelectElement("strategyId").Text())
		if trustEle := forkPRDiscoverTrait.SelectElement("trust"); trustEle != nil {
			trustClass := trustEle.SelectAttr("class").Value
			trust := strings.Split(trustClass, "$")
			if prTrust := GitHubPRDiscoverTrust(1).ParseFromString(trust[1]); prTrust.IsValid() {
				githubSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: strategyId,
					Trust:    prTrust.Value(),
				}
			} else {
				klog.Warningf("invalid Gitlab discover PR trust value: %s", trust[1])
			}
		}

		githubSource.CloneOption = parseFromCloneTrait(traits.SelectElement("jenkins.plugins.git.traits.CloneOptionTrait"))

		if regexTrait := traits.SelectElement(
			"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
			if regex := regexTrait.SelectElement("regex"); regex != nil {
				githubSource.RegexFilter = regex.Text()
			}
		}

		if skipNotificationTrait := traits.SelectElement(
			"org.jenkinsci.plugins.github.notifications.NotificationsSkipTrait"); skipNotificationTrait == nil {
			githubSource.AcceptJenkinsNotification = true
		}
	}
	return &githubSource
}
