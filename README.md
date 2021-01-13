# Playbook Dispatcher

Playbook Dispatcher is a service for running Ansible Playbooks on hosts connected via [Cloud Connector](https://github.com/RedHatInsights/cloud-connector).

Playbook Dispatcher takes care of:

- dispatching the request to run a Playbook
- tracking the progress of a Playbook run

This project is WIP.
In the meantime see [API schema](./schemas/api.spec.yaml) for proposed API.

![Sequence diagram](./docs/sequence.svg)
