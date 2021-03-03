package rbac

import (
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("RBAC", func() {
	DescribeTable("filter permissions",
		func(expected bool, body string) {
			doer := test.MockHttpClient(200, body)

			client := NewRbacClientWithHttpRequestDoer(config.Get(), &doer)
			permissions, err := client.GetPermissions(test.TestContext())
			Expect(err).ToNot(HaveOccurred())
			matchingPermissions := FilterPermissions(permissions, DispatcherPermission("run", "read"))
			Expect(len(matchingPermissions) > 0).To(Equal(expected))
		},

		Entry("no permissions", false, `{
			"meta": {
				"count": 0,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": []
		}`),

		Entry("single permission positive", true, `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:run:read"
				}
			]
		}`),

		Entry("single permission negative", true, `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:run:read"
				}
			]
		}`),

		Entry("single permission wildcard verb", true, `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:run:*"
				}
			]
		}`),

		Entry("single permission wildcard resource type", true, `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:*:read"
				}
			]
		}`),

		Entry("single permission wildcards", true, `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:*:*"
				}
			]
		}`),

		Entry("single permission wildcard verb negative", false, `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:salad:*"
				}
			]
		}`),

		Entry("single permission wildcard resource type", false, `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:*:destroy"
				}
			]
		}`),

		Entry("multiple permissions", true, `{
			"meta": {
				"count": 2,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:run:write"
				},
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:run:read"
				}
			]
		}`),

		Entry("permissions for different application", false, `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "patch:*:*"
				}
			]
		}`),
	)

	Describe("errors", func() {
		It("detects page overflow", func() {
			doer := test.MockHttpClient(200, `{
				"meta": {
					"count": 1002,
					"limit": 1000,
					"offset": 0
				},
				"links": {
					"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
					"next": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=1000",
					"previous": null,
					"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
				},
				"data": []
			}`)

			client := NewRbacClientWithHttpRequestDoer(config.Get(), &doer)
			_, err := client.GetPermissions(test.TestContext())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`RBAC page overflow`))
		})

		It("detects unexpected status code", func() {
			doer := test.MockHttpClient(500, `{}`)

			client := NewRbacClientWithHttpRequestDoer(config.Get(), &doer)
			_, err := client.GetPermissions(test.TestContext())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unexpected status code "500"`))
		})
	})

	DescribeTable("predicate values",
		func(body string, values ...interface{}) {
			doer := test.MockHttpClient(200, body)

			client := NewRbacClientWithHttpRequestDoer(config.Get(), &doer)
			permissions, err := client.GetPermissions(test.TestContext())
			Expect(err).ToNot(HaveOccurred())
			matchingPermissions := FilterPermissions(permissions, DispatcherPermission("run", "read"))
			services := GetPredicateValues(matchingPermissions, "service")
			Expect(services).To(ConsistOf(values...))
		},

		Entry("empty", `{
			"meta": {
				"count": 0,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": []
		}`),

		Entry("no predicates", `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [],
					"permission": "playbook-dispatcher:run:read"
				}
			]
		}`),

		Entry("no matching permission", `{
			"meta": {
				"count": 1,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [
						{
							"attributeFilter": {
								"key": "service",
								"value": "remediations",
								"operation": "equal"
							}
						}
					],
					"permission": "playbook-dispatcher:run:write"
				}
			]
		}`),

		Entry("service predicates", `{
			"meta": {
				"count": 3,
				"limit": 1000,
				"offset": 0
			},
			"links": {
				"first": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0",
				"next": null,
				"previous": null,
				"last": "/api/rbac/v1/access/?application=playbook-dispatcher&limit=1000&offset=0"
			},
			"data": [
				{
					"resourceDefinitions": [
						{
							"attributeFilter": {
								"key": "service",
								"value": "remediations",
								"operation": "equal"
							}
						}
					],
					"permission": "playbook-dispatcher:run:read"
				},
				{
					"resourceDefinitions": [
						{
							"attributeFilter": {
								"key": "service",
								"value": "config-manager",
								"operation": "equal"
							}
						}
					],
					"permission": "playbook-dispatcher:run:read"
				},
				{
					"resourceDefinitions": [
						{
							"attributeFilter": {
								"key": "service",
								"value": "test",
								"operation": "equal"
							}
						}
					],
					"permission": "playbook-dispatcher:run:destroy"
				}
			]
		}`, "remediations", "config-manager"),
	)
})
