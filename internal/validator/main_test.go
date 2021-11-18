package validator

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Validator", func() {
	Describe("Kafka message predicate", func() {
		It("accepts a playbook payload", func() {
			predicate := newPlaybookServiceMessagePredicate(zap.NewNop().Sugar())

			Expect(predicate(&kafka.Message{
				Headers: []kafka.Header{{Key: "service", Value: []byte("playbook")}},
			})).To(BeTrue())
		})

		It("rejects non-playbook payload", func() {
			predicate := newPlaybookServiceMessagePredicate(zap.NewNop().Sugar())

			Expect(predicate(&kafka.Message{
				Headers: []kafka.Header{{Key: "service", Value: []byte("advisor")}},
			})).To(BeFalse())
		})

		It("rejects on missing header", func() {
			predicate := newPlaybookServiceMessagePredicate(zap.NewNop().Sugar())

			topic := "payloads"
			Expect(predicate(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: 0,
					Offset:    kafka.Offset(0),
				},
				Headers: []kafka.Header{},
			})).To(BeFalse())
		})
	})
})
