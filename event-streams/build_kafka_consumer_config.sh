#!/usr/bin/env bash

JAAS_FILE=$1

cat <<HERE > $JAAS_FILE
sasl.mechanism=SCRAM-SHA-512
security.protocol=SASL_SSL
sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required \
  username="$KAFKA_USERNAME" \
  password="$KAFKA_SECRET";
HERE

