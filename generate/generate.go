package generate

import (
	"fmt"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var lock sync.Mutex

type RespBody struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func Handler(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		lock.Lock()
		now := time.Now().Nanosecond()
		lock.Unlock()
		body := RespBody{
			Type: "generate_ok",
			ID:   fmt.Sprintf("%s_%d", msg.Dest, now),
		}

		return n.Reply(msg, body)
	}
}
