# Trigger GCP Cloudbuild

## References

- ["Authenticating to Cloud Platform with Service Accounts"](https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform)

## Test setup

- Create GCP service account with "Cloud Container Builder" role (check the box "Furnish a new private key" and select key type "JSON".
- Note the `[/PATH/TO/SERVICE/ACCOUNT/KEYFILE].json`

## Local test steps

- Run the built test image:

    ```console
    $ docker run --rm -v [/PATH/TO/SERVICE/ACCOUNT/KEYFILE].json:/key.json --env PROJECT_ID=[GCP-PROJECT-ID] --env GOOGLE_APPLICATION_CREDENTIALS=/key.json docker.io/r6by/trigger-gcp-cloudbuild
    ```

## Local test steps during development

- Build vendor directory and packages:

    ```console
    $ dep ensure -v
    ```
- In your local session, set the `$GOOGLE_APPLICATION_CREDENTIALS` variable that `golang.org/x/oauth2/google` `FindDefaultCredentials()` looks for, and the `$PROJECT_ID` variable with the name of your GCP project ID:

    ```console
    $ export GOOGLE_APPLICATION_CREDENTIALS=[/PATH/TO/SERVICE/ACCOUNT/KEYFILE].json'
    $ export PROJECT_ID=[GCP-PROJECT-ID]
    ```
- Run the main package:

    ```console
    $ go run main.go
    ```

## GKE test steps

- Create generic secret for `$GOOGLE_APPLICATION_CREDENTIALS` ENV var:

    ```console
    $ kubectl create secret generic google-application-credentials --from-file=key.json=[/PATH/TO/SERVICE/ACCOUNT/KEYFILE].json
    ```
- Deploy the main test app:

    ```console
    $ kubectl create -f deploy.yaml
    ```
- **To-do:** Allow automatically setting their GCP project ID as an ENV var in the Deployment object. We may want to change this YAML file to a simple Helm chart just for this and other ENV vars needed as we go.
- Monitor the output with `kubectl logs` (or - shameless plug - try [kpoof](https://github.com/farmotive/kpoof) for fast, prompted k8s logs)
