package broker

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/testorg/msg-queue/internal/config"
	"github.com/testorg/msg-queue/internal/metadata"
	"github.com/testorg/msg-queue/proto"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// --- fake metadata.Store for unit test ---
type fakeStore struct {
	*metadata.Store // embed real struct so type matches
	topics          map[string]metadata.TopicDesc
	brokers         map[string]string
	partMeta        *metadata.PartitionState
}

func (f *fakeStore) RegisterBroker(ctx context.Context, id, addr string, ttl int64) error {
	f.brokers[id] = addr
	return nil
}
func (f *fakeStore) ListBrokers(ctx context.Context) (map[string]string, error) {
	return f.brokers, nil
}
func (f *fakeStore) GetTopic(ctx context.Context, t string) (*metadata.TopicDesc, error) {
	if td, ok := f.topics[t]; ok {
		return &td, nil
	}
	return nil, context.Canceled
}
func (f *fakeStore) GetPartitionState(ctx context.Context, t string, p int) (*metadata.PartitionState, error) {
	return f.partMeta, nil
}
func (f *fakeStore) CreateTopic(ctx context.Context, td metadata.TopicDesc) error {
	f.topics[td.Name] = td
	return nil
}
func (f *fakeStore) AssignInitialLeaders(ctx context.Context, topic string, parts int, rf int) error {
	return nil
}
func (f *fakeStore) CommitOffset(ctx context.Context, group, topic string, part int, off int64) error {
	return nil
}
func (f *fakeStore) FetchGroupOffsets(ctx context.Context, group, topic string) ([]metadata.GroupPartitionOffset, error) {
	return nil, nil
}

// --- helper to spin up in-memory server ---
func startTestBroker(t *testing.T, dir string) (*Server, proto.MQClient, func()) {
	z := zap.NewNop()
	cfg := config.Config{
		Server: config.ServerConfig{
			NodeID:   "n1",
			GRPCAddr: "127.0.0.1:0",
			HTTPAddr: "127.0.0.1:0",
		},
		Storage: config.StorageConfig{
			DataDir:      dir,
			SegmentBytes: 1024,
		},
	}
	// fake etcd
	etcd := &clientv3.Client{} // unused in fake store
	s := NewServer(cfg, z, etcd)

	// override store with fake one
	s.store = &fakeStore{
		topics: map[string]metadata.TopicDesc{
			"events": {Name: "events", Partitions: 1, RF: 1},
		},
		brokers: map[string]string{"n1": "127.0.0.1:0"},
		partMeta: &metadata.PartitionState{
			Leader:      "n1",
			LeaderEpoch: 1,
			ISR:         []string{"n1"},
		},
	}

	// real GRPC server on random port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen err: %v", err)
	}
	s.GRPCLis = lis
	s.GRPCSrv = grpc.NewServer()
	proto.RegisterMQServer(s.GRPCSrv, s)

	go s.GRPCSrv.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	client := proto.NewMQClient(conn)

	cleanup := func() {
		s.Shutdown(context.Background())
		_ = conn.Close()
	}
	return s, client, cleanup
}

func TestProduceAndFetch(t *testing.T) {
	dir := t.TempDir()
	_, client, cleanup := startTestBroker(t, dir)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- Produce ---
	msg := &proto.Message{Key: []byte("k1"), Value: []byte("v1")}
	resp, err := client.Produce(ctx, &proto.ProduceRequest{
		Topic:     "events",
		Partition: 0,
		Messages:  []*proto.Message{msg},
	})
	if err != nil {
		t.Fatalf("Produce error: %v", err)
	}
	if len(resp.Offsets) != 1 || resp.Offsets[0] != 0 {
		t.Fatalf("expected offset=0 got %v", resp.Offsets)
	}

	// --- Fetch ---
	fresp, err := client.Fetch(ctx, &proto.FetchRequest{
		Topic:       "events",
		Partition:   0,
		Offset:      0,
		MaxMessages: 10,
	})
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if len(fresp.Records) != 1 {
		t.Fatalf("expected 1 record got %d", len(fresp.Records))
	}
	got := fresp.Records[0].Message
	if string(got.Value) != "v1" {
		t.Fatalf("expected value v1 got %s", got.Value)
	}
}
