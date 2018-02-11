package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/ghodss/yaml"
	"golang.org/x/oauth2/google"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
)

const (
	cloudbuildYAML  = "source/cloudbuild.yaml"
	storageFileName = "source.tgz"
)

func main() {
	y, err := ioutil.ReadFile(cloudbuildYAML)
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
	storageBucketName := projectID + "_trigger-gcp-cloudbuild"
	storageBucket, err := getStorageBucket(storageBucketName)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	sberr := createStorageBucket(storageBucket, projectID, &storage.BucketAttrs{})
	if sberr != nil {
		fmt.Printf("could not create storage bucket: %v\n", sberr)
		return
	}
	url, err := copySourceToStorage(storageBucket, storageBucketName)
	if err != nil {
		fmt.Printf("could not copy source to storage: %v\n", err)
		return
	}
	fmt.Printf("source url: %s\n", url)

	build := &cloudbuild.Build{
		// Add build substitution _PROJECT_ID for cloudbuild.yaml.
		// ref: https://cloud.google.com/container-builder/docs/configuring-builds/substitute-variable-values
		// > User-defined substitutions must begin with an underscore (_) and
		// 	 use only uppercase-letters and numbers (respecting the regular
		//	 expression _[A-Z0-9_]+).
		Substitutions: map[string]string{"_PROJECT_ID": projectID},
		Source: &cloudbuild.Source{
			StorageSource: &cloudbuild.StorageSource{
				Bucket: storageBucketName,
				Object: "source.tgz"}}}
	err2 := json.Unmarshal([]byte(j), build)
	if err2 != nil {
		fmt.Print(err2)
	}
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

func createStorageBucket(bucket *storage.BucketHandle, name string, attrs *storage.BucketAttrs) error {
	ctx := context.Background()
	err := bucket.Create(ctx, name, attrs)

	// If the bucket exists, return without error, but give a friendly message.
	// > Error 409: You already own this bucket. Please select another name., conflict
	if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusConflict {
		fmt.Println("storage bucket already exists")
		return nil
	}

	// If any other error return it.
	if err != nil {
		return err
	}
	return nil
}

func getStorageBucket(name string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(name), nil
}

// Thanks to GoogleCloudPlatform/golang-samples.
// ref: https://github.com/GoogleCloudPlatform/golang-samples/blob/master/getting-started/bookshelf/app/app.go#L218
func copySourceToStorage(storageBucket *storage.BucketHandle, storageBucketName string) (url string, err error) {
	ctx := context.Background()
	w := storageBucket.Object(storageFileName).NewWriter(ctx)
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = "application/gzip"
	w.CacheControl = "public, max-age=86400"

	in, err := os.Open(storageFileName)
	if err != nil {
		return
	}
	defer in.Close()

	if _, err := io.Copy(w, in); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	const publicURL = "https://storage.googleapis.com/%s/%s"
	return fmt.Sprintf(publicURL, storageBucketName, storageFileName), nil
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
