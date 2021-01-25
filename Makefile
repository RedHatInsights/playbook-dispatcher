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
	./migrate.linux-amd64 -source file://./migrations -database postgresql://insights:insights@${DB_HOST}:5432/insights?sslmode=disable up

migrate-db-undo:
	./migrate.linux-amd64 -source file://./migrations -database postgresql://insights:insights@${DB_HOST}:5432/insights?sslmode=disable down 2

run: migrate-db
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . run

test: migrate-db
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go test -p 1 -v ./...
