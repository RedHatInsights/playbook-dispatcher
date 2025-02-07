FROM registry.access.redhat.com/ubi9/go-toolset as builder
USER 0

WORKDIR /go/src/app

COPY go.mod go.sum ./
COPY internal/ internal/
COPY cmd/ cmd/
COPY main.go main.go

RUN go mod download

RUN go build -v -o app main.go

FROM registry.access.redhat.com/ubi9-minimal

ARG BUILD_COMMIT=unknown

RUN microdnf update -y

COPY --from=builder /go/src/app/app .
COPY schema /schema
COPY migrations /migrations

ENV BUILD_COMMIT=${BUILD_COMMIT}

COPY licenses/LICENSE /licenses/LICENSE

USER 1001

ENTRYPOINT [ "/app" ]
CMD ["run"]
