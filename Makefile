CLOUD_CONNECTOR_SCHEMA ?= https://raw.githubusercontent.com/RedHatInsights/cloud-connector/master/internal/controller/api/api.spec.json
RBAC_CONNECTOR_SCHEMA ?= https://cloud.redhat.com/api/rbac/v1/openapi.json
PSK ?= secret

init:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen
	go get github.com/atombender/go-jsonschema/...@main
	go install github.com/atombender/go-jsonschema/cmd/gojsonschema
	pip install json2yaml
	go get github.com/kulshekhar/fungen
	go install github.com/kulshekhar/fungen

clean:
	rm -f internal/api/controllers/public/spec.gen.go \
		internal/api/controllers/public/types.gen.go \
		internal/api/controllers/private/spec.gen.go \
		internal/api/controllers/private/types.gen.go
	rm -f internal/api/tests/public/client.gen.go internal/api/tests/private/client.gen.go
	rm -f internal/api/connectors/cloudConnector.gen.go cloud-connector.yaml
	rm -f internal/api/rbac/rbac.gen.go rbac.yaml

generate-api: \
 internal/api/controllers/public/spec.gen.go \
 internal/api/controllers/public/types.gen.go \
 internal/api/controllers/private/spec.gen.go \
 internal/api/controllers/private/types.gen.go

# public API
internal/api/controllers/public/spec.gen.go: schema/public.openapi.yaml
	~/go/bin/oapi-codegen -generate server,spec -package public -o internal/api/controllers/public/spec.gen.go schema/public.openapi.yaml

internal/api/controllers/public/types.gen.go: schema/public.openapi.yaml
	~/go/bin/oapi-codegen -generate types -package public -o internal/api/controllers/public/types.gen.go schema/public.openapi.yaml

# internal API
internal/api/controllers/private/spec.gen.go: schema/private.openapi.yaml
	~/go/bin/oapi-codegen -generate server,spec -package private -o internal/api/controllers/private/spec.gen.go -import-mapping=./public.openapi.yaml:playbook-dispatcher/internal/api/controllers/public schema/private.openapi.yaml

internal/api/controllers/private/types.gen.go: schema/private.openapi.yaml
	~/go/bin/oapi-codegen -generate types -package private -o internal/api/controllers/private/types.gen.go -import-mapping=./public.openapi.yaml:playbook-dispatcher/internal/api/controllers/public schema/private.openapi.yaml

generate-clients: internal/api/tests/public/client.gen.go internal/api/tests/private/client.gen.go

internal/api/tests/public/client.gen.go: schema/public.openapi.yaml
	~/go/bin/oapi-codegen -generate client,types -package public -o internal/api/tests/public/client.gen.go schema/public.openapi.yaml

internal/api/tests/private/client.gen.go: schema/private.openapi.yaml
	~/go/bin/oapi-codegen -generate client,types -package private -o internal/api/tests/private/client.gen.go -import-mapping=./public.openapi.yaml:playbook-dispatcher/internal/api/controllers/public schema/private.openapi.yaml

generate-messages: \
  ./internal/common/model/message/runner.types.gen.go \
  ./internal/common/model/message/rhcsat.types.gen.go

./internal/common/model/message/runner.types.gen.go: schema/playbookRunResponse.message.yaml
	~/go/bin/gojsonschema --yaml-extension yaml -p message schema/playbookRunResponse.message.yaml > ./internal/common/model/message/runner.types.gen.go

./internal/common/model/message/rhcsat.types.gen.go: schema/playbookSatRunResponse.message.yaml
	~/go/bin/gojsonschema --yaml-extension yaml -p message schema/playbookSatRunResponse.message.yaml > ./internal/common/model/message/rhcsat.types.gen.go

generate-cloud-connector: internal/api/connectors/cloudConnector.gen.go

internal/api/connectors/cloudConnector.gen.go: cloud-connector.yaml
	~/go/bin/oapi-codegen -generate client,types -package connectors -o internal/api/connectors/cloudConnector.gen.go cloud-connector.yaml
	rm cloud-connector.json

cloud-connector.yaml:
	curl -s ${CLOUD_CONNECTOR_SCHEMA} -o cloud-connector.json
	json2yaml cloud-connector.json cloud-connector.yaml

generate-rbac: internal/api/rbac/rbac.gen.go

internal/api/rbac/rbac.gen.go: rbac.yaml
	~/go/bin/oapi-codegen -generate client,types -package rbac -include-tags Access -o internal/api/rbac/rbac.gen.go rbac.yaml
	rm rbac.json

rbac.yaml:
	curl -s ${RBAC_CONNECTOR_SCHEMA} -o rbac.json
	json2yaml rbac.json rbac.yaml

generate-utils:
	go generate ./...

generate: generate-api generate-messages generate-cloud-connector generate-utils generate-clients generate-rbac

build:
	go build -o pd .

migrate-db:
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . migrate up

migrate-db-undo-all:
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run . migrate down-all

run: migrate-db
	ACG_CONFIG=$(shell pwd)/cdappconfig.json PSK_AUTH_TEST=xwKhCUzgJ8 go run . run

test: migrate-db
	SCHEMA_API_PRIVATE=$(shell pwd)/schema/private.openapi.yaml ACG_CONFIG=$(shell pwd)/cdappconfig.json PSK_AUTH_TEST=xwKhCUzgJ8 PSK_AUTH_TEST02=9yh9WuXWDj go test -p 1 -v ./...

sample_request:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload.json" http://localhost:8000/internal/dispatch

sample_request_v2:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload-v2.json" http://localhost:8000/internal/v2/dispatch

sample_request_sat:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload-sat.json" http://localhost:8000/internal/v2/dispatch

sample_recipient_status:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d '[{"recipient": "35720ecb-bc23-4b06-a8cd-f0c264edf2c1", "org_id": "5318290"}]' http://localhost:8000/internal/v2/recipients/status

sample_upload:
	curl -v -F "file=@examples/events-success.jsonl;type=application/vnd.redhat.playbook.v1+jsonl" -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample_upload.gz:
	curl -v -F "file=@examples/events-success.jsonl.gz;type=application/vnd.redhat.playbook.v1+gzip" -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample_upload.xz:
	curl -v -F "file=@examples/events-success.jsonl.xz;type=application/vnd.redhat.playbook.v1+xz" -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample_rhc_sat_upload:
	curl -v -F "file=@examples/rhcsat-success.jsonl;type=application/vnd.redhat.playbook-sat.v3+jsonl" -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlN5c3RlbSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

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
