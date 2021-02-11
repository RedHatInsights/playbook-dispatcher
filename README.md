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
{"event": "executor_on_start", "uuid": "4533e4d7-5034-4baf-b578-821305c96da4", "counter": -1, "stdout": "", "start_line": 0, "end_line": 0, "event_data": {"dispatcher_correlation_id": "c37278ac-f41c-424a-8461-7f41b4b87c8e"}}
{"event": "playbook_on_start", "uuid": "cb93301e-5ff8-4f75-ade6-57d0ec2fc662", "counter": 0, "stdout": "", "start_line": 0, "end_line": 0}
{"event": "playbook_on_stats", "uuid": "998a4bd2-2d6b-4c31-905c-2d5ad7a7f8ab", "counter": 1, "stdout": "", "start_line": 0, "end_line": 0}
```

The structure of each event is validated using a JSON Schema defined in [ansibleRunnerJobEvent.yaml](./schema/ansibleRunnerJobEvent.yaml).
Note that additional attributes (not defined by the schema) are allowed.

The expected content type of the uploaded file is `application/vnd.redhat.playbook.v1+jsonl` for a plain file or `application/vnd.redhat.playbook.v1+tgz` if the content is compressed (not implemented yet).

### Non-standard event types

Besides Ansible Runner event types (`playbook_*` and `runner_*`) the services recognizes two additional event types.

Firstly, `executor_on_start` event type is produced before Ansible Runner is invoked.

```json
{"event": "executor_on_start", "uuid": "4533e4d7-5034-4baf-b578-821305c96da4", "counter": -1, "stdout": "", "start_line": 0, "end_line": 0, "event_data": {
    "crc_dispatcher_correlation_id": "c37278ac-f41c-424a-8461-7f41b4b87c8e"
}}
```

The event carries the correlation id in the `event_data` dictionary.
This identifier is necessary to tie the uploaded file to a Playbook run.

Secondly, `executor_on_failed` event type may be produced should it not be possible to start a Playbook run for some reason.
This may happen e.g. if the Playbook signature is not valid, Ansible Runner binary is not available/executable, etc.

```json
{"event": "executor_on_failed", "uuid": "14f07467-0020-4ab7-a075-d7d8c96d6fdc", "counter": -1, "stdout": "", "start_line": 0, "end_line": 0, "event_data": {
    "crc_dispatcher_correlation_id": "c37278ac-f41c-424a-8461-7f41b4b87c8e",
    "crc_dispatcher_error_code": "SIGNATURE_INVALID",
    "crc_dispatcher_error_details": "Signature \"783701f7599830824fa73488f80eb79894f6f14203264b6a3ac3f0a14012c25f\" is not valid for Play \"run insights to obtain latest diagnosis info\"",
}}
```

Again, the correlation id is defined.
In addition, an error code and detailed information should be provided.

## Cloud Connector integration

Playbook Dispatcher uses [Cloud Connector](https://github.com/RedHatInsights/cloud-connector) to invoke Playbooks on connected hosts.
For each Playbook run request it sents the following message to Cloud Connector:

```json
{
    "account":"540155",
    "directive":"playbook",
    "recipient":"869fe355-4b69-43f6-82ff-d151dddee472", // id of the cloud connector client
    "metadata":{
        "crc_dispatcher_correlation_id":"e957564e-b823-4047-9ad7-0277dc61c88f", // see Non-standard event types for more details
        "response_interval":"600", // how often the recipient should send back responses
        "return_url":"https://cloud.redhat.com/api/v1/ingres/upload" // URL to post responses to
    },
    // playbook to execute
    "payload": "https://cloud.redhat.com/api/v1/remediations/1234/playbook?hosts=8f876606-5289-47f7-bb65-3966f0ba3ae1"
}
```

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
