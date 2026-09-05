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
	for _, t := range md.Topics {
		if t.Topic == nil || t.ErrorCode != 0 {
			continue
		}
		if t.IsInternal || strings.HasPrefix(*t.Topic, "__") {
			continue
		}
		topics = append(topics, TopicInfo{Name: *t.Topic, Partitions: len(t.Partitions)})
	}
	sort.Slice(topics, func(i, j int) bool { return topics[i].Name < topics[j].Name })
	return topics, nil
}
