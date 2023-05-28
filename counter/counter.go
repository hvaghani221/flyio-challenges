package counter

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const KEY = "total"

type counter struct {
	value     *atomic.Int64
	kv        *maelstrom.KV
	ctx       context.Context
	cancel    context.CancelFunc
	valueChan chan int
	once      *sync.Once
}

func NewCounter(n *maelstrom.Node) *counter {
	ctx, cancel := context.WithCancel(context.Background())
	c := &counter{
		kv:        maelstrom.NewSeqKV(n),
		value:     &atomic.Int64{},
		ctx:       ctx,
		cancel:    cancel,
		valueChan: make(chan int, 1024),
		once:      &sync.Once{},
	}

	return c
}

func (c *counter) Read(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		resp := map[string]any{
			"type":  "read_ok",
			"value": c.value.Load(),
		}

		return n.Reply(msg, resp)
	}
}

type addMessage struct {
	Delta int `json:"delta"`
}

func (c *counter) Add(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		c.once.Do(func() {
			_ = c.kv.CompareAndSwap(c.ctx, KEY, 0, 0, true)
			go c.worker(n)
		})
		var body addMessage
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		c.valueChan <- body.Delta

		resp := map[string]any{
			"type": "add_ok",
		}
		return n.Reply(msg, resp)
	}
}

func (c *counter) worker(n *maelstrom.Node) {
	timer := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-c.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			val, err := c.kv.ReadInt(c.ctx, KEY)
			if err != nil {
				continue
			}
			c.value.Swap(int64(val))
		case delta := <-c.valueChan:
			for {
				val, err := c.kv.ReadInt(c.ctx, KEY)
				if err != nil {
					continue
				}
				if err := c.kv.CompareAndSwap(c.ctx, KEY, val, val+delta, true); err != nil {
					continue
				}
				c.value.Store(int64(val + delta))
				break
			}
		}
	}
}

func (c *counter) Shutdown() {
	c.cancel()
}
