![Hello emoji](https://raw.githubusercontent.com/googlei18n/noto-emoji/master/png/128/emoji_u1f44b.png) ![World emoji](https://raw.githubusercontent.com/googlei18n/noto-emoji/master/png/128/emoji_u1f30d.png) ![Cloud emoji](https://raw.githubusercontent.com/googlei18n/noto-emoji/master/png/128/emoji_u2601.png) ![Build emoji](https://raw.githubusercontent.com/googlei18n/noto-emoji/master/png/128/emoji_u1f3d7.png)

# Trigger GCP Cloudbuild

A hello world for Google Cloudbuild.

## Goal

Running the [trigger image](https://hub.docker.com/r/r6by/trigger-gcp-cloudbuild/) asks Google cloudbuild to create and store a second image defined by the included `source` directory. Running the built image should output:
> Built by GCP Cloudbuild

## Test setup

- Create GCP service account with "Cloud Container Builder Editor" and "Storage Object Admin" roles (check the box "Furnish a new private key" and select key type "JSON".
- Note the `[/PATH/TO/SERVICE/ACCOUNT/KEYFILE].json`

## GKE test steps

- Create generic secret for `$GOOGLE_APPLICATION_CREDENTIALS` ENV var:

    ```console
    $ kubectl create secret generic google-application-credentials --from-file=key.json=[/PATH/TO/SERVICE/ACCOUNT/KEYFILE].json
    ```
- Deploy the main test app:

    ```console
    $ helm install trigger-gcp-cloudbuild/ --set projectID=[GCP-PROJECT-ID] --name trigger-gcp-cloudbuild
    ```
- Monitor the output with `kubectl logs` (or - shameless plug - try [kpoof](https://github.com/farmotive/kpoof) for fast, prompted k8s logs)
- Cleanup:

    ```console
    $ helm delete --purge trigger-gcp-cloudbuild
    $ kubectl delete secret google-application-credentials
    ```

## Local test steps

- Run the built test image:

    ```console
    $ docker run --rm -v [/PATH/TO/SERVICE/ACCOUNT/KEYFILE].json:/key.json --env PROJECT_ID=[GCP-PROJECT-ID] --env GOOGLE_APPLICATION_CREDENTIALS=/key.json docker.io/r6by/trigger-gcp-cloudbuild
    ```

## Local development steps

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
