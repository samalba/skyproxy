package main

import (
	"fmt"
	"log"
	"os"

	"github.com/samalba/skyproxy/client"
	"github.com/samalba/skyproxy/server"

	"github.com/codegangsta/cli"
)

func validateArgs(c *cli.Context, args []string) error {
	for _, arg := range args {
		if c.String(arg) == "" {
			fmt.Printf("Missing argument: --%s\n", arg)
			os.Exit(1)
		}
	}
	return nil
}

func globalCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "serve",
			Usage:  "Start a server",
			Action: runServer,
			Before: func(c *cli.Context) error {
				return validateArgs(c, []string{"address", "http-address"})
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "address",
					Value: "0.0.0.0:4242",
					Usage: "Address to listen on (for receiving skyproxy clients)",
				},
				cli.StringFlag{
					Name:  "http-address",
					Value: "0.0.0.0:80",
					Usage: "Address to listen on (for receiving HTTP traffic)",
				},
			},
		},
		{
			Name:   "connect",
			Usage:  "Connects to a local receiver",
			Action: runClient,
			Before: func(c *cli.Context) error {
				return validateArgs(c, []string{"server", "receiver", "http-host"})
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Value: "",
					Usage: "Remote skyproxy server address (ex: 0.0.0.0:1080)",
				},
				cli.StringFlag{
					Name:  "receiver",
					Value: "",
					Usage: "Local HTTP receiver to direct the traffic to (ex: localhost:8080)",
				},
				cli.StringFlag{
					Name:  "http-host",
					Value: "",
					Usage: "HTTP host to announce (ex: my.website.tld)",
				},
			},
		},
	}
}

func runClient(c *cli.Context) {
	server := c.String("server")
	receiver := c.String("receiver")
	httpHost := c.String("http-host")
	log.Printf("Connecting to server: %s", server)
	log.Printf("Registering HTTP Host: %s", httpHost)
	client := &client.Client{HTTPHost: httpHost, ReceiverAddr: receiver}
	if err := client.Connect(server); err != nil {
		log.Fatalf("Cannot connect: %s", err)
	}
	log.Printf("Connection established, forwarding the traffic to: %s", receiver)
	client.Forward()
}

func runServer(c *cli.Context) {
	serv := server.NewServer(c.String("address"), c.String("http-address"))
	// Listen for skyproxy clients
	go func() {
		if err := serv.ListenForReceivers(); err != nil {
			log.Fatal(err)
		}
	}()
	// Listen for HTTP traffic
	if err := serv.ListenForHTTP(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "skyproxy"
	app.Version = "0.1.0"
	app.Usage = "Reverse tunnel HTTP proxy"
	app.Commands = globalCommands()
	app.Run(os.Args)
}
