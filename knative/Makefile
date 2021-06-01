image-build:
	docker build -t quay.io/jharting/playbook-dispatcher:knative ..

image-push:
	docker push quay.io/jharting/playbook-dispatcher:knative

get-runs_mk: export ROUTE := $$(kn route describe playbook-dispatcher-api -o jsonpath='{.status.url}')
get-runs_mk:
	curl -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlVzZXIiLCAiaW50ZXJuYWwiOiB7Im9yZ19pZCI6ICIwMDAwMDEifX19" '$(ROUTE)/api/playbook-dispatcher/v1/runs?limit=1&fields[data]=id,recipient,status,url,service' | jq .


get-runs:
	curl -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlVzZXIiLCAiaW50ZXJuYWwiOiB7Im9yZ19pZCI6ICIwMDAwMDEifX19" 'http://playbook-dispatcher-api-default.apps-crc.testing/api/playbook-dispatcher/v1/runs?limit=1&fields[data]=id,recipient,status,url,service' | jq .

ingress_port_forward:
	oc port-forward svc/ingress-service 8080:8000

sample_request:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@../examples/payload.json" http://playbook-dispatcher-api-default.apps-crc.testing/internal/dispatch
	
sample_request_mk: export ROUTE := $$(kn route describe playbook-dispatcher-api -o jsonpath='{.status.url}')
sample_request_mk:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@../examples/payload.json" $(ROUTE)/internal/dispatch

sample_upload:
	curl -v -F "file=@../examples/events-success.jsonl;type=application/vnd.redhat.playbook.v1+jsonl" -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlVzZXIiLCAiaW50ZXJuYWwiOiB7Im9yZ19pZCI6ICIwMDAwMDEifX19" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload

sample_upload_fail:
	curl -v -F "file=@../examples/events-failure.jsonl;type=application/vnd.redhat.playbook.v1+jsonl" -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlVzZXIiLCAiaW50ZXJuYWwiOiB7Im9yZ19pZCI6ICIwMDAwMDEifX19" -H "x-rh-request_id: 380b4a04-7eae-4dff-a0b8-6e1af9186df0" http://localhost:8080/api/ingress/v1/upload
