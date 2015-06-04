package main

import (
	"log"
	"os"

	"github.com/samalba/skyproxy/server"

	"github.com/codegangsta/cli"
)

func globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "0.0.0.0:4242",
			Usage: "Address to connect to (or to listen to if launched with `serve' option)",
		},
	}
}

func globalCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "serve",
			Usage:  "Start a server",
			Action: runServer,
		},
	}
}

func runClient(c *cli.Context) {
	log.Println("TODO")
}

func runServer(c *cli.Context) {
	server := &server.Server{
		ListenAddress: c.String("address"),
	}
	if err := server.ListenForClients(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "skyproxy"
	app.Version = "0.1.0"
	app.Usage = "Reverse tunnel HTTP proxy"
	app.Flags = globalFlags()
	app.Commands = globalCommands()
	app.Action = runClient
	app.Run(os.Args)
}
