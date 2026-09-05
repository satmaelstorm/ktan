package kafka

import (
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Client struct {
	cl *kgo.Client
	// mu защищает переключение потребляемых партиций: bubbletea-команды
	// выполняются в отдельных горутинах и могут пересекаться.
	mu sync.Mutex
}

func New(brokers []string) (*Client, error) {
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return nil, err
	}
	return &Client{cl: cl}, nil
}

func (c *Client) Close() {
	c.cl.Close()
}
