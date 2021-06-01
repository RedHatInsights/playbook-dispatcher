# Running playbook-dispatcher using knative

## Prerequisities (Openshift/CRC)

1. Install [CodeReady Containers](https://developers.redhat.com/products/codeready-containers/overview)
1. Install [OpenShift Serverless operator](https://docs.openshift.com/container-platform/4.7/serverless/admin_guide/installing-openshift-serverless.html#installing-openshift-serverless) by running

    ```
    oc apply -f knative.yaml
    ```
## Prerequisities (Minikube)

1. Install [Minikube](https://minikube.sigs.k8s.io/docs/start/).  Recommend the Docker/Podman driver, with 8G RAM
1. Start minikube tunnel

   ```
   minikube tunnel
   ```
1. Install [Knative](https://github.com/knative) by running

    ```
    knative-install-minikube.sh
    ```
    If you get an error about "resources not found", rerun the command after a brief wait.

1. Deploy the database

    ```
    oc project default
    oc apply -f db.yaml
    ```

1. Deploy Clowder

    ```
    oc apply -f https://github.com/RedHatInsights/clowder/releases/download/0.12.0/clowder-manifest-0.12.0.yaml --validate=false
    oc apply -f clowdenv.yml
    ```

1. Deploy ingress

```
    oc process -f ingress.yaml | oc apply -f -
    make ingress_port_forward
```

## Running playbook dispatcher

Run

1. `oc apply -f api.yaml` to deploy the API using Knative Serving
1. `oc apply -f validator.yaml` to deploy the validator service.
    This also deploys:

    - a KafkaSource for translating kafka messages from ingress to CloudEvents
    - a trigger for routing these messages to the validator service

1. `oc apply -f response-consumer.yaml` to deploy the response consumer.
    This also deploys a trigger for routing validator events to the validator service

Afterwards, run:

1. `make get-runs` (`make get-runs_mk` if using minikube) to get the list of playbook runs
1. `make sample_request` (`make sample_request_mk` if using minikube) to trigger a new playbook run (go to step 1 to verify it got created)
1. `make sample_upload` to upload playbook run artifacts

## Making code changes

After a change is made run `make image-build image-push` to build+promote to quay.
Then, update the image reference for the new image to get deployed.

TODO:
- split private/public API
- set up routing rules https://knative.dev/docs/serving/samples/knative-routing-go/
