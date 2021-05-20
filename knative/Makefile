
image-build:
	docker build -t quay.io/jharting/playbook-dispatcher:knative ..

image-push:
	docker push quay.io/jharting/playbook-dispatcher:knative

get-runs:
	curl -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJ0eXBlIjogIlVzZXIiLCAiaW50ZXJuYWwiOiB7Im9yZ19pZCI6ICIwMDAwMDEifX19" http://playbook-dispatcher-api-default.apps-crc.testing/api/playbook-dispatcher/v1/runs | jq .

sample_request:
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@../examples/payload.json" http://playbook-dispatcher-api-default.apps-crc.testing/internal/dispatch
