package jenkins

import (
	"fmt"
	"github.com/opswave/go-jenkins/devops"
	"net/url"
	"testing"
)

func TestCreateJenkins(t *testing.T) {
	jenkins := CreateJenkins(nil, "http://114.55.35.38:13000/", 100, "admin", "Jimoretech0226")
	job, err := jenkins.GetJob("zeus-be")
	apiURL, err := url.Parse("http://114.55.172.45:30180/")
	param := &devops.HttpParameters{
		Method: "GET",
		Url:    apiURL,
	}
	pipeline, err := jenkins.GetPipeline("jimore-1lkh2b", "zeus-be", param)
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	fmt.Print(pipeline)
	fmt.Print(job)
}
