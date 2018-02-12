package cloudbuild

import (
	"context"
	"encoding/json"
	"log"

	auth "golang.org/x/oauth2/google"
	cb "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
)

// GetBuild gets a Build resource.
// Note all substitutions for manually triggered cloudbuilds are user-defined.
// ref: https://cloud.google.com/container-builder/docs/configuring-builds/substitute-variable-values
// > User-defined substitutions must begin with an underscore (_) and
// 	 use only uppercase-letters and numbers (respecting the regular
//	 expression _[A-Z0-9_]+).
func GetBuild(j []byte, bucket string, object string, substitutions map[string]string) (*cb.Build, error) {
	build := &cb.Build{
		Substitutions: substitutions,
		Source: &cb.Source{
			StorageSource: &cb.StorageSource{
				Bucket: bucket,
				Object: object}}}
	err := json.Unmarshal(j, build)
	if err != nil {
		return nil, err
	}
	return build, nil
}

// TriggerCloudBuild triggers a GCP Cloudbuild.
func TriggerCloudBuild(projectID string, build *cb.Build) (*cb.Operation, error) {
	ctx := context.Background()
	// Note that even though we define the GCP service account key via
	// GOOGLE_APPLICATION_CREDENTIALS ENV var, and roles are defined via service
	// accout IAM, we must define a scope here.
	client, err := auth.DefaultClient(ctx, cb.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	cbService, err := cb.New(client)
	if err != nil {
		return nil, err
	}

	operation, err := cbService.Projects.Builds.Create(projectID, build).Do()
	// Google API verbose debugging info.
	if gerr, ok := err.(*googleapi.Error); ok {
		log.Println(gerr.Body)
	}
	if err != nil {
		return nil, err
	}

	return operation, nil
}
