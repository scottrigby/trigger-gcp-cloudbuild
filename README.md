![Hello emoji](https://raw.githubusercontent.com/googlei18n/noto-emoji/master/png/128/emoji_u1f44b.png) ![World emoji](https://raw.githubusercontent.com/googlei18n/noto-emoji/master/png/128/emoji_u1f30d.png) ![Cloud emoji](https://raw.githubusercontent.com/googlei18n/noto-emoji/master/png/128/emoji_u2601.png) ![Build emoji](https://raw.githubusercontent.com/googlei18n/noto-emoji/master/png/128/emoji_u1f3d7.png)

# Trigger GCP Cloudbuild

A "hello world" for triggering Google Cloudbuild in Golang.

## IAM setup

1. Create a GCP service account:

    ```console
    $ gcloud iam service-accounts create trigger-gcb
    ```

1. Add the "Cloud Container Builder Editor" and "Storage Object Admin" roles to the service account.

    ```console
    $ export SA_EMAIL=$(gcloud iam service-accounts list --filter="name:trigger-gcb" --format='value(email)')
    $ export PROJECT=$(gcloud info --format='value(config.project)')
    $ gcloud projects add-iam-policy-binding $PROJECT --member serviceAccount:$SA_EMAIL --role roles/storage.admin
    $ gcloud projects add-iam-policy-binding $PROJECT --member serviceAccount:$SA_EMAIL --role roles/cloudbuild.builds.editor
    ```

1. Create a JSON key for the service-account.

    ```console
    $ gcloud iam service-accounts keys create trigger-gcb.json --iam-account $SA_EMAIL
    ```

## GKE test steps

- Create generic secret for `$GOOGLE_APPLICATION_CREDENTIALS` ENV var:

    ```console
    $ kubectl create secret generic google-application-credentials --from-file=key.json=trigger-gcb.json
    ```
- Deploy the main test app:

    ```console
    $ export PROJECT=$(gcloud info --format='value(config.project)')
    $ helm install trigger-gcp-cloudbuild/ --set projectID=$PROJECT --name gcb
    ```
- Monitor the output with `kubectl logs` (or - shameless plug - try [klog](https://github.com/farmotive/klog) for fast, prompted k8s logs)
    - The `gcb-built` Job pod logs should output:
      > Built by GCP Cloudbuild
- Cleanup:

    ```console
    $ helm delete --purge gcb
    $ kubectl delete secret google-application-credentials
    ```

## IAM, Storage and Images cleanup

- Delete the service account:

    ```console
    $ gcloud iam service-accounts delete $SA_EMAIL
    ```
- Remove the storage source file, then bucket:

    ```console
    $ gsutil rm gs://${PROJECT}_trigger-gcp-cloudbuild/source.tgz
    $ gsutil rb gs://${PROJECT}_trigger-gcp-cloudbuild
    ```

- Remove the built images:

    ```console
    $ gcloud container images list-tags gcr.io/${PROJECT}/built-by-gcp-cloudbuild --format='get(digest)' | while read -r d; do command gcloud container images delete gcr.io/${PROJECT}/built-by-gcp-cloudbuild@"$d" --force-delete-tags --quiet; done
    ```

## Local test steps

- Remove any existing built images:

    ```console
    $ gcloud container images delete gcr.io/${PROJECT}/built-by-gcp-cloudbuild --quiet
    $ docker rmi gcr.io/${PROJECT}/built-by-gcp-cloudbuild
    ```

- Trigger cloudbuild locally with Docker:

    ```console
    $ docker run --rm -v trigger-gcb.json:/key.json --env PROJECT_ID=${PROJECT} --env GOOGLE_APPLICATION_CREDENTIALS=/key.json docker.io/r6by/trigger-gcp-cloudbuild
    ```
- Run the built test image:

    ```console
    $ docker run --rm gcr.io/${PROJECT}/built-by-gcp-cloudbuild
    ```
    Should output:
    > Built by GCP Cloudbuild

## Local development steps

- Build vendor directory and packages:

    ```console
    $ dep ensure -v
    ```
- In your local session, set the `$GOOGLE_APPLICATION_CREDENTIALS` variable that `golang.org/x/oauth2/google` `FindDefaultCredentials()` looks for, and the `$PROJECT_ID` variable with the name of your GCP project ID:

    ```console
    $ export GOOGLE_APPLICATION_CREDENTIALS=$(pwd)/trigger-gcb.json
    $ export PROJECT_ID=${PROJECT}
    ```
- Run the main package:

    ```console
    $ go run main.go
    ```
