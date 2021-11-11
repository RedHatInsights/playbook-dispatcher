FROM registry.redhat.io/ubi8/go-toolset as builder

WORKDIR /go/src/app
COPY . .

USER 0

RUN go build -v -o app .

FROM registry.redhat.io/ubi8-minimal

ARG BUILD_COMMIT=unknown

COPY --from=builder /go/src/app/app .
COPY schema /schema
COPY migrations /migrations

ENV SCHEMA_MESSAGE_RESPONSE=/schema/playbookRunResponse.message.yaml \
    SCHEMA_RUNNER_EVENT=/schema/ansibleRunnerJobEvent.yaml \
    SCHEMA_API_PRIVATE=/schema/private.openapi.yaml \
    MIGRATIONS_DIR=/migrations \
    BUILD_COMMIT=${BUILD_COMMIT}

USER 1001

ENTRYPOINT [ "/app" ]
CMD ["run"]
