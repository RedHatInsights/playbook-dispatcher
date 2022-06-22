package public

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/common/utils/test"
	"strings"

	. "github.com/onsi/gomega"
)

func fieldTester(fn func(params ...interface{}) *http.Response) func(...string) {
	doTest := func(params []interface{}, fields []string) {
		res := fn(params...)
		Expect(res.StatusCode).To(Equal(http.StatusOK))

		bodyBytes, err := ioutil.ReadAll(res.Body)
		Expect(err).ToNot(HaveOccurred())
		defer res.Body.Close()

		representation := make(map[string]interface{})
		err = json.Unmarshal(bodyBytes, &representation)
		Expect(err).ToNot(HaveOccurred())

		runs := representation["data"].([]interface{})
		Expect(runs[0]).To(HaveLen(len(fields)))
		for _, field := range fields {
			Expect(runs[0]).To(HaveKey(field))
		}
	}

	return func(fields ...string) {
		paramsShort := []interface{}{}
		paramsShort = append(paramsShort, "fields[data]")
		paramsShort = append(paramsShort, strings.Join(fields, ","))

		params := []interface{}{}
		for _, value := range fields {
			params = append(params, "fields[data]")
			params = append(params, value)
		}

		doTest(params, fields)
		doTest(paramsShort, fields)
	}
}

func doGet(baseUrl string, keysAndValues ...interface{}) *http.Response {
	url := utils.BuildUrl(baseUrl, keysAndValues...)

	req, err := http.NewRequest("GET", url, nil)
	Expect(err).ToNot(HaveOccurred())
	req.Header.Set("x-rh-identity", test.IdentityHeaderMinimal(accountNumber()))
	resp, err := test.Client.Do(req)
	Expect(err).ToNot(HaveOccurred())
	return resp
}
