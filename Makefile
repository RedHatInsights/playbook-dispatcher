init:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen
	go get github.com/atombender/go-jsonschema/...@master
	go install github.com/atombender/go-jsonschema/cmd/gojsonschema

generate-api:
	~/go/bin/oapi-codegen -generate server,spec -package controllers -o internal/api/controllers/spec.gen.go schema/api.spec.yaml
	~/go/bin/oapi-codegen -generate types -package controllers -o internal/api/controllers/types.gen.go schema/api.spec.yaml
	~/go/bin/oapi-codegen -generate client,types -package tests -o internal/api/tests/client.gen.go schema/api.spec.yaml

generate-messages:
	~/go/bin/gojsonschema --yaml-extension yaml -p message schema/playbookRunResponse.message.yaml > ./internal/common/model/message/types.gen.go

generate: generate-api generate-messages

build:
	go build -o pd .

migrate-db:
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . migrate up

migrate-db-undo-all:
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . migrate down-all

run: migrate-db
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . run

test: migrate-db
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go test -p 1 -v ./...

sample_request:
	curl -v -H "content-type: application/json" -d "@examples/payload.json" http://localhost:8000/internal/dispatch

sample_upload:
	curl -v -F "file=@examples/events-success.jsonl;type=application/vnd.redhat.playbook.v1+jsonl" -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample: sample_request sample_upload
