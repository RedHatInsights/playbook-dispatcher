#!/bin/bash
# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="playbook-dispatcher"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="playbook-dispatcher"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/playbook-dispatcher"

IQE_PLUGINS="playbook-dispatcher"
IQE_MARKER_EXPRESSION="smoke"
IQE_FILTER_EXPRESSION=""
IQE_CJI_TIMEOUT="30m"


# Install bonfire repository/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh

# Build Playbook Dispatcher image based on the latest commit
source $CICD_ROOT/build.sh

# Execute unit tests
#source $APP_ROOT/unit_test.sh

# Deploy the new image to an ephemeral environment
source $CICD_ROOT/deploy_ephemeral_env.sh

# Deploy an IQE pod and run the smoke tests
source $CICD_ROOT/cji_smoke_test.sh
