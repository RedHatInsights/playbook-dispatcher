FROM registry.access.redhat.com/ubi8/openjdk-11 as builder
USER root

ENV MAVEN_DEP_DESTINATION=/deps \
    DEBEZIUM_VERSION=1.4.2.Final \
    DEBEZIUM_CHECKSUM=39fcc96ba02533779921ba87c513811a

WORKDIR /src
ADD event-streams /src
ADD schema/run.event.yaml /src

# Taken from https://github.com/debezium/docker-images/blob/master/connect-base/1.5/docker-maven-download.sh
RUN mkdir $MAVEN_DEP_DESTINATION && \
    /src/docker-maven-download.sh debezium postgres ${DEBEZIUM_VERSION} ${DEBEZIUM_CHECKSUM}

RUN ./mvnw -Dschema.event.run=/src/run.event.yaml clean package && \
    mv target/playbook-dispatcher-smt-*.jar target/playbook-dispatcher-smt.jar


FROM quay.io/strimzi/kafka:latest-kafka-2.7.0

ENV CONNECT_PLUGIN_PATH=/opt/kafka/plugins

USER root

RUN mkdir -p ${CONNECT_PLUGIN_PATH} /opt/kafka/custom-config/
COPY --from=builder /src/target/playbook-dispatcher-smt.jar $CONNECT_PLUGIN_PATH

# TODO: package separately https://issues.redhat.com/browse/RHCLOUD-12882
COPY --from=builder /src/target/playbook-dispatcher-smt.jar /opt/kafka/libs

COPY --from=builder /deps $CONNECT_PLUGIN_PATH
COPY event-streams/log4j.properties /opt/kafka/custom-config

USER 1001

