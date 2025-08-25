package client

import (
	"context"
	"hash/fnv"

	"github.com/testorg/msg-queue/proto"
	"google.golang.org/grpc"
)

type Producer struct {
	brokers []string
	cli     proto.MQClient
}

func NewProducer(addr string, opts ...grpc.DialOption) (*Producer, error) {
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Producer{cli: proto.NewMQClient(conn)}, nil
}

func (p *Producer) partitionFor(key []byte, n int) int32 {
	h := fnv.New32a()
	h.Write(key)
	return int32(h.Sum32() % uint32(n))
}

func (p *Producer) Produce(ctx context.Context, topic string, key, val []byte, parts int, acks proto.ProduceRequest_Acks) (int32, int64, error) {
	part := p.partitionFor(key, parts)
	res, err := p.cli.Produce(ctx, &proto.ProduceRequest{Topic: topic, Partition: part, Messages: []*proto.Message{{Key: key, Value: val}}, Acks: acks})
	if err != nil {
		return 0, 0, err
	}
	return res.Partition, res.Offsets[0], nil
}
