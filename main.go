package main

import (
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"

	"github.com/hvaghani221/flyio-challenges/broadcast"
	"github.com/hvaghani221/flyio-challenges/echo"
	"github.com/hvaghani221/flyio-challenges/generate"
)

func main() {
	n := maelstrom.NewNode()

	// Register a handler for the "echo" message that responds with an "echo_ok".
	n.Handle("echo", echo.Handler(n))
	n.Handle("generate", generate.Handler(n))

	broadcast := broadcast.New()
	defer broadcast.Shutdown()
	n.Handle("broadcast", broadcast.BroadcastHandler(n))
	n.Handle("read", broadcast.ReadHandler(n))
	n.Handle("topology", broadcast.TopologyHandler(n))
	n.Handle("gossip", broadcast.GossipHandler(n))

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
