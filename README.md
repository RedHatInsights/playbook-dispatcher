# Playbook Dispatcher

Playbook Dispatcher is a service for running Ansible Playbooks on hosts connected via [Cloud Connector](https://github.com/RedHatInsights/cloud-connector).

Playbook Dispatcher takes care of:

- dispatching the request to run a Playbook
- tracking the progress of a Playbook run

This project is WIP.
In the meantime see [API schema](./schemas/api.spec.yaml) for proposed API.

![Sequence diagram](./docs/sequence.svg)

## Development

### Prerequisities

- Golang >= 1.14

### Running the service

Use `make run` to start the playbook-dispatcher

The application can be accessed at <http://localhost:8000/api/playbook-dispatcher/v1/runs>

```sh
curl -v -H "x-rh-identity: eyJpZGVudGl0eSI6eyJpbnRlcm5hbCI6eyJvcmdfaWQiOiI1MzE4MjkwIn0sImFjY291bnRfbnVtYmVyIjoiOTAxNTc4IiwidXNlciI6e30sInR5cGUiOiJVc2VyIn19Cg==" http://localhost:8000/api/playbook-dispatcher/v1/runs
```

### Running tests

`make test`
