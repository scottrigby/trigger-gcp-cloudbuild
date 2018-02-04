package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"golang.org/x/oauth2/google"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
)

func main() {
	file := "source/cloudbuild.yaml"
	y, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(string(y))

	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Println(string(j))

	projectID := os.Getenv("PROJECT_ID")
	build := &cloudbuild.Build{}
	err2 := json.Unmarshal([]byte(j), build)
	if err2 != nil {
		fmt.Print(err2)
	}
	// Add build substitution _PROJECT_ID for cloudbuild.yaml.
	// ref: https://cloud.google.com/container-builder/docs/configuring-builds/substitute-variable-values
	// > User-defined substitutions must begin with an underscore (_) and use
	// 	 only uppercase-letters and numbers (respecting the regular expression
	//	 _[A-Z0-9_]+).
	build.Substitutions = map[string]string{"_PROJECT_ID": projectID}
	fmt.Printf("build: %+v\n", build)

	call, err3 := TriggerCloudBuild(projectID, build)
	if err3 != nil {
		fmt.Print(err3)
	}

	fmt.Printf("call: %+v\n", call)
}

func readFile(file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

// TriggerCloudBuild triggers a GCP Cloudbuild.
// @todo Check Create()'s *ProjectsBuildsCreateCall return value for success.
func TriggerCloudBuild(projectID string, build *cloudbuild.Build) (*cloudbuild.ProjectsBuildsCreateCall, error) {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx)
	if err != nil {
		return nil, err
	}

	cloudbuildService, err := cloudbuild.New(client)
	if err != nil {
		return nil, err
	}

	call := cloudbuildService.Projects.Builds.Create(projectID, build)

	return call, nil
}
