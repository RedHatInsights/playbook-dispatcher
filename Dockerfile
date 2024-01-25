FROM registry.access.redhat.com/ubi8/go-toolset:1.20.10-10 as builder

WORKDIR /go/src/app
COPY . .

USER 0

RUN go build -v -o app main.go

FROM registry.access.redhat.com/ubi8-minimal:latest

ARG BUILD_COMMIT=unknown

COPY --from=builder /go/src/app/app .
COPY schema /schema
COPY migrations /migrations

ENV BUILD_COMMIT=${BUILD_COMMIT}

USER 1001

ENTRYPOINT [ "/app" ]
CMD ["run"]
