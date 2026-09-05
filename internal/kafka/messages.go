package kafka

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
)

type Message struct {
	Partition int32
	Offset    int64
	Timestamp time.Time
	Key       []byte
	Value     []byte
}

// TailMessages возвращает последние limit сообщений топика:
// для каждой партиции читаем хвост от (high watermark - limit),
// затем оставляем limit самых свежих по времени.
func (c *Client) TailMessages(ctx context.Context, topic string, limit int) ([]Message, error) {
	partitions, err := c.topicPartitions(ctx, topic)
	if err != nil {
		return nil, err
	}
	if len(partitions) == 0 {
		return nil, nil
	}

	// Прямое потребление конкретных партиций с хвоста.
	// AddConsumePartitions работает только для direct-консьюмера (без групп),
	// что соответствует нашему клиенту.
	assign := map[string]map[int32]kgo.Offset{topic: {}}
	for _, p := range partitions {
		assign[topic][p] = kgo.NewOffset().AtEnd().Relative(-int64(limit))
	}
	c.mu.Lock()
	c.cl.AddConsumePartitions(assign)
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.cl.RemoveConsumePartitions(map[string][]int32{topic: partitions})
		c.mu.Unlock()
	}()

	pollCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var (
		out     []Message
		pollErr error
	)
	for pollCtx.Err() == nil {
		fetches := c.cl.PollRecords(pollCtx, limit*len(partitions))
		if fetches.IsClientClosed() {
			break
		}
		fetches.EachError(func(_ string, _ int32, err error) {
			pollErr = err
		})
		fetches.EachRecord(func(r *kgo.Record) {
			out = append(out, Message{
				Partition: r.Partition,
				Offset:    r.Offset,
				Timestamp: r.Timestamp,
				Key:       r.Key,
				Value:     r.Value,
			})
		})
		if len(out) >= limit*len(partitions) {
			break
		}
	}
	if len(out) == 0 && pollErr != nil {
		return nil, fmt.Errorf("poll %q: %w", topic, pollErr)
	}

	sort.Slice(out, func(i, j int) bool {
		if !out[i].Timestamp.Equal(out[j].Timestamp) {
			return out[i].Timestamp.Before(out[j].Timestamp)
		}
		if out[i].Partition != out[j].Partition {
			return out[i].Partition < out[j].Partition
		}
		return out[i].Offset < out[j].Offset
	})
	if len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

// topicPartitions возвращает список партиций топика через metadata-запрос.
func (c *Client) topicPartitions(ctx context.Context, topic string) ([]int32, error) {
	req := kmsg.NewPtrMetadataRequest()
	req.AllowAutoTopicCreation = false
	reqTopic := kmsg.NewMetadataRequestTopic()
	reqTopic.Topic = kmsg.StringPtr(topic)
	req.Topics = []kmsg.MetadataRequestTopic{reqTopic}
	md, err := c.cl.RequestCachedMetadata(ctx, req, 0)
	if err != nil {
		return nil, fmt.Errorf("metadata for %q: %w", topic, err)
	}
	if len(md.Topics) == 0 {
		return nil, fmt.Errorf("topic %q not found", topic)
	}
	t := &md.Topics[0]
	if t.Topic == nil {
		return nil, fmt.Errorf("topic %q: %w", topic, kerr.ErrorForCode(t.ErrorCode))
	}
	if t.ErrorCode != 0 {
		return nil, fmt.Errorf("topic %q: %w", topic, kerr.ErrorForCode(t.ErrorCode))
	}
	ps := make([]int32, 0, len(t.Partitions))
	for _, p := range t.Partitions {
		ps = append(ps, p.Partition)
	}
	return ps, nil
}
