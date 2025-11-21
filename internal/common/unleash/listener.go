package unleash

import (
	unleash "github.com/Unleash/unleash-go-sdk/v5"
	"go.uber.org/zap"
)

// DispatcherListener implements Unleash listener interfaces for logging
type DispatcherListener struct {
	log *zap.SugaredLogger
}

// NewDispatcherListener creates a new listener with logger
func NewDispatcherListener(log *zap.SugaredLogger) *DispatcherListener {
	return &DispatcherListener{log: log}
}

// OnError logs errors from Unleash client
func (l *DispatcherListener) OnError(err error) {
	l.log.Warnw("Unleash client error", "error", err)
}

// OnWarning logs warnings from Unleash client
func (l *DispatcherListener) OnWarning(warning error) {
	l.log.Warnw("Unleash client warning", "warning", warning)
}

// OnReady logs when the Unleash client is ready
func (l *DispatcherListener) OnReady() {
	l.log.Info("Unleash client ready")
}

// OnCount prints to the console when the feature is queried
func (l *DispatcherListener) OnCount(name string, enabled bool) {
	// Intentionally not logged - called very frequently
}

// OnSent is called when metrics are uploaded to Unleash
func (l *DispatcherListener) OnSent(payload unleash.MetricsData) {
	// Intentionally not logged - called on every request
}

// OnRegistered logs when the client has registered with Unleash
func (l *DispatcherListener) OnRegistered(payload unleash.ClientData) {
	l.log.Debugw("Unleash client registered", "payload", payload)
}
