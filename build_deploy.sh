#!/bin/bash

set -exv

IMAGE="quay.io/cloudservices/playbook-dispatcher"
IMAGE_CONNECT="quay.io/cloudservices/playbook-dispatcher-connect"
IMAGE_TAG=$(git rev-parse --short=7 HEAD)
BUILD_COMMIT=$(git rev-parse HEAD)

if [[ -z "$QUAY_USER" || -z "$QUAY_TOKEN" ]]; then
    echo "QUAY_USER and QUAY_TOKEN must be set"
    exit 1
fi

if [[ -z "$RH_REGISTRY_USER" || -z "$RH_REGISTRY_TOKEN" ]]; then
    echo "RH_REGISTRY_USER and RH_REGISTRY_TOKEN  must be set"
    exit 1
fi

DOCKER_CONF="$PWD/.docker"
mkdir -p "$DOCKER_CONF"

docker --config="$DOCKER_CONF" login -u="$QUAY_USER" -p="$QUAY_TOKEN" quay.io
docker --config="$DOCKER_CONF" login -u="$RH_REGISTRY_USER" -p="$RH_REGISTRY_TOKEN" registry.redhat.io

docker --config="$DOCKER_CONF" build --build-arg BUILD_COMMIT=${BUILD_COMMIT} -t "${IMAGE}:${IMAGE_TAG}" .
docker --config="$DOCKER_CONF" tag "${IMAGE}:${IMAGE_TAG}" "${IMAGE}:latest"
docker --config="$DOCKER_CONF" push "${IMAGE}:${IMAGE_TAG}"
docker --config="$DOCKER_CONF" push "${IMAGE}:latest"

docker --config="$DOCKER_CONF" build --build-arg BUILD_COMMIT=${BUILD_COMMIT} -f event-streams/Dockerfile -t "${IMAGE_CONNECT}:${IMAGE_TAG}" .
docker --config="$DOCKER_CONF" tag "${IMAGE_CONNECT}:${IMAGE_TAG}" "${IMAGE_CONNECT}:latest"
docker --config="$DOCKER_CONF" push "${IMAGE_CONNECT}:${IMAGE_TAG}"
docker --config="$DOCKER_CONF" push "${IMAGE_CONNECT}:latest"
