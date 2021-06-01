#!/bin/bash

set -eo pipefail

KNATIVE_VERSION=${KNATIVE_VERSION:-0.23.0}

set -u
echo "Installing knative operator....."
n=0
until [ $n -ge 2 ]; do
  kubectl apply -f https://github.com/knative/operator/releases/download/v$KNATIVE_VERSION/operator.yaml && break
  n=$[$n+1]
  sleep 5
done
kubectl wait --for=condition=Established --all crd

echo "Installing  eventing and serving components...."
n=0
until [ $n -ge 2 ]; do
  kubectl apply -f knative-minikube.yaml && break
  n=$[$n+1]
  sleep 5
done

kubectl wait pod --timeout=-1s --for=condition=Ready -l '!job-name' -n knative-serving
kubectl wait pod --timeout=-1s --for=condition=Ready -l '!job-name' -n knative-eventing

echo "Configuring serving ingress....."
EXTERNAL_IP=$(kubectl -n knative-serving get service kourier -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
while [  -z $EXTERNAL_IP ]; do
  sleep 5
  EXTERNAL_IP=$(kubectl -n knative-serving get service kourier -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
done

echo "The EXTERNAL_IP is $EXTERNAL_IP"
KNATIVE_DOMAIN="$EXTERNAL_IP.nip.io"
kubectl patch configmap -n knative-serving config-domain -p "{\"data\": {\"$KNATIVE_DOMAIN\": \"\"}}"

echo "Installing  kafka and source components...."
kubectl apply -f https://github.com/knative-sandbox/eventing-kafka-broker/releases/download/v$KNATIVE_VERSION/eventing-kafka-controller.yaml
kubectl apply -f https://github.com/knative-sandbox/eventing-kafka-broker/releases/download/v$KNATIVE_VERSION/eventing-kafka-broker.yaml
kubectl apply -f https://github.com/knative-sandbox/eventing-kafka/releases/download/v$KNATIVE_VERSION/source.yaml



