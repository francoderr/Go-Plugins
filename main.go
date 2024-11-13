package main

import (
	"errors"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/shared"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	hclog "github.com/hashicorp/go-hclog"
)

func main() {
	// Create a hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command("./plugin/squarer"),
		Logger:          logger,
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("squarer")
	if err != nil {
		log.Fatal(err)
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	squarer := raw.(shared.Squarer)
	//fmt.Println(squarer.Square(50))

	e := echo.New()

	e.GET("/square/:num", func(e echo.Context) error {
		num := e.Param("num")
		if num == "" {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "missing parameter num"})
		}
		val, err := strconv.Atoi(num)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid parameter num"})
		}
		square := squarer.Square(val)
		return e.JSON(http.StatusOK, map[string]int{"square": square})
	})

	if err := e.Start(":8800"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("shutting down the server: %v", err)
	}

}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user-friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"squarer": &shared.SquarerPlugin{},
}
