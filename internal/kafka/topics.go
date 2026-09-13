package kafka

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/twmb/franz-go/pkg/kmsg"
)

type TopicInfo struct {
	Name       string
	Partitions int
	// Messages — приблизительное число сообщений в топике: сумма end offsets
	// (high watermark) по партициям. -1, если узнать не удалось (retention мог
	// сдвинуть log start offset, поэтому это оценка, а не точный подсчёт).
	Messages int64
}

func (c *Client) ListTopics(ctx context.Context) ([]TopicInfo, error) {
	req := kmsg.NewPtrMetadataRequest()
	req.AllowAutoTopicCreation = false
	req.Topics = nil // nil = все топики
	md, err := c.cl.RequestCachedMetadata(ctx, req, 0)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}

	topics := make([]TopicInfo, 0, len(md.Topics))
	loReq := kmsg.NewPtrListOffsetsRequest()
	loReq.ReplicaID = -1
	for _, t := range md.Topics {
		if t.Topic == nil || t.ErrorCode != 0 {
			continue
		}
		if t.IsInternal || strings.HasPrefix(*t.Topic, "__") {
			continue
		}
		topics = append(topics, TopicInfo{Name: *t.Topic, Partitions: len(t.Partitions), Messages: -1})

		loTopic := kmsg.NewListOffsetsRequestTopic()
		loTopic.Topic = *t.Topic
		for _, p := range t.Partitions {
			lp := kmsg.NewListOffsetsRequestTopicPartition()
			lp.Partition = p.Partition
			lp.CurrentLeaderEpoch = -1
			lp.Timestamp = -1 // -1 = последний offset партиции (high watermark)
			loTopic.Partitions = append(loTopic.Partitions, lp)
		}
		loReq.Topics = append(loReq.Topics, loTopic)
	}

	if len(loReq.Topics) > 0 {
		c.fillMessageCounts(ctx, topics, loReq)
	}

	sort.Slice(topics, func(i, j int) bool { return topics[i].Name < topics[j].Name })
	return topics, nil
}

// fillMessageCounts запрашивает end offsets одним ListOffsets-запросом
// (kgo сам шардит его по лидерам партиций) и суммирует их по топикам.
// Ошибка запроса не фатальна: топики остаются с Messages = -1.
func (c *Client) fillMessageCounts(ctx context.Context, topics []TopicInfo, req *kmsg.ListOffsetsRequest) {
	raw, err := c.cl.Request(ctx, req)
	if err != nil {
		return
	}
	resp, ok := raw.(*kmsg.ListOffsetsResponse)
	if !ok {
		return
	}
	byName := make(map[string]int64, len(resp.Topics))
	for _, t := range resp.Topics {
		var sum int64
		for _, p := range t.Partitions {
			if p.ErrorCode != 0 || p.Offset < 0 {
				continue
			}
			sum += p.Offset
		}
		byName[t.Topic] = sum
	}
	for i := range topics {
		if sum, ok := byName[topics[i].Name]; ok {
			topics[i].Messages = sum
		}
	}
}
