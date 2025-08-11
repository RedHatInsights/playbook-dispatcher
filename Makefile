CLOUD_CONNECTOR_SCHEMA ?= https://raw.githubusercontent.com/RedHatInsights/cloud-connector/master/internal/controller/api/api.spec.json
RBAC_CONNECTOR_SCHEMA ?= https://raw.githubusercontent.com/RedHatInsights/insights-rbac/refs/heads/master/docs/source/specs/openapi.json
# Newer commits do not generate 3/11/25
INVENTORY_CONNECTOR_SCHEMA ?= https://raw.githubusercontent.com/RedHatInsights/insights-host-inventory/ffa3cab521f907e006f392d1698bf730346bed94/swagger/openapi.json
SOURCES_CONNECTOR_SCHEMA ?= https://console.redhat.com/api/sources/v3.1/openapi.json

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
LOCAL_BIN_PATH := ${PROJECT_PATH}/bin

export PATH := ${LOCAL_BIN_PATH}:${PATH}


PSK ?= secret

all: init generate build test run-lint

init:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/atombender/go-jsonschema@v0.17.0
	go install github.com/kulshekhar/fungen@latest

generate-api: internal/api/controllers/public/spec.gen.go \
	          internal/api/controllers/public/types.gen.go \
	          internal/api/controllers/private/spec.gen.go \
	          internal/api/controllers/private/types.gen.go

internal/api/controllers/public/spec.gen.go internal/api/controllers/public/types.gen.go: schema/public.openapi.yaml
	# public API
	${GOPATH}/bin/oapi-codegen -generate server,spec -package public -o internal/api/controllers/public/spec.gen.go schema/public.openapi.yaml
	${GOPATH}/bin/oapi-codegen -generate types -package public -o internal/api/controllers/public/types.gen.go schema/public.openapi.yaml

internal/api/controllers/private/spec.gen.go internal/api/controllers/private/types.gen.go: schema/public.openapi.yaml schema/private.openapi.yaml
	# private API
	${GOPATH}/bin/oapi-codegen -generate server,spec -package private -o internal/api/controllers/private/spec.gen.go -import-mapping=./public.openapi.yaml:playbook-dispatcher/internal/api/controllers/public schema/private.openapi.yaml
	${GOPATH}/bin/oapi-codegen -generate types -package private -o internal/api/controllers/private/types.gen.go -import-mapping=./public.openapi.yaml:playbook-dispatcher/internal/api/controllers/public schema/private.openapi.yaml

# use this when the oapi-codegen module version is updated and the golang validator fails
.PHONY: update-local-api
update-local-api: touch-openapi-files generate-api

.PHONY: touch-openapi-files
touch-openapi-files:
	# touch the openapi source config files
	touch schema/private.openapi.yaml
	touch schema/public.openapi.yaml

generate-clients: internal/api/tests/public/client.gen.go \
	              internal/api/tests/private/client.gen.go

internal/api/tests/public/client.gen.go: schema/public.openapi.yaml schema/private.openapi.yaml
	${GOPATH}/bin/oapi-codegen -generate client,types -package public -o internal/api/tests/public/client.gen.go schema/public.openapi.yaml

internal/api/tests/private/client.gen.go: schema/public.openapi.yaml schema/private.openapi.yaml
	${GOPATH}/bin/oapi-codegen -generate client,types -package private -o internal/api/tests/private/client.gen.go -import-mapping=./public.openapi.yaml:playbook-dispatcher/internal/api/controllers/public schema/private.openapi.yaml

generate-messages: internal/common/model/message/runner.types.gen.go \
	               internal/common/model/message/rhcsat.types.gen.go

internal/common/model/message/runner.types.gen.go: schema/playbookRunResponse.message.yaml
	${GOPATH}/bin/go-jsonschema --yaml-extension yaml -p message schema/playbookRunResponse.message.yaml > ./internal/common/model/message/runner.types.gen.go

internal/common/model/message/rhcsat.types.gen.go: schema/playbookSatRunResponse.message.yaml
	${GOPATH}/bin/go-jsonschema --yaml-extension yaml -p message schema/playbookSatRunResponse.message.yaml > ./internal/common/model/message/rhcsat.types.gen.go

generate-cloud-connector:
	curl -s ${CLOUD_CONNECTOR_SCHEMA} -o cloud-connector.json
	${GOPATH}/bin/oapi-codegen -generate client,types -package connectors -o internal/api/connectors/cloudConnector.gen.go cloud-connector.json
	rm cloud-connector.json

generate-rbac:
	curl -s ${RBAC_CONNECTOR_SCHEMA} -o rbac.json
	${GOPATH}/bin/oapi-codegen -generate client,types -package rbac -include-tags Access -o internal/api/rbac/rbac.gen.go rbac.json
	rm rbac.json

generate-inventory:
	curl -s ${INVENTORY_CONNECTOR_SCHEMA} -o inventory.json
	patch -p1 inventory.json oapi_codegen/inventory_xgo_name.patch
	${GOPATH}/bin/oapi-codegen -config oapi_codegen/oapi-codegen-inventory-cfg.yaml -generate client,types -package inventory -o internal/api/connectors/inventory/inventory.gen.go inventory.json
	rm inventory.json

generate-sources:
	curl -s ${SOURCES_CONNECTOR_SCHEMA} -o sources.json
	patch -p1 sources.json oapi_codegen/sources_xgo_name.patch
	${GOPATH}/bin/oapi-codegen -config oapi_codegen/oapi-codegen-sources-cfg.yaml -generate client,types -package sources -o internal/api/connectors/sources/sources.gen.go sources.json
	rm sources.json

generate-utils: internal/api/controllers/private/utils.gen.go \
	            internal/api/controllers/private/utils.v2.gen.go \
	            internal/api/controllers/private/cancel_utils.v2.gen.go

internal/api/controllers/private/utils.gen.go: internal/api/controllers/private/runsCreate.go
	go generate ./...

internal/api/controllers/private/utils.v2.gen.go: internal/api/controllers/private/runsCreateV2.go
	go generate ./...

internal/api/controllers/private/cancel_utils.v2.gen.go: internal/api/controllers/private/runsCancelV2.go
	go generate ./...

generate: generate-api generate-clients generate-messages generate-cloud-connector generate-rbac generate-inventory generate-sources generate-utils

build:
	go build -v -o app .

migrate-db:
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . migrate up

migrate-db-undo-all:
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . migrate down-all

run: migrate-db
	ACG_CONFIG=$(shell pwd)/cdappconfig.json PSK_AUTH_TEST=xwKhCUzgJ8 go run . run

run-sasl: migrate-db
	ACG_CONFIG=$(shell pwd)/cdappconfig-sasl.json PSK_AUTH_TEST=xwKhCUzgj8 go run . run

test: migrate-db
	SCHEMA_API_PRIVATE=$(shell pwd)/schema/private.openapi.yaml ACG_CONFIG=$(shell pwd)/cdappconfig.json PSK_AUTH_TEST=xwKhCUzgJ8 PSK_AUTH_TEST02=9yh9WuXWDj go test -p 1 -v ./...

sample_request:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload.json" http://localhost:8000/internal/dispatch

sample_request_v2:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload-v2.json" http://localhost:8000/internal/v2/dispatch

sample_request_multiple_v2:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload-multiple-run-v2.json" http://localhost:8000/internal/v2/dispatch

sample_request_sat:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload-sat.json" http://localhost:8000/internal/v2/dispatch

sample_request_sat_subman:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload-sat-subman.json" http://localhost:8000/internal/v2/dispatch

sample_recipient_status:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d '[{"recipient": "35720ecb-bc23-4b06-a8cd-f0c264edf2c1", "org_id": "5318290"}]' http://localhost:8000/internal/v2/recipients/status

sample_high_level_recipient_status:
	curl -v -H "content-type: application/json" -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -d '{"hosts": ["35720ecb-bc23-4b06-a8cd-f0c264edf2c1", "45720ecb-bc23-4b06-a8cd-f0c264edf2c1"], "org_id": "5318290"}' http://localhost:8000/internal/v2/connection_status

sample_upload:
	curl -v -F 'file=@examples/events-success.jsonl;headers=Content-type: application/vnd.redhat.playbook.v1+jsonl' -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample_upload.gz:
	curl -v -F 'file=@examples/events-success.jsonl.gz;headers=Content-type: application/vnd.redhat.playbook.v1+gzip' -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample_upload.xz:
	curl -v -F 'file=@examples/events-success.jsonl.xz;headers=Content-type: application/vnd.redhat.playbook.v1+xz' -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample_rhc_sat_upload:
	curl -v -F 'file=@examples/rhcsat-success.jsonl;headers=Content-type: application/vnd.redhat.playbook-sat.v3+jsonl' -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample_blocked_upload:
	curl -v -F 'file=@examples/events-success.jsonl;headers=Content-type: application/vnd.redhat.playbook.v1+jsonl' -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjEzMzcifX19" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample: sample_request sample_upload

connector_create:
	curl -i -H 'Content-Type:application/json' -X POST localhost:8083/connectors/ -d @examples/connector-local.json

connector_delete:
	curl -XDELETE localhost:8083/connectors/playbook-dispatcher-source

grafana:
	oc create configmap grafana-dashboard-insights-playbook-dispatcher --from-file=$(shell pwd)/dashboard/dashboard.json -o yaml --dry-run=client  > $(shell pwd)/dashboard/grafana-dashboard-insights-playbook-dispatcher.configmap.yaml
	echo -e '  labels:\n    grafana_dashboard: "true"\n  annotations:\n    grafana-folder: /grafana-dashboard-definitions/Insights' >> $(shell pwd)/dashboard/grafana-dashboard-insights-playbook-dispatcher.configmap.yaml

run_cleaner:
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . clean

ci-port-forward:
	oc port-forward svc/playbook-dispatcher-api 8000:8000 -n playbook-dispatcher-ci

ci-dispatch:
	curl -v -H "content-type: application/json" -H "Authorization: PSK ${PSK}" -d "@examples/payload.json" http://localhost:8000/internal/dispatch

golangci-lint:
ifeq (, $(shell which $(LOCAL_BIN_PATH)/golangci-lint 2> /dev/null))
	@{ \
	set -e ;\
	VERSION="v1.43.0" ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$${VERSION}/install.sh | sh -s -- -b ${LOCAL_BIN_PATH} $${VERSION} ;\
	}
endif

run-lint: golangci-lint
	$(LOCAL_BIN_PATH)/golangci-lint run

