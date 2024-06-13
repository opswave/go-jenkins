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

package jclient

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/jenkins-zh/jenkins-client/pkg/job"
	"github.com/opswave/go-jenkins/devops"
	"github.com/opswave/go-jenkins/devops/v1alpha3"
	devopsv1alpha3 "github.com/opswave/go-jenkins/devops/v1alpha3"
	"k8s.io/klog/v2"
)

// CreateProjectPipeline creates the pipeline
func (j *JenkinsClient) CreateProjectPipeline(projectID string, pipeline *v1alpha3.Pipeline) (string, error) {
	jclient := job.Client{
		JenkinsCore: j.Core,
	}
	projectPipelineName := fmt.Sprintf("%s %s", projectID, pipeline.Name)
	job, _ := jclient.GetJob(projectPipelineName)
	if job != nil {
		err := fmt.Errorf("job name [%s] has been used", job.Name)
		return "", restful.NewError(http.StatusConflict, err.Error())
	}
	switch pipeline.Spec.Type {
	case devopsv1alpha3.NoScmPipelineType:
		createPayload, err := getCreatePayload(pipeline.Spec.Pipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		err = jclient.CreateJobInFolder(*createPayload, projectID)
		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}
		return pipeline.Name, nil
	case devopsv1alpha3.MultiBranchPipelineType:
		createPayload, err := getCreateMultiBranchPipelinePayload(pipeline.Spec.MultiBranchPipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		err = jclient.CreateJobInFolder(*createPayload, projectID)
		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}
		return pipeline.Name, nil
	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
}

// DeleteProjectPipeline deletes pipeline
func (j *JenkinsClient) DeleteProjectPipeline(projectID string, pipelineID string) (string, error) {
	jclient := job.Client{
		JenkinsCore: j.Core,
	}
	projectPipelineName := fmt.Sprintf("%s %s", projectID, pipelineID)
	err := jclient.Delete(projectPipelineName)
	if err != nil {
		return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
	}
	return pipelineID, nil
}

// UpdateProjectPipeline updates pipeline
func (j *JenkinsClient) UpdateProjectPipeline(projectID string, pipeline *devopsv1alpha3.Pipeline) (string, error) {
	return j.jenkins.UpdateProjectPipeline(projectID, pipeline)
}

// GetProjectPipelineConfig returns the pipeline config
func (j *JenkinsClient) GetProjectPipelineConfig(projectID, pipelineID string) (*devopsv1alpha3.Pipeline, error) {
	// TODO: get a pipeline config
	return j.jenkins.GetProjectPipelineConfig(projectID, pipelineID)
}

func getCreatePayload(pipeline *devopsv1alpha3.NoScmPipeline) (jobPayload *job.CreateJobPayload, err error) {
	// NoScmPipeline do not have copy mode to create a pipeline
	jobPayload = &job.CreateJobPayload{
		Mode: "org.jenkinsci.plugins.workflow.job.WorkflowJob",
		Name: pipeline.Name,
	}
	return
}

func getCreateMultiBranchPipelinePayload(pipeline *devopsv1alpha3.MultiBranchPipeline) (jobPayload *job.CreateJobPayload, err error) {
	// NoScmPipeline do not have copy mode to create a pipeline
	jobPayload = &job.CreateJobPayload{
		Mode: "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject",
		Name: pipeline.Name,
	}
	return
}
