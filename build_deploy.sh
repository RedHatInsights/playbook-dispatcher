#!/bin/bash

set -exv

IMAGE="quay.io/cloudservices/playbook-dispatcher"
IMAGE_CONNECT="quay.io/cloudservices/playbook-dispatcher-connect"
IMAGE_TAG=$(git rev-parse --short=7 HEAD)

if [[ -z "$QUAY_USER" || -z "$QUAY_TOKEN" ]]; then
    echo "QUAY_USER and QUAY_TOKEN must be set"
    exit 1
fi

DOCKER_CONF="$PWD/.docker"
mkdir -p "$DOCKER_CONF"

docker --config="$DOCKER_CONF" login -u="$QUAY_USER" -p="$QUAY_TOKEN" quay.io

docker --config="$DOCKER_CONF" build -t "${IMAGE}:${IMAGE_TAG}" .
docker --config="$DOCKER_CONF" tag "${IMAGE}:${IMAGE_TAG}" "${IMAGE}:latest"
docker --config="$DOCKER_CONF" push "${IMAGE}:${IMAGE_TAG}"
docker --config="$DOCKER_CONF" push "${IMAGE}:latest"

docker --config="$DOCKER_CONF" build -f event-streams/Dockerfile -t "${IMAGE_CONNECT}:${IMAGE_TAG}" .
docker --config="$DOCKER_CONF" tag "${IMAGE_CONNECT}:${IMAGE_TAG}" "${IMAGE_CONNECT}:latest"
docker --config="$DOCKER_CONF" push "${IMAGE_CONNECT}:${IMAGE_TAG}"
docker --config="$DOCKER_CONF" push "${IMAGE_CONNECT}:latest"
