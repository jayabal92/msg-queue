package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/testorg/msg-queue/internal/config"
	"go.uber.org/zap"
)

type fakeServer struct {
	started  bool
	shutdown bool
	startErr error
	shutErr  error
}

func (f *fakeServer) Start(ctx context.Context) error {
	f.started = true
	return f.startErr
}

func (f *fakeServer) Shutdown(ctx context.Context) error {
	f.shutdown = true
	return f.shutErr
}

func TestRun_StartAndShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{Server: config.ServerConfig{NodeID: "n1"}}
	fs := &fakeServer{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Simulate sending SIGINT after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)
	}()

	err := run(ctx, cfg, logger, nil, fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !fs.started || !fs.shutdown {
		t.Errorf("expected server to start and shutdown, got %+v", fs)
	}
}

func TestRun_StartFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{Server: config.ServerConfig{NodeID: "n2"}}
	fs := &fakeServer{startErr: context.Canceled}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := run(ctx, cfg, logger, nil, fs)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
