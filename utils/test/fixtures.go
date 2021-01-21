package test

import (
	"fmt"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
)

func WithAccountNumber() func() string {
	var base = uuid.New().String()[29:]
	var test int

	BeforeEach(func() {
		test++
	})

	return func() string {
		return fmt.Sprintf("%s-%d", base, test)
	}
}
