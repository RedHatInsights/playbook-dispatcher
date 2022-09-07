# Onboarding guide

Playbook Dispatcher is a service that simplifies running Ansible Playbooks on connected hosts.
This guide explains how an application be integrated with it in order to benefit from:

* common API for dispatching playbooks
* automated tracking of playbook state
* common model for representing playbook results
* built-in authorization support

## Dispatching playbooks

Currently, dispatching of playbooks is only possible with the [Internal/Private API](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/private.openapi.yaml), and it will only be available to accept requests within the cluster.

### Authenticating with the Internal API

In order to send requests to the Internal API, the client is required to provide a service specific pre-shared key under the `Authorization` header of the HTTP request.

For example:
```javascript
POST /internal/dispatch HTTP/1.1
Authorization: PSK <pre-shared key>
```

The service specific pre-shared keys are configured via environment variables.
If you are interested in adding a dispatcher pre-shared key for your service, you can contact the pipeline team at `#team-consoledot-pipeline` channel on the CoreOs slack workspace.
A pipeline workstream memeber will then follow the [instructions outlined here](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/docs/console.redhat.com/app-sops/playbook-dispatcher/onboarding-new-application.md) to set up the pre-shared key for you.

### Obtaining the necessary infomation

To dispatch a playbook, you will need to POST to the private api's `/internal/v2/dispatch` endpoint.
The openapi schema for this endpoint can be found [here](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/private.openapi.yaml#L77).

The endpoint accepts the following parameters:
* `recipient`
  * An uuid used by [Cloud Connector](https://github.com/RedHatInsights/cloud-connector) to route the playbook signal. This field is required.
    * For `rhc-worker-playbook` or "directly connected hosts", `recipient` can be obtained from the [system profile](https://github.com/RedHatInsights/inventory-schemas/blob/8000191d960da05c4ebf7960f4af8f7cf68bf616/schemas/system_profile/v1.yaml#L197).
    * For satellite hosts, this will be provided by the [Sources Api](https://github.com/RedHatInsights/sources-api-go), where it is know as the `rhc_id`.
    The `rhc_id`, or `recipient`, can be obtained using the `/sources/{id}/rhc_connections` [endpoint](https://github.com/RedHatInsights/sources-api-go/blob/9a7c9288be53f84717a6337063a481dcf533f615/public/openapi-3-v3.1.json#L1333).
    For example:
      ```sh
        curl -v -H "x-rh-identity: <identity-header>" localhost:3000/api/sources/v3.1/sources/<source-id>/rhc_connections
      ```
      The `id` in the endpoint is the sources id and can be obtained using the upper-level `/sources` [endpoint](https://github.com/RedHatInsights/sources-api-go/blob/9a7c9288be53f84717a6337063a481dcf533f615/public/openapi-3-v3.1.json#L1498) by doing the following:
        ```sh
          curl -v -H "x-rh-identity: <identity-header>" localhost:3000/api/sources/v3.1/sources?filter[source_ref]=<satellite_instance_id>
        ```
      The satellite instance id is discussed below.
* `org_id`
  * The account/tenant this playbook is part of.
    This field is required.
* `principal`
  * Username of the user this playbook belongs to.
    This field is required.
* `url`
  * The url that is hosting the playbook.
    This url will be used to execute the playbook.
    This field is required.
    Also, a playbook that needs to be run by `rhc-worker-playbook`, must and only point to the `localhost` information under the `hosts` information.
* `name`
  * The name of the playbook.
    Human readable name of the playbook run.
    Used to present the given playbook run in external systems (Satellite).
    This field is required.
* `web_console_url`
  * The url where the user can find more information on the playbook run.
    Optional, but highly suggested.
* `labels`
  * Metadata for the playbook run.
    This will be helpful when filtering playbook runs.
* `timeout`
  * Number of seconds to wait before the playbook run is marked as failed due to timeout.
* `hosts`
  * An array of objects containing information on the hosts involved in the run.
    Please refer to [this section](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/private.openapi.yaml#L202) of the schema for more infomation.
    For satellite playbook requests, the `inventory_id` will need to be provided for each hosts.
* recipient_config
  * Recipient config for satellite.
    Required for satellite playbooks.
    * `sat_id` - Satellite instance id.
      Can be obtained from [host records provided by Inventory](https://github.com/RedHatInsights/insights-host-inventory/blob/9d2c837ee37a6fe50b628880ac5d823319749a82/swagger/api.spec.yaml#L901).
    * `sat_org_id` - Satellite organization id.
      This is present within [Inventory's host Facts](https://github.com/RedHatInsights/insights-host-inventory/blob/4e09b9154c364d2553c259cfeef2b99772aef06d/swagger/api.spec.yaml#L848).


Please note that, you can dispatch playbooks to multiple satellite hosts in one request (described below), however, for directly connected hosts there should be a separate request for each host.
Some minimal `rhc-worker-playbook` and satellite hosts dispatch request bodies can be found [here](https://github.com/RedHatInsights/playbook-dispatcher#dispatching-of-playbooks).

## Fetching playbook run information

Playbook dispatcher provides an event interface on the `platform.playbook-dispatcher.runs` kafka topic which can be used to listen for state changes of playbooks.
All the playbook information provided by the event interface is described in [this JSON schema](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/run.event.yaml).
Currently, the event interface only provides information on the playbook level.
Host level information, such as console logs, are not available through this interface.
Please refer to [this section of the docs](https://github.com/RedHatInsights/playbook-dispatcher#event-interface) to learn more about the event interface.

You can also use playbook dispatcher's public API to fetch playbook run or host information on demand.
Please refer to the following schema definitions for the details:

* Public API endpoint to fetch playbook run information: [here](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/public.openapi.yaml#L17).
* Public API endpoint to fetch host information: [here](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/public.openapi.yaml#L44).

## Helpful Tips

### Dispatching multiple satellite playbook runs

The Internal API's dispatch endpoint supports the dispatching of multiple playbooks in one request.
This [sample request body](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/examples/payload-multiple-run-v2.json) demonstrates how to dispatch 2 different playbooks in a single request - 3 satellite hosts in one instance and then 3 more satellite hosts in another.

You can try it out locally using the following curl command
```sh
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload-multiple-run-v2.json" http://localhost:8000/internal/v2/dispatch
```

### Using labels to group runs

You can use the `labels` object in the request body of your dispatch request metioned earlier to group together multiple runs.

For example:
```javascript
labels: {
    "playbook-type": "satellite",
    "service": "remediations"
}
```

Then, when fetching playbook run information through dispatcher's Public API, you can use the following to filter the runs based on the provided labels.
```sh
/api/playbook-dispatcher/v1/runs?filter[labels][playbook-type]=sat-playbook&filter[labels][service]=remediations
```
