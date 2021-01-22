FROM registry.access.redhat.com/ubi8/go-toolset as builder

WORKDIR /go/src/app
COPY . .

USER 0

RUN go build -v -o app cmd/pd/pd.go

FROM registry.access.redhat.com/ubi8-minimal

COPY --from=builder /go/src/app/app .
COPY schema /schema

ENV SCHEMA_MESSAGE_RESPONSE=/schema/playbookRunResponse.message.yaml

USER 1001

CMD ["/app"]
