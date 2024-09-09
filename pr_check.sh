#!/bin/bash
# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="playbook-dispatcher"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="playbook-dispatcher"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
CONNECT_COMPONENT_NAME="playbook-dispatcher-connect"
IMAGE="quay.io/cloudservices/playbook-dispatcher"
IQE_CJI_TIMEOUT="30m"
IQE_ENV_VARS="DYNACONF_USER_PROVIDER__rbac_enabled=false"
REF_ENV="insights-stage"

# Install bonfire repository/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh

# Build Playbook Dispatcher image based on the latest commit
source $CICD_ROOT/build.sh

# Execute unit tests
#source $APP_ROOT/unit_test.sh

# Build and Deploy Playbook Dispatcher Connect image
IMAGE="quay.io/cloudservices/playbook-dispatcher-connect"
DOCKERFILE="event-streams/Dockerfile"

IMAGE_DISPATCHER="quay.io/cloudservices/playbook-dispatcher"
IMAGE_CONNECT="quay.io/cloudservices/playbook-dispatcher-connect"

source $CICD_ROOT/build.sh

# IMAGE is set to the Connect image, setting dispatcher image as an extra arg
# hardcode connect to use a ref that works in ephemeral
EXTRA_DEPLOY_ARGS="--set-image-tag ${IMAGE_DISPATCHER}=${IMAGE_TAG} --set-template-ref ${CONNECT_COMPONENT_NAME}=${GIT_COMMIT}"

# Deploy to an ephemeral environment
source $CICD_ROOT/deploy_ephemeral_env.sh

# Re-deploy Playbook Dispatcher to an ephemeral environment, this time enabling the communication with Cloud Connector
# The connect image template is overridden to make use of the connect.yaml file from before managed kafka was put in place
bonfire deploy playbook-dispatcher cloud-connector \
    --source=appsre \
    --ref-env ${REF_ENV} \
    --set-template-ref ${COMPONENT_NAME}=${GIT_COMMIT} \
    --set-image-tag ${IMAGE_DISPATCHER}=${IMAGE_TAG} \
    --set-image-tag ${IMAGE_CONNECT}=${IMAGE_TAG} \
    --set-template-ref ${CONNECT_COMPONENT_NAME}=${GIT_COMMIT} \
    --namespace ${NAMESPACE} \
    --timeout ${DEPLOY_TIMEOUT} \
    --set-parameter playbook-dispatcher/CLOUD_CONNECTOR_IMPL=impl

# Run Playbook Dispatcher isolated tests
IQE_PLUGINS="playbook-dispatcher"
IQE_MARKER_EXPRESSION="smoke"
source $CICD_ROOT/cji_smoke_test.sh

# Run RHC Contract integration tests
COMPONENT_NAME="cloud-connector"
IQE_PLUGINS="rhc-contract"
IQE_IMAGE_TAG="rhc-contract"
source $CICD_ROOT/cji_smoke_test.sh
source $CICD_ROOT/post_test_results.sh
