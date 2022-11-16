pf:
	oc port-forward svc/playbook-dispatcher-connect 8083:8083

connector_create:
	curl -i -H 'Content-Type:application/json' -X POST localhost:8083/connectors/ -d @connector.json

connector_delete:
	curl -XDELETE localhost:8083/connectors/playbook-dispatcher-source

generate_java_classes:
	./mvnw generate-sources
