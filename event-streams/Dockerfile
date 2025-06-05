FROM registry.redhat.io/ubi9/openjdk-17 as builder
USER root

ENV EXTRA_PLUGINS=/deps/plugins \
  EXTRA_LIBS=/deps/libs \
  DEBEZIUM_VERSION=3.0.7.Final \
  DEBEZIUM_CHECKSUM=1a92e55fe78bc44a302e45da2f1233db \
  CONFIG_PROVIDERS_VERSION=0.1.0 \
  CONFIG_PROVIDERS_CHECKSUM=108e0bf4148a37676bed866ff45e1199

WORKDIR /src
ADD event-streams /src
ADD schema/run.event.yaml /src
ADD schema/run.host.event.yaml /src

RUN microdnf update -y && microdnf install -y jq && microdnf install -y gzip

# Taken from https://github.com/debezium/docker-images/blob/master/connect-base/1.9/docker-maven-download.sh
# Debezium PostgreSQL checksum from https://github.com/debezium/container-images/blob/main/connect/2.7/Dockerfile
RUN mkdir -p $EXTRA_PLUGINS $EXTRA_LIBS && \
  MAVEN_DEP_DESTINATION=$EXTRA_PLUGINS /src/docker-maven-download.sh debezium postgres ${DEBEZIUM_VERSION} ${DEBEZIUM_CHECKSUM} && \
  MAVEN_DEP_DESTINATION=$EXTRA_LIBS /src/docker-maven-download.sh central com/redhat/insights/kafka config-providers ${CONFIG_PROVIDERS_VERSION} ${CONFIG_PROVIDERS_CHECKSUM}

RUN ./mvnw -Dschema.event.run=/src/run.event.yaml -Dschema.event.run.host=/src/run.host.event.yaml clean package && \
  mv target/playbook-dispatcher-smt-*.jar target/playbook-dispatcher-smt.jar


# https://access.redhat.com/documentation/en-us/red_hat_amq/2021.q3/html-single/deploying_and_upgrading_amq_streams_on_openshift/index#creating-new-image-from-base-str
FROM registry.redhat.io/amq-streams/kafka-39-rhel9:2.9.0-11
