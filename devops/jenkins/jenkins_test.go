package jenkins

import (
	"fmt"
	"github.com/opswave/go-jenkins/devops"
	"net/url"
	"testing"
)

func TestCreateJenkins(t *testing.T) {
	jenkins := CreateJenkins(nil, "http://localhost:13000/", 100, "admin", "admin")
	job, err := jenkins.GetJob("zeus-be")
	apiURL, err := url.Parse("http://localhost:30180/")
	param := &devops.HttpParameters{
		Method: "GET",
		Url:    apiURL,
	}
	pipeline, err := jenkins.GetPipeline("test", "test", param)
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	fmt.Println(pipeline.ScmSource)
	fmt.Println(job)
}
