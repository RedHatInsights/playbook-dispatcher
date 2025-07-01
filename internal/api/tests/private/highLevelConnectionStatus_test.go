package private

import (
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/tests/common"
	"strconv"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getConnectionStatus(payload ApiInternalHighlevelConnectionStatusJSONRequestBody) (*ApiInternalHighlevelConnectionStatusResponse, error) {
	orgId := "12345"
	// Build a test client that passes an identity header because the high
	// level interface requires the identity header
	identityPassingClient := &Client{
		Server:         common.TestServer,
		Client:         common.TestClient,
		RequestEditors: []RequestEditorFn{common.TestRequestEditor},
	}
	ctx := common.ContextWithIdentity(orgId)
	resp, err := identityPassingClient.ApiInternalHighlevelConnectionStatus(ctx, payload)
	if err != nil {
		return nil, err
	}
	res, err := ParseApiInternalHighlevelConnectionStatusResponse(resp)

	return res, err
}

var _ = Describe("high level connection status", func() {
	It("get status for multiple different recipients", func() {
		satID := SatelliteId("bd54e0e9-5310-45be-b107-fd7c96672ce5")
		satOrgID := SatelliteOrgId("5")
		satelliteHost := []HostId{"c484f980-ab8d-401b-90e7-aa1d4ccf8c0e"}
		directConnectHost := []HostId{"fe30b997-c15a-44a9-89df-c236c3b5c540"}

		payload := ApiInternalHighlevelConnectionStatusJSONRequestBody{
			Hosts: []string{"c484f980-ab8d-401b-90e7-aa1d4ccf8c0e"},
			OrgId: "12345",
		}

		response, err := getConnectionStatus(payload)

		Expect(err).ToNot(HaveOccurred())

		result := response.JSON200
		Expect(response.StatusCode()).To(Equal(200))
		Expect(*result).To(HaveLen(2))
		Expect((*result)[0].Recipient).To(Equal(public.RunRecipient(uuid.MustParse("d415fc2d-9700-4e30-9621-6a410ccc92d8"))))
		Expect((*result)[0].RecipientType).To(Equal(Satellite))
		Expect((*result)[0].OrgId).To(Equal(payload.OrgId))
		Expect((*result)[0].SatId).To(Equal(satID))
		Expect((*result)[0].SatOrgId).To(Equal(satOrgID))
		Expect((*result)[0].Status).To(Equal(Connected))
		Expect((*result)[0].Systems).To(Equal(satelliteHost))

		Expect((*result)[1].Recipient).To(Equal(public.RunRecipient(uuid.MustParse("32af5948-301f-449a-a25b-ff34c83264a2"))))
		Expect((*result)[1].RecipientType).To(Equal(DirectConnect))
		Expect((*result)[1].OrgId).To(Equal(payload.OrgId))
		Expect((*result)[1].SatId).To(BeEmpty())
		Expect((*result)[1].SatOrgId).To(BeEmpty())
		Expect((*result)[1].Status).To(Equal(Connected))
		Expect((*result)[1].Systems).To(Equal(directConnectHost))
	})
	It("disallow more than 50 hosts", func() {

		hosts := make([]string, 51)
		for i := 0; i < 51; i++ {
			hosts[i] = "host" + strconv.Itoa(i+1)
		}

		payload := ApiInternalHighlevelConnectionStatusJSONRequestBody{
			Hosts: hosts,
			OrgId: "12345",
		}

		response, err := getConnectionStatus(payload)
		Expect(err).ToNot(HaveOccurred())
		Expect(response.StatusCode()).To(Equal(400))
	})
})
