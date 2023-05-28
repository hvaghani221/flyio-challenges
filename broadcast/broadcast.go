package broadcast

import (
	"context"
	"encoding/json"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type nodeBroadcast struct {
	topology map[string][]string
	messages *messages
	known    map[string]*messages
	ctx      context.Context
	cancel   func()
}

func New() *nodeBroadcast {
	ctx, cancel := context.WithCancel(context.Background())
	return &nodeBroadcast{
		messages: newMessages(),
		known:    make(map[string]*messages),
		ctx:      ctx,
		cancel:   cancel,
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

		sb.messages.Add(body.Message)

		resp := map[string]string{
			"type": "broadcast_ok",
		}
		return n.Reply(msg, resp)
	}
}

func (sb *nodeBroadcast) ReadHandler(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		resp := map[string]any{
			"type":     "read_ok",
			"messages": sb.messages.ReadAll(),
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
		if sb.topology == nil {
			var body topologyResponse
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				return err
			}
			sb.topology = body.Topology
			for n := range sb.topology {
				sb.known[n] = newMessagesSeenOnly()
			}
		}
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

		sb.messages.AddAll(body.Messages)
		sb.known[msg.Src].RememberAll(body.Messages)
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
			for _, id := range nodes {
				messagesCopy := sb.messages.ReadAll()
				go func(id string) {
					body := gossipBody{
						Type:     "gossip",
						Messages: sb.known[id].ReadFiltered(messagesCopy),
					}
					_ = n.Send(id, body)
				}(id)
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
