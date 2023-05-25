package broadcast

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type nodeBroadcast struct {
	topology map[string][]string
	seen     []int
	seenMap  map[int]struct{}
	lock     sync.RWMutex
	ctx      context.Context
	cancel   func()
}

func New() *nodeBroadcast {
	ctx, cancel := context.WithCancel(context.Background())
	return &nodeBroadcast{
		seen:    make([]int, 0, 1024),
		seenMap: make(map[int]struct{}, 1024),
		ctx:     ctx,
		cancel:  cancel,
	}
}

type broadcastBody struct {
	Type    string `json:"type,omitempty"`
	Message int    `json:"message,omitempty"`
}

func (sb *nodeBroadcast) BroadcastHandler(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) (err error) {
		var body broadcastBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		sb.lock.Lock()
		defer sb.lock.Unlock()
		if _, ok := sb.seenMap[body.Message]; !ok {
			sb.seenMap[body.Message] = struct{}{}
			sb.seen = append(sb.seen, body.Message)
		}

		resp := map[string]string{
			"type": "broadcast_ok",
		}
		return n.Reply(msg, resp)
	}
}

func (sb *nodeBroadcast) ReadHandler(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		sb.lock.RLock()
		defer sb.lock.RUnlock()
		resp := map[string]any{
			"type":     "read_ok",
			"messages": sb.seen,
		}

		return n.Reply(msg, resp)
	}
}

type topologyResponse struct {
	Type     string              `json:"type"`
	Topology map[string][]string `json:"topology"`
}

func (sb *nodeBroadcast) TopologyHandler(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body topologyResponse
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		sb.topology = body.Topology

		resp := map[string]string{
			"type": "topology_ok",
		}

		return n.Reply(msg, resp)
	}
}

type gossipBody struct {
	Type     string `json:"type,omitempty"`
	Messages []int  `json:"messages,omitempty"`
}

func (sb *nodeBroadcast) GossipHandler(n *maelstrom.Node) maelstrom.HandlerFunc {
	go sb.gossipLoop(sb.ctx, n)
	return func(msg maelstrom.Message) error {
		var body gossipBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		sb.lock.Lock()
		defer sb.lock.Unlock()
		for _, message := range body.Messages {
			if _, ok := sb.seenMap[message]; !ok {
				sb.seenMap[message] = struct{}{}
				sb.seen = append(sb.seen, message)
			}
		}
		return nil
	}
}

func (sb *nodeBroadcast) gossipLoop(ctx context.Context, n *maelstrom.Node) {
	timer := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-timer.C:
			var nodes []string
			cid := n.ID()

			nodes = sb.topology[cid]
			body := gossipBody{
				Type:     "gossip",
				Messages: sb.seen,
			}
			for _, id := range nodes {
				go func(id string) { _ = n.Send(id, body) }(id)
			}
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
}

func (sb *nodeBroadcast) Shutdown() {
	sb.cancel()
}
