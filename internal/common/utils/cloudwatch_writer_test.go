package utils

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// mockCloudWatchClient implements a mock CloudWatch Logs client for testing
type mockCloudWatchClient struct {
	createLogStreamFunc func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error)
	putLogEventsFunc    func(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error)
	calls               struct {
		createLogStream int
		putLogEvents    int
		mu              sync.Mutex
	}
}

func (m *mockCloudWatchClient) CreateLogStream(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
	m.calls.mu.Lock()
	m.calls.createLogStream++
	m.calls.mu.Unlock()
	if m.createLogStreamFunc != nil {
		return m.createLogStreamFunc(ctx, params, optFns...)
	}
	return &cloudwatchlogs.CreateLogStreamOutput{}, nil
}

func (m *mockCloudWatchClient) PutLogEvents(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error) {
	m.calls.mu.Lock()
	m.calls.putLogEvents++
	m.calls.mu.Unlock()
	if m.putLogEventsFunc != nil {
		return m.putLogEventsFunc(ctx, params, optFns...)
	}
	return &cloudwatchlogs.PutLogEventsOutput{
		NextSequenceToken: aws.String("next-token"),
	}, nil
}

func (m *mockCloudWatchClient) getCreateLogStreamCalls() int {
	m.calls.mu.Lock()
	defer m.calls.mu.Unlock()
	return m.calls.createLogStream
}

func (m *mockCloudWatchClient) getPutLogEventsCalls() int {
	m.calls.mu.Lock()
	defer m.calls.mu.Unlock()
	return m.calls.putLogEvents
}

// newMockCloudWatchWriter creates a CloudWatchWriter with a mock client for testing
func newMockCloudWatchWriter(mock *mockCloudWatchClient, logGroupName, logStreamName string) (*CloudWatchWriter, error) {
	writer := &CloudWatchWriter{
		client:        (*cloudwatchlogs.Client)(nil), // We'll bypass this in tests
		logGroupName:  logGroupName,
		logStreamName: logStreamName,
		buffer:        make([]types.InputLogEvent, 0),
		flushInterval: 100 * time.Millisecond, // Shorter for testing
		maxBufferSize: 10,
		done:          make(chan struct{}),
	}

	// Mock the client by creating a wrapper that uses our mock
	// For testing, we'll override the flush method to use the mock
	writer.mockClient = mock

	// Create log stream
	_, err := mock.CreateLogStream(context.Background(), &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})
	if err != nil {
		var resourceAlreadyExists *types.ResourceAlreadyExistsException
		if !errors.As(err, &resourceAlreadyExists) {
			return nil, err
		}
	}

	// Start background flusher
	go writer.backgroundFlusher()

	return writer, nil
}

func TestNewCloudWatchWriter(t *testing.T) {
	t.Run("creates writer successfully", func(t *testing.T) {
		mock := &mockCloudWatchClient{}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer writer.Close()

		if writer.logGroupName != "test-group" {
			t.Errorf("expected logGroupName to be 'test-group', got %s", writer.logGroupName)
		}
		if writer.logStreamName != "test-stream" {
			t.Errorf("expected logStreamName to be 'test-stream', got %s", writer.logStreamName)
		}
		if mock.getCreateLogStreamCalls() != 1 {
			t.Errorf("expected CreateLogStream to be called once, got %d", mock.getCreateLogStreamCalls())
		}
	})

	t.Run("handles existing log stream", func(t *testing.T) {
		mock := &mockCloudWatchClient{
			createLogStreamFunc: func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, &types.ResourceAlreadyExistsException{
					Message: aws.String("log stream already exists"),
				}
			},
		}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("expected no error when stream exists, got %v", err)
		}
		defer writer.Close()
	})

	t.Run("returns error on create stream failure", func(t *testing.T) {
		expectedErr := errors.New("aws error")
		mock := &mockCloudWatchClient{
			createLogStreamFunc: func(ctx context.Context, params *cloudwatchlogs.CreateLogStreamInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogStreamOutput, error) {
				return nil, expectedErr
			},
		}
		_, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestCloudWatchWriter_Write(t *testing.T) {
	t.Run("buffers log events", func(t *testing.T) {
		mock := &mockCloudWatchClient{}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		message := "test log message"
		n, err := writer.Write([]byte(message))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if n != len(message) {
			t.Errorf("expected to write %d bytes, wrote %d", len(message), n)
		}

		writer.bufferMutex.Lock()
		if len(writer.buffer) != 1 {
			t.Errorf("expected buffer to have 1 event, got %d", len(writer.buffer))
		}
		if *writer.buffer[0].Message != message {
			t.Errorf("expected message '%s', got '%s'", message, *writer.buffer[0].Message)
		}
		writer.bufferMutex.Unlock()
	})

	t.Run("flushes when buffer is full", func(t *testing.T) {
		mock := &mockCloudWatchClient{}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		// Write enough messages to trigger flush
		for i := 0; i < writer.maxBufferSize; i++ {
			_, err := writer.Write([]byte("test message"))
			if err != nil {
				t.Fatalf("write failed: %v", err)
			}
		}

		// Give flush goroutine time to execute
		time.Sleep(50 * time.Millisecond)

		// Verify PutLogEvents was called
		if mock.getPutLogEventsCalls() == 0 {
			t.Error("expected PutLogEvents to be called when buffer is full")
		}
	})
}

func TestCloudWatchWriter_Flush(t *testing.T) {
	t.Run("flushes buffered events", func(t *testing.T) {
		var capturedInput *cloudwatchlogs.PutLogEventsInput
		mock := &mockCloudWatchClient{
			putLogEventsFunc: func(ctx context.Context, params *cloudwatchlogs.PutLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutLogEventsOutput, error) {
				capturedInput = params
				return &cloudwatchlogs.PutLogEventsOutput{
					NextSequenceToken: aws.String("new-token"),
				}, nil
			},
		}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		// Write some messages
		_, err = writer.Write([]byte("message 1"))
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}
		_, err = writer.Write([]byte("message 2"))
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		// Manually flush using mock client
		err = writer.flushWithClient(mock)
		if err != nil {
			t.Fatalf("flush failed: %v", err)
		}

		if capturedInput == nil {
			t.Fatal("expected PutLogEvents to be called")
		}
		if len(capturedInput.LogEvents) != 2 {
			t.Errorf("expected 2 log events, got %d", len(capturedInput.LogEvents))
		}
		if *capturedInput.LogGroupName != "test-group" {
			t.Errorf("expected log group 'test-group', got '%s'", *capturedInput.LogGroupName)
		}
		if *capturedInput.LogStreamName != "test-stream" {
			t.Errorf("expected log stream 'test-stream', got '%s'", *capturedInput.LogStreamName)
		}
		if writer.sequenceToken == nil || *writer.sequenceToken != "new-token" {
			t.Error("expected sequence token to be updated")
		}
	})

	t.Run("does not flush empty buffer", func(t *testing.T) {
		mock := &mockCloudWatchClient{}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		initialCalls := mock.getPutLogEventsCalls()
		err = writer.flushWithClient(mock)
		if err != nil {
			t.Fatalf("flush failed: %v", err)
		}

		if mock.getPutLogEventsCalls() != initialCalls {
			t.Error("expected PutLogEvents not to be called for empty buffer")
		}
	})

	t.Run("clears buffer after flush", func(t *testing.T) {
		mock := &mockCloudWatchClient{}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		_, err = writer.Write([]byte("test message"))
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}
		err = writer.flushWithClient(mock)
		if err != nil {
			t.Fatalf("flush failed: %v", err)
		}

		writer.bufferMutex.Lock()
		if len(writer.buffer) != 0 {
			t.Errorf("expected buffer to be empty after flush, got %d events", len(writer.buffer))
		}
		writer.bufferMutex.Unlock()
	})
}

func TestCloudWatchWriter_BackgroundFlusher(t *testing.T) {
	t.Run("flushes periodically", func(t *testing.T) {
		mock := &mockCloudWatchClient{}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("failed to create writer: %v", err)
		}
		defer writer.Close()

		// Write a message
		_, err = writer.Write([]byte("test message"))
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		// Wait for background flush
		time.Sleep(150 * time.Millisecond)

		// Verify flush occurred
		if mock.getPutLogEventsCalls() == 0 {
			t.Error("expected background flusher to call PutLogEvents")
		}
	})
}

func TestCloudWatchWriter_Close(t *testing.T) {
	t.Run("flushes remaining events on close", func(t *testing.T) {
		mock := &mockCloudWatchClient{}
		writer, err := newMockCloudWatchWriter(mock, "test-group", "test-stream")
		if err != nil {
			t.Fatalf("failed to create writer: %v", err)
		}

		_, err = writer.Write([]byte("test message"))
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		initialCalls := mock.getPutLogEventsCalls()
		err = writer.closeWithClient(mock)
		if err != nil {
			t.Fatalf("close failed: %v", err)
		}

		if mock.getPutLogEventsCalls() <= initialCalls {
			t.Error("expected close to flush remaining events")
		}
	})
}
