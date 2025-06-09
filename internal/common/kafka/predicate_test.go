package kafka

import (
	k "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Kafka", func() {
	Describe("Kafka message predicate", func() {
		It("accepts a playbook and playbook-sat payloads", func() {
			predicate := FilterByHeaderPredicate(zap.NewNop().Sugar(), "service", "playbook", "playbook-sat")

			Expect(predicate(&k.Message{
				Headers: []k.Header{{Key: "service", Value: []byte("playbook")}},
			})).To(BeTrue())
			Expect(predicate(&k.Message{
				Headers: []k.Header{{Key: "service", Value: []byte("playbook-sat")}},
			})).To(BeTrue())
		})

		It("rejects non-playbook payload", func() {
			predicate := FilterByHeaderPredicate(zap.NewNop().Sugar(), "service", "playbook")

			Expect(predicate(&k.Message{
				Headers: []k.Header{{Key: "service", Value: []byte("advisor")}},
			})).To(BeFalse())
		})

		It("rejects on missing header", func() {
			predicate := FilterByHeaderPredicate(zap.NewNop().Sugar(), "service", "playbook")

			topic := "payloads"
			Expect(predicate(&k.Message{
				TopicPartition: k.TopicPartition{
					Topic:     &topic,
					Partition: 0,
					Offset:    k.Offset(0),
				},
				Headers: []k.Header{},
			})).To(BeFalse())
		})
	})
})
