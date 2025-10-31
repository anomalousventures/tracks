package context

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/rs/zerolog"
)

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.InfoLevel)

	newCtx := WithLogger(ctx, logger)

	if newCtx == nil {
		t.Fatal("WithLogger returned nil context")
	}

	if newCtx == ctx {
		t.Error("WithLogger returned same context, expected new context")
	}
}

func TestGetLogger_WithLogger(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.InfoLevel)

	ctx = WithLogger(ctx, logger)
	retrieved := GetLogger(ctx)

	retrieved.Info().Msg("test message")

	if !bytes.Contains(buf.Bytes(), []byte("test message")) {
		t.Error("Retrieved logger did not write to expected buffer")
	}
}

func TestGetLogger_NoLogger(t *testing.T) {
	ctx := context.Background()

	logger := GetLogger(ctx)

	var buf bytes.Buffer
	logger = logger.Output(&buf)
	logger.Info().Msg("test message")

	if buf.Len() > 0 {
		t.Error("No-op logger should not produce output")
	}
}

func TestContextPropagation_Concurrent(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.InfoLevel)

	ctx = WithLogger(ctx, logger)

	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			retrieved := GetLogger(ctx)
			retrieved.Info().Int("goroutine", id).Msg("concurrent access")
		}(i)
	}

	wg.Wait()

	output := buf.String()
	for i := 0; i < concurrency; i++ {
		if !bytes.Contains([]byte(output), []byte("concurrent access")) {
			t.Error("Logger not safely accessible from concurrent goroutines")
			break
		}
	}
}

func TestContextPropagation_Nested(t *testing.T) {
	ctx := context.Background()
	var buf1, buf2 bytes.Buffer
	logger1 := zerolog.New(&buf1).Level(zerolog.InfoLevel)
	logger2 := zerolog.New(&buf2).Level(zerolog.InfoLevel)

	ctx1 := WithLogger(ctx, logger1)
	ctx2 := WithLogger(ctx1, logger2)

	logger1Retrieved := GetLogger(ctx1)
	logger1Retrieved.Info().Msg("message1")

	logger2Retrieved := GetLogger(ctx2)
	logger2Retrieved.Info().Msg("message2")

	if !bytes.Contains(buf1.Bytes(), []byte("message1")) {
		t.Error("First logger did not receive message")
	}

	if !bytes.Contains(buf2.Bytes(), []byte("message2")) {
		t.Error("Second logger did not receive message")
	}

	if bytes.Contains(buf1.Bytes(), []byte("message2")) {
		t.Error("First logger should not receive second message")
	}

	if bytes.Contains(buf2.Bytes(), []byte("message1")) {
		t.Error("Second logger should not receive first message")
	}
}
