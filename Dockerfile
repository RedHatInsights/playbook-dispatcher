FROM registry.access.redhat.com/ubi9/go-toolset as builder
USER 0

WORKDIR /go/src/app

COPY go.mod go.sum ./
COPY inventory_xgo_name.patch sources_xgo_name.patch oapi-codegen-sources-cfg.yaml oapi-codegen-inventory-cfg.yaml ./
COPY internal/ internal/
COPY cmd/ cmd/
COPY schema schema/
COPY main.go main.go
COPY Makefile ./

ENV GOPATH="/opt/app-root/src"
ENV PATH="/opt/app-root/src/.go/bin:/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

RUN make init generate build

FROM registry.access.redhat.com/ubi9/ubi-minimal

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
