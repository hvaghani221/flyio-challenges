package main

import (
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"

	"github.com/hvaghani221/flyio-challenges/counter"
)

func main() {
	n := maelstrom.NewNode()

	// n.Handle("echo", echo.Handler(n))
	// n.Handle("generate", generate.Handler(n))

	// broadcast := broadcast.New()
	// defer broadcast.Shutdown()
	// n.Handle("broadcast", broadcast.BroadcastHandler(n))
	// n.Handle("read", broadcast.ReadHandler(n))
	// n.Handle("topology", broadcast.TopologyHandler(n))
	// n.Handle("gossip", broadcast.GossipHandler(n))

	c := counter.NewCounter(n)
	defer c.Shutdown()
	n.Handle("add", c.Add(n))
	n.Handle("read", c.Read(n))

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
