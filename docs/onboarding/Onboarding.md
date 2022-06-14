## Onboarding guide

If you are an application trying to integrate with playbook dispatcher, we hope you find the following information helpful.

### Dispatching playbooks

Currently, dispatching of playbooks is only possible with the [Internal/Private API](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/private.openapi.yaml), and it will only be available to accept requests within the cluster.

#### Authenticating with the Internal API

In order to send requests to the Internal API, the client is required to provide a service specific pre-shared key under the `Authorization` header of the HTTP request.

For example:
```javascript
POST /internal/dispatch HTTP/1.1
Authorization: PSK <pre-shared key>
```

The service specific pre-kshared keys are configured via environment variables.
You can open a PR in this repository to add a dispatcher pre-shared key for your service.
The environment variable will need to be in the form of `PSK_AUTH_<service id>`, for example, `PSK_AUTH_REMEDIATIONS`. Here is a PR that added a PSK for the Tasks app [#217](https://github.com/RedHatInsights/playbook-dispatcher/pull/217).

#### Obtaining the necessary infomation

To dispatch a playbook, you will need to POST to the private api's `/internal/v2/dispatch` endpoint. The openapi schema for this endpoint can be found [here](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/private.openapi.yaml#L77).

The endpoint accepts the following parameters:
* `recipient`
  * An uuid used by [Cloud Connector](https://github.com/RedHatInsights/cloud-connector) to route the playbook signal. For `rhc-worker-playbook` or "directly connected hosts", `recipient` can be obtained from the [system profile](https://github.com/RedHatInsights/inventory-schemas/blob/8000191d960da05c4ebf7960f4af8f7cf68bf616/schemas/system_profile/v1.yaml#L197). For satellite hosts, this will be provided by the [Sources Api](https://github.com/RedHatInsights/sources-api-go). This field is required.
* `org_id`
  * The account/tenant this playbook is part of. This field is required.
* `principal`
  * Username of the user this playbook belongs to. This field is required.
* `url`
  * The url that is hosting the playbook. This url will be used to execute the playbook. This field is required.
* `name`
  * The name of the playbook. Human readable name of the playbook run. Used to present the given playbook run in external systems (Satellite). This field is required.
* `web_console_url`
  * The url where the user can find more information on the playbook run. Optional, but highly suggested.
* `labels`
  * Metadata for the playbook run. This will be helpful when filtering playbook runs.
* `timeout`
  * Number of seconds to wait before the playbook run is marked as failed due to timeout.
* `hosts`
  * An array of objects containing information on the hosts involved in the run. Please refer to [this section](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/private.openapi.yaml#L202) of the schema for more infomation. For satellite playbook requests, the `inventory_id` will need to be provided for each hosts.
* recipient_config
  * Recipient config for satellite. Required for satellite playbooks.
    * `sat_id` - Satellite instance id.
    * `sat_org_id` - Satellite organization id.


Some minimal `rhc-worker-playbook` and satellite hosts dispatch request bodies can be found [here](https://github.com/RedHatInsights/playbook-dispatcher#dispatching-of-playbooks).

### Helpful Tips

#### Dispatching multiple playbook runs

The Internal API's dispatch enpoint supports the dispatching of multiple playbooks in one request.
This [sample request body](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/examples/payload-multiple-run-v2.json) demonstrates how to dispatch 3 different playbooks in a single request - 3 satellite hosts in one instance and then 3 more satellite hosts in another along with 3 more directly connected hosts.

You can try it out locally using the following curl command
```sh
	curl -v -H "content-type: application/json" -H "Authorization: PSK xwKhCUzgJ8" -d "@examples/payload-multiple-run-v2.json" http://localhost:8000/internal/v2/dispatch
```

#### Using labels to group runs

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

### Fetching information

Playbook dispatcher provides an event interface on the `platform.playbook-dispatcher.runs` kafka topic which can be used to listen for state changes. Please refer to [this section of the docs](https://github.com/RedHatInsights/playbook-dispatcher#event-interface) to learn more about the event interface.

You can also use playbook dispatcher's public API to fetch playbook run or host information on demand.

* Public API endpoint to fetch playbook run information: [here](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/public.openapi.yaml#L17).
* Public API endpoint to fetch host information: [here](https://github.com/RedHatInsights/playbook-dispatcher/blob/master/schema/public.openapi.yaml#L44).
