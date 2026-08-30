package provider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/swuecho/chat_backend/models"
)

type streamModelFunc func(context.Context, Request) (<-chan StreamChunk, error)

func (f streamModelFunc) Stream(ctx context.Context, request Request) (<-chan StreamChunk, error) {
	return f(ctx, request)
}

func TestConsumeStreamCancelsProducerAfterDeliveryFailure(t *testing.T) {
	producerStopped := make(chan struct{})
	model := streamModelFunc(func(ctx context.Context, _ Request) (<-chan StreamChunk, error) {
		ch := make(chan StreamChunk)
		go func() {
			defer close(producerStopped)
			defer close(ch)
			if !emitChunk(ctx, ch, StreamChunk{Content: "one"}) {
				return
			}
			<-ctx.Done()
		}()
		return ch, nil
	})
	want := errors.New("client write failed")
	_, err := ConsumeStream(context.Background(), model, Request{}, func(StreamChunk) error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case <-producerStopped:
	case <-time.After(time.Second):
		t.Fatal("producer was not canceled")
	}
}

func TestConsumeStreamRequiresOneFinalAnswer(t *testing.T) {
	closed := streamModelFunc(func(context.Context, Request) (<-chan StreamChunk, error) {
		ch := make(chan StreamChunk)
		close(ch)
		return ch, nil
	})
	if _, err := ConsumeStream(context.Background(), closed, Request{}, nil); err == nil {
		t.Fatal("expected missing terminal answer to fail")
	}

	valid := streamModelFunc(func(context.Context, Request) (<-chan StreamChunk, error) {
		ch := make(chan StreamChunk, 1)
		ch <- StreamChunk{Done: true, FinalAnswer: &models.LLMAnswer{AnswerId: "a", Answer: "ok"}}
		close(ch)
		return ch, nil
	})
	answer, err := ConsumeStream(context.Background(), valid, Request{}, nil)
	if err != nil || answer.Answer != "ok" {
		t.Fatalf("answer=%#v err=%v", answer, err)
	}
}

func TestConsumeStreamHonorsCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	model := streamModelFunc(func(ctx context.Context, _ Request) (<-chan StreamChunk, error) {
		ch := make(chan StreamChunk)
		go func() {
			defer close(ch)
			<-ctx.Done()
		}()
		return ch, nil
	})
	cancel()
	_, err := ConsumeStream(ctx, model, Request{}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}
}
