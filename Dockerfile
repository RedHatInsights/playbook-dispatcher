FROM registry.access.redhat.com/ubi8/go-toolset as builder

WORKDIR /go/src/app
COPY . .

USER 0

RUN go build -v -o app

FROM registry.access.redhat.com/ubi8-minimal

COPY --from=builder /go/src/app/app .

USER 1001

CMD ["/app"]
