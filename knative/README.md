# Running playbook-dispatcher using knative

## Prerequisities

1. Install [CodeReady Containers](https://developers.redhat.com/products/codeready-containers/overview)
1. Install [OpenShift Serverless operator](https://docs.openshift.com/container-platform/4.7/serverless/admin_guide/installing-openshift-serverless.html#installing-openshift-serverless) by running

    ```
    oc apply -f knative.yaml
    ```

1. Deploy the database

    ```
    oc project default
    oc apply -f db.yaml
    ```

## Deploying API

Run `oc apply -f api.yaml` to deploy the API using Knative Serving

Use

```
curl -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlVzZXIiLCAiaW50ZXJuYWwiOiB7Im9yZ19pZCI6ICIwMDAwMDEifX19" http://playbook-dispatcher-api-default.apps-crc.testing/api/playbook-dispatcher/v1/runs
```
to verify it is working as expected.

TODO:
- split private/public API
- set up routing rules https://knative.dev/docs/serving/samples/knative-routing-go/

## Making code changes

After a change is made run `make image-build image-push` to build+promote to quay.
Then, update the image reference for the new image to get deployed.
