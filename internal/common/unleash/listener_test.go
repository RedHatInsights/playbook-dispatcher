package unleash

import (
	"errors"
	"testing"

	"github.com/Unleash/unleash-go-sdk/v6"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewDispatcherListener(t *testing.T) {
	log := zap.NewNop().Sugar()

	listener := NewDispatcherListener(log)

	assert.NotNil(t, listener)
	assert.Equal(t, log, listener.log)
}

func TestDispatcherListener_OnError(t *testing.T) {
	core, recorded := observer.New(zap.WarnLevel)
	log := zap.New(core).Sugar()
	listener := NewDispatcherListener(log)

	testErr := errors.New("test error")
	listener.OnError(testErr)

	assert.Equal(t, 1, recorded.Len())
	entry := recorded.All()[0]
	assert.Equal(t, "Unleash client error", entry.Message)
	assert.Equal(t, zap.WarnLevel, entry.Level)
}

func TestDispatcherListener_OnWarning(t *testing.T) {
	core, recorded := observer.New(zap.WarnLevel)
	log := zap.New(core).Sugar()
	listener := NewDispatcherListener(log)

	testWarning := errors.New("test warning")
	listener.OnWarning(testWarning)

	assert.Equal(t, 1, recorded.Len())
	entry := recorded.All()[0]
	assert.Equal(t, "Unleash client warning", entry.Message)
	assert.Equal(t, zap.WarnLevel, entry.Level)
}

func TestDispatcherListener_OnReady(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	log := zap.New(core).Sugar()
	listener := NewDispatcherListener(log)

	listener.OnReady()

	assert.Equal(t, 1, recorded.Len())
	entry := recorded.All()[0]
	assert.Equal(t, "Unleash client ready", entry.Message)
	assert.Equal(t, zap.InfoLevel, entry.Level)
}

func TestDispatcherListener_OnCount(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	log := zap.New(core).Sugar()
	listener := NewDispatcherListener(log)

	// OnCount should not log anything (intentionally silent)
	listener.OnCount("test-feature", true)

	assert.Equal(t, 0, recorded.Len())
}

func TestDispatcherListener_OnSent(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	log := zap.New(core).Sugar()
	listener := NewDispatcherListener(log)

	// OnSent should not log anything (intentionally silent)
	payload := unleash.MetricsData{}
	listener.OnSent(payload)

	assert.Equal(t, 0, recorded.Len())
}

func TestDispatcherListener_OnRegistered(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	log := zap.New(core).Sugar()
	listener := NewDispatcherListener(log)

	payload := unleash.ClientData{
		AppName: "test-app",
	}
	listener.OnRegistered(payload)

	assert.Equal(t, 1, recorded.Len())
	entry := recorded.All()[0]
	assert.Equal(t, "Unleash client registered", entry.Message)
	assert.Equal(t, zap.DebugLevel, entry.Level)
}
