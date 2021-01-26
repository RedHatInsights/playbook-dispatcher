# Playbook Dispatcher

Playbook Dispatcher is a service for running Ansible Playbooks on hosts connected via [Cloud Connector](https://github.com/RedHatInsights/cloud-connector).

Playbook Dispatcher takes care of:

- dispatching the request to run a Playbook
- tracking the progress of a Playbook run

This project is WIP.
In the meantime see [API schema](./schemas/api.spec.yaml) for proposed API.

Playbook Dispatcher consists of 3 parts:

- API, which can be used to dispatch a Playbook or to query it's state
- validator, which validates archives uploaded via ingress service
- response consumer, which processes validated archives and updates the internal state accordingly

![Sequence diagram](./docs/sequence.svg)

## Expected input format

The service expects the uploaded file to contain Ansible Runner [job events](https://ansible-runner.readthedocs.io/en/stable/intro.html#runner-artifact-job-events-host-and-playbook-events).
Job events should be stored in the [newline-delimited JSON](https://jsonlines.org/) format.
Each line in the file matches one job event.

```jsonl
{"event": "playbook_on_start", "uuid": "cb93301e-5ff8-4f75-ade6-57d0ec2fc662", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0}
{"event": "playbook_on_stats", "uuid": "998a4bd2-2d6b-4c31-905c-2d5ad7a7f8ab", "counter": 1, "stdout": "", "start_line": 0, "end_line": 0}
```

The structure of each event is validated using a JSON Schema defined in [ansibleRunnerJobEvent.yaml](./schema/ansibleRunnerJobEvent.yaml).
Note that additional attributes (not defined by the schema) are allowed.

The expected content type of the uploaded file is `application/vnd.redhat.playbook.v1+jsonl` for a plain file or `application/vnd.redhat.playbook.v1+tgz` if the content is compressed (not implemented yet).

## Development

### Prerequisities

- Golang >= 1.14

### Running the service

Run `docker-compose up --build` to start the service and its dependencies

The API can be accessed at <http://localhost:8000/api/playbook-dispatcher/v1/runs>

```sh
curl -v -H "x-rh-identity: eyJpZGVudGl0eSI6eyJpbnRlcm5hbCI6eyJvcmdfaWQiOiI1MzE4MjkwIn0sImFjY291bnRfbnVtYmVyIjoiOTAxNTc4IiwidXNlciI6e30sInR5cGUiOiJVc2VyIn19Cg==" http://localhost:8000/api/playbook-dispatcher/v1/runs
```

Useful commands:

- `make sample_request` can be used to dispatch a new sample Playbook
- `make sample_upload` can be used to upload a sample archive via Ingress

### Running tests

`make test`
