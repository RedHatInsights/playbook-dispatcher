package kafka

import (
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// kafka_consumer_lag is scraped periodically from the consumer event loop
// (not per message) to avoid watermark RPCs on the hot path.
//
// Each pod reports only its assigned partitions; PromQL can aggregate across replicas.
// Stale series after rebalance are left to Prometheus staleness handling rather than
// manual DeleteLabelValues bookkeeping.
var consumerLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "kafka_consumer_lag",
	Help: "Kafka consumer lag (high watermark - committed offset) per partition",
}, []string{"topic", "partition", "consumer_group"})

// kafka_consumer_lag_scrape_errors_total surfaces failed lag scrapes so silent
// metric staleness is visible in dashboards/alerts.
var consumerLagScrapeErrors = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "kafka_consumer_lag_scrape_errors_total",
	Help: "Errors encountered while scraping kafka_consumer_lag",
}, []string{"consumer_group", "reason"})

const (
	scrapeReasonAssignment = "assignment"
	scrapeReasonCommitted  = "committed"
	scrapeReasonWatermark  = "watermark"
)

// reportConsumerLag refreshes kafka_consumer_lag for this consumer's assignment.
//
// Must be called from the same goroutine that owns ReadMessage — Consumer is
// not concurrency-safe.
//
// Lag = QueryWatermarkOffsets high − Committed offset (confluent-kafka-go has
// no Lag field). Committed (not Position) matches group-acked progress; with
// auto-commit this can trail the in-memory read position by up to
// auto.commit.interval.ms.
func reportConsumerLag(log *zap.SugaredLogger, consumer *kafka.Consumer, groupID string, timeoutMs int) {
	assignment, err := consumer.Assignment()
	if err != nil {
		consumerLagScrapeErrors.WithLabelValues(groupID, scrapeReasonAssignment).Inc()
		log.Warnw("Failed to read kafka consumer assignment for lag scrape", "err", err, "consumer_group", groupID)
		return
	}
	if len(assignment) == 0 {
		// Startup or mid-rebalance — not an error.
		return
	}

	committed, err := consumer.Committed(assignment, timeoutMs)
	if err != nil {
		consumerLagScrapeErrors.WithLabelValues(groupID, scrapeReasonCommitted).Inc()
		log.Warnw("Failed to read kafka committed offsets for lag scrape", "err", err, "consumer_group", groupID)
		return
	}

	for _, tp := range committed {
		if tp.Topic == nil {
			continue
		}
		// OffsetInvalid: nothing committed yet — skip to avoid a false huge lag.
		if tp.Offset < 0 {
			continue
		}

		_, high, err := consumer.QueryWatermarkOffsets(*tp.Topic, tp.Partition, timeoutMs)
		if err != nil {
			consumerLagScrapeErrors.WithLabelValues(groupID, scrapeReasonWatermark).Inc()
			log.Warnw("Failed to query kafka watermark offsets for lag scrape",
				"err", err,
				"consumer_group", groupID,
				"topic", *tp.Topic,
				"partition", tp.Partition,
			)
			continue
		}

		partition := strconv.Itoa(int(tp.Partition))
		consumerLag.WithLabelValues(*tp.Topic, partition, groupID).Set(lagFromOffsets(int64(tp.Offset), high))
	}
}

// lagFromOffsets returns high − committed, clamped at 0.
func lagFromOffsets(committed, high int64) float64 {
	lag := float64(high - committed)
	if lag < 0 {
		return 0
	}
	return lag
}
