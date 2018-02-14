package main

import (
	"fmt"
	"io/ioutil"
	"os"

	st "cloud.google.com/go/storage"
	"github.com/ghodss/yaml"
	"github.com/scottrigby/trigger-gcp-cloudbuild/cloudbuild"
	"github.com/scottrigby/trigger-gcp-cloudbuild/storage"
)

const (
	cloudbuildYAML  = "source/cloudbuild.yaml"
	storageFileName = "source.tgz"
)

func main() {
	y, err := ioutil.ReadFile(cloudbuildYAML)
	if err != nil {
		fmt.Print(err)
		return
	}

	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	projectID := os.Getenv("PROJECT_ID")
	bucketName := projectID + "_trigger-gcp-cloudbuild"
	bucketHandle, err := storage.GetBucketHandle(bucketName)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	sberr := storage.CreateBucket(bucketHandle, projectID, &st.BucketAttrs{})
	if sberr != nil {
		fmt.Printf("could not create storage bucket: %v\n", sberr)
		return
	}

	tgzerr := storage.MakeTarball("source", storageFileName)
	if tgzerr != nil {
		fmt.Printf("could not compress source tarball: %v\n", tgzerr)
		return
	}

	url, err := storage.WriteToStorage(bucketHandle, bucketName, storageFileName)
	if err != nil {
		fmt.Printf("could not copy source to storage: %v\n", err)
		return
	}
	fmt.Printf("source url: %s\n", url)

	// Add build substitution _PROJECT_ID for cloudbuild.yaml example.
	substitutions := map[string]string{"_PROJECT_ID": projectID}
	build, err := cloudbuild.GetBuild(j, bucketName, storageFileName, substitutions)
	if err != nil {
		fmt.Print(err)
		return
	}

	operation, err := cloudbuild.TriggerCloudBuild(projectID, build)
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Printf("cloudbuild operation: %s\n", operation)

	// Tidy up.
	os.Remove(storageFileName)
}
