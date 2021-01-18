generate:
	~/go/bin/oapi-codegen -generate server,types,spec -package api -o api/generated.go schema/api.spec.yaml

run:
	ACG_CONFIG=$(shell pwd)/cdappconfig.json go run main.go
