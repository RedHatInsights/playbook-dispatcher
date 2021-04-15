#!/bin/bash
# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="playbook-dispatcher"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="playbook-dispatcher"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
TEST_SECRETS_COMPONENT_NAME="test-secret-psk"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/playbook-dispatcher"

IQE_PLUGINS="rhc"
IQE_MARKER_EXPRESSION="smoke"
IQE_FILTER_EXPRESSION="playbook_dispatcher"

# Install bonfire repository/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh -o bootstrap.sh
source bootstrap.sh  # checks out bonfire and changes to "cicd" dir...

# Build Playbook Dispatcher image based on the latest commit
source build.sh

# Deploy the new image to an ephemeral environment
source _common_deploy_logic.sh
result=$(bonfire config deploy \
    --ref-env insights-stage \
    --app $APP_NAME \
    --set-template-ref $COMPONENT_NAME=$GIT_COMMIT \
    --set-template-ref $TEST_SECRETS_COMPONENT_NAME=$GIT_COMMIT \
    --set-image-tag $IMAGE=$IMAGE_TAG \
    --get-dependencies)

if [ $? -eq 0 ]; then
    export NAMESPACE=$result
fi

# Deploy Ingress services to the ephemeral environment
bonfire config deploy --ref-env insights-stage --app ingress --get-dependencies --namespace $NAMESPACE

# Deploy an IQE pod and run the smoke tests
source smoke_test.sh
