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
				return validateArgs(c, []string{"address"})
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "address",
					Value: "0.0.0.0:80",
					Usage: "Address to listen on",
				},
				cli.StringFlag{
					Name:  "tls-cert",
					Value: "",
					Usage: "TLS Certificate file (disabled by default)",
				},
				cli.StringFlag{
					Name:  "tls-key",
					Value: "",
					Usage: "TLS Key file (disabled by default)",
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
				cli.StringFlag{
					Name:  "tls-ca",
					Value: "",
					Usage: "TLS CA Certificate file (disabled by default)",
				},
			},
		},
	}
}

func runClient(c *cli.Context) {
	server := c.String("server")
	receiver := c.String("receiver")
	httpHost := c.String("http-host")
	tlsCA := c.String("tls-ca")
	log.SetPrefix("[client] ")
	log.Printf("Connecting to server: %s", server)
	log.Printf("Registering HTTP Host: %s", httpHost)
	skyClient := &client.Client{HTTPHost: httpHost}
	if tlsCA != "" {
		tlsConfig := &client.TLSConfig{
			CAFile: tlsCA,
		}
		if err := skyClient.Connect(server, tlsConfig); err != nil {
			log.Fatalf("Cannot connect: %s", err)
		}
	} else {
		if err := skyClient.Connect(server, nil); err != nil {
			log.Fatalf("Cannot connect: %s", err)
		}
	}
	log.Printf("Connection established, forwarding the traffic to: %s", receiver)
	skyClient.Tunnel(receiver)
}

func runServer(c *cli.Context) {
	address := c.String("address")
	tlsCert := c.String("tls-cert")
	tlsKey := c.String("tls-key")
	log.SetPrefix("[server] ")
	serv := server.NewServer()
	if tlsCert != "" && tlsKey != "" {
		tlsConfig := &server.TLSConfig{
			CertFile: tlsCert,
			KeyFile:  tlsKey,
		}
		// Start the HTTPS server
		log.Printf("Starting HTTPS server at %s", address)
		if err := serv.StartServer(address, tlsConfig); err != nil {
			log.Fatal(err)
		}
		return
	}
	// Start the HTTP server
	log.Printf("Starting HTTP server at %s", address)
	if err := serv.StartServer(address, nil); err != nil {
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
