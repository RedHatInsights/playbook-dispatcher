FROM registry.access.redhat.com/ubi8/go-toolset as builder

WORKDIR /go/src/app

COPY go.mod go.sum .
COPY internal/ internal/
COPY cmd/ cmd/
COPY main.go main.go

RUN go mod download

USER 0

RUN go build -v -o app main.go

FROM registry.access.redhat.com/ubi8-minimal

ARG BUILD_COMMIT=unknown

COPY --from=builder /go/src/app/app .
COPY schema /schema
COPY migrations /migrations

ENV BUILD_COMMIT=${BUILD_COMMIT}

USER 1001

ENTRYPOINT [ "/app" ]
CMD ["run"]
