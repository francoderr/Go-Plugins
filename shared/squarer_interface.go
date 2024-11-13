package shared

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// Greeter is the interface that we're exposing as a plugin.
type Squarer interface {
	Square(val int) int
}

// Here is an implementation that talks over RPC
type SquarerRPC struct{ client *rpc.Client }

func (g *SquarerRPC) Square(val int) int {
	var resp int
	err := g.client.Call("Plugin.Square", val, &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

// Here is the RPC server that GreeterRPC talks to, conforming to
// the requirements of net/rpc
type SquarerRPCServer struct {
	// This is the real implementation
	Impl Squarer
}

func (s *SquarerRPCServer) Square(args int, resp *int) error {
	*resp = s.Impl.Square(args)
	return nil
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a GreeterRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return GreeterRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type SquarerPlugin struct {
	// Impl Injection
	Impl Squarer
}

func (p *SquarerPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &SquarerRPCServer{Impl: p.Impl}, nil
}

func (SquarerPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &SquarerRPC{client: c}, nil
}
