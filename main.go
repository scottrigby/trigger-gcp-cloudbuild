package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"
	"golang.org/x/oauth2/google"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
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

	operation, err := TriggerCloudBuild(projectID, build)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Printf("cloudbuild do operation: %+v\n", operation)
}

func readFile(file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

// TriggerCloudBuild triggers a GCP Cloudbuild.
func TriggerCloudBuild(projectID string, build *cloudbuild.Build) (*cloudbuild.Operation, error) {
	ctx := context.Background()
	// Note that even though we define the GCP service account key via
	// GOOGLE_APPLICATION_CREDENTIALS ENV var, and roles are defined via service
	// accout IAM, we must define a scope here.
	client, err := google.DefaultClient(ctx, cloudbuild.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	cloudbuildService, err := cloudbuild.New(client)
	if err != nil {
		return nil, err
	}

	operation, err := cloudbuildService.Projects.Builds.Create(projectID, build).Do()
	// Google API verbose debugging info.
	if gerr, ok := err.(*googleapi.Error); ok {
		log.Println(gerr.Body)
	}
	if err != nil {
		return nil, err
	}

	return operation, nil
}
