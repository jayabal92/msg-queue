package client

import (
	"context"

	"github.com/testorg/msg-queue/proto"
	"google.golang.org/grpc"
)

type Consumer struct{ cli proto.MQClient }

func NewConsumer(addr string, opts ...grpc.DialOption) (*Consumer, error) {
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Consumer{cli: proto.NewMQClient(conn)}, nil
}

func (c *Consumer) Fetch(ctx context.Context, topic string, partition int32, offset int64, max int32) (*proto.FetchResponse, error) {
	return c.cli.Fetch(ctx, &proto.FetchRequest{Topic: topic, Partition: partition, Offset: offset, MaxMessages: max})
}

func (c *Consumer) Commit(ctx context.Context, group string, topic string, partition int32, nextOffset int64) error {
	_, err := c.cli.CommitOffsets(ctx, &proto.CommitOffsetsRequest{GroupId: group, Topic: topic, Partition: partition, Offset: nextOffset})
	return err
}
