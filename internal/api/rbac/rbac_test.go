package rbac

import (
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/utils/test"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("RBAC", func() {
	table.DescribeTable("filter permissions",
		func(expected bool, body string) {
			doer := test.MockHttpClient(200, body)

			client := NewRbacClientWithHttpRequestDoer(config.Get(), &doer)
			permissions, err := client.GetPermissions(test.TestContext())
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			matchingPermissions := FilterPermissions(permissions, DispatcherPermission("run", "read"))
			gomega.Expect(len(matchingPermissions) > 0).To(gomega.Equal(expected))
		},

		table.Entry("no permissions", false, `{
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

		table.Entry("single permission positive", true, `{
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

		table.Entry("single permission negative", true, `{
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

		table.Entry("single permission wildcard verb", true, `{
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

		table.Entry("single permission wildcard resource type", true, `{
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

		table.Entry("single permission wildcards", true, `{
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

		table.Entry("single permission wildcard verb negative", false, `{
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

		table.Entry("single permission wildcard resource type", false, `{
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

		table.Entry("multiple permissions", true, `{
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

		table.Entry("permissions for different application", false, `{
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

	ginkgo.Describe("errors", func() {
		ginkgo.It("detects page overflow", func() {
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
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring(`RBAC page overflow`))
		})

		ginkgo.It("detects unexpected status code", func() {
			doer := test.MockHttpClient(500, `{}`)

			client := NewRbacClientWithHttpRequestDoer(config.Get(), &doer)
			_, err := client.GetPermissions(test.TestContext())
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring(`unexpected status code "500"`))
		})
	})

	table.DescribeTable("predicate values",
		func(body string, values ...interface{}) {
			doer := test.MockHttpClient(200, body)

			client := NewRbacClientWithHttpRequestDoer(config.Get(), &doer)
			permissions, err := client.GetPermissions(test.TestContext())
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			matchingPermissions := FilterPermissions(permissions, DispatcherPermission("run", "read"))
			services := GetPredicateValues(matchingPermissions, "service")
			gomega.Expect(services).To(gomega.ConsistOf(values...))
		},

		table.Entry("empty", `{
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

		table.Entry("no predicates", `{
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

		table.Entry("no matching permission", `{
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

		table.Entry("service predicates", `{
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
