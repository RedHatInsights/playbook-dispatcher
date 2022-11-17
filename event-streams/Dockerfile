FROM registry.redhat.io/ubi8/openjdk-11 as builder
USER root

ENV EXTRA_PLUGINS=/deps/plugins \
    EXTRA_LIBS=/deps/libs \
    DEBEZIUM_VERSION=1.4.2.Final \
    DEBEZIUM_CHECKSUM=39fcc96ba02533779921ba87c513811a \
    CONFIG_PROVIDERS_VERSION=0.1.0 \
    CONFIG_PROVIDERS_CHECKSUM=108e0bf4148a37676bed866ff45e1199

WORKDIR /src
ADD event-streams /src
ADD schema/run.event.yaml /src
ADD schema/run.host.event.yaml /src

RUN microdnf install gzip

# Taken from https://github.com/debezium/docker-images/blob/master/connect-base/1.5/docker-maven-download.sh
RUN mkdir -p $EXTRA_PLUGINS $EXTRA_LIBS && \
    MAVEN_DEP_DESTINATION=$EXTRA_PLUGINS /src/docker-maven-download.sh debezium postgres ${DEBEZIUM_VERSION} ${DEBEZIUM_CHECKSUM} && \
    MAVEN_DEP_DESTINATION=$EXTRA_LIBS /src/docker-maven-download.sh central com/redhat/insights/kafka config-providers ${CONFIG_PROVIDERS_VERSION} ${CONFIG_PROVIDERS_CHECKSUM}

RUN ./mvnw -Dschema.event.run=/src/run.event.yaml -Dschema.event.host.run=/src/run.host.event.yaml clean package && \
    mv target/playbook-dispatcher-smt-*.jar target/playbook-dispatcher-smt.jar


# https://access.redhat.com/documentation/en-us/red_hat_amq/2021.q3/html-single/deploying_and_upgrading_amq_streams_on_openshift/index#creating-new-image-from-base-str
FROM registry.redhat.io/amq7/amq-streams-kafka-31-rhel8:2.2.0

ARG BUILD_COMMIT=unknown

ENV CONNECT_PLUGIN_PATH=/opt/kafka/plugins \
    CONNECT_LIB_PATH=/opt/kafka/libs \
    BUILD_COMMIT=${BUILD_COMMIT}

USER root

RUN microdnf install jq

COPY event-streams/check-connectors.sh /
COPY examples/connector-local.json /

RUN mkdir -p ${CONNECT_PLUGIN_PATH} /opt/kafka/custom-config/

COPY --from=builder /src/target/playbook-dispatcher-smt.jar $CONNECT_PLUGIN_PATH
COPY --from=builder /deps/plugins $CONNECT_PLUGIN_PATH

COPY --from=builder /deps/libs $CONNECT_LIB_PATH

COPY event-streams/log4j.properties /opt/kafka/custom-config

USER 1001

