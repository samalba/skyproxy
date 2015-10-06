package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/samalba/skyproxy/client"
	"github.com/samalba/skyproxy/server"

	"github.com/codegangsta/cli"
)

const (
	cliOneArg        = iota
	cliAllArgs       = iota
	cliAllIfFirstArg = iota
)

func requireArgs(c *cli.Context, option int, args []string) error {
	for _, arg := range args {
		if option == cliAllArgs && c.String(arg) == "" {
			fmt.Printf("You need to specify all the following arguments: %s\n", arg)
			os.Exit(1)
		}
		if option == cliOneArg && c.String(arg) != "" {
			return nil
		}
		if option == cliAllIfFirstArg && c.String(args[0]) != "" && c.String(arg) == "" {

			fmt.Printf("After setting the argument \"%s\", you need to specify all the following arguments: %s\n", args[0], args)
			os.Exit(1)
		}
	}
	if option == cliOneArg {
		fmt.Printf("You need to specify one of the following argument: %s\n", args)
		os.Exit(1)
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
				if err := requireArgs(c, cliOneArg, []string{"proxy-http", "proxy-https"}); err != nil {
					return err
				}
				if err := requireArgs(c, cliAllIfFirstArg, []string{"proxy-https", "proxy-tls-cert", "proxy-tls-key"}); err != nil {
					return err
				}
				if err := requireArgs(c, cliAllIfFirstArg, []string{"clients-https", "clients-tls-cert", "clients-tls-key"}); err != nil {
					return err
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "proxy-http",
					Value: "",
					Usage: "HTTP proxy address to listen on (ex: \":80\")",
				},
				cli.StringFlag{
					Name:  "proxy-https",
					Value: "",
					Usage: "HTTPs proxy address to listen on (ex: \":443\")",
				},
				cli.StringFlag{
					Name:  "proxy-tls-cert",
					Value: "",
					Usage: "TLS Certificate file (use with --listen-https)",
				},
				cli.StringFlag{
					Name:  "proxy-tls-key",
					Value: "",
					Usage: "TLS Key file (use with --listen-https)",
				},
				cli.StringFlag{
					Name:  "clients-http",
					Value: "",
					Usage: "HTTP address to listen to SkyProxy clients (ex: \":80\")",
				},
				cli.StringFlag{
					Name:  "clients-https",
					Value: "",
					Usage: "HTTPs address to listen to SkyProxy clients (ex: \":443\")",
				},
				cli.StringFlag{
					Name:  "clients-tls-cert",
					Value: "",
					Usage: "TLS Certificate file (use with --clients-https)",
				},
				cli.StringFlag{
					Name:  "clients-tls-key",
					Value: "",
					Usage: "TLS Key file (use with --clients-https)",
				},
			},
		},
		{
			Name:   "connect",
			Usage:  "Connects to a local receiver",
			Action: runClient,
			Before: func(c *cli.Context) error {
				return requireArgs(c, cliAllArgs, []string{"server", "receiver", "http-host"})
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
	var wg sync.WaitGroup
	proxyHTTP := c.String("proxy-http")
	proxyHTTPS := c.String("proxy-https")
	proxyTLSCert := c.String("proxy-tls-cert")
	proxyTLSKey := c.String("proxy-tls-key")
	clientsHTTP := c.String("clients-http")
	clientsHTTPS := c.String("clients-https")
	clientsTLSCert := c.String("clients-tls-cert")
	clientsTLSKey := c.String("clients-tls-key")
	log.SetPrefix("[server] ")
	serv := server.NewServer()
	if proxyHTTPS != "" {
		go func() {
			wg.Add(1)
			tlsConfig := &server.TLSConfig{
				CertFile: proxyTLSCert,
				KeyFile:  proxyTLSKey,
			}
			// Start the HTTPS proxy server
			log.Printf("Starting HTTPS proxy server at %s", proxyHTTPS)
			if err := serv.StartServer(proxyHTTPS, false, tlsConfig); err != nil {
				log.Fatal(err)
			}
		}()
	}
	if proxyHTTP != "" {
		go func() {
			wg.Add(1)
			// Start the HTTP server
			log.Printf("Starting HTTP proxy server at %s", proxyHTTP)
			if err := serv.StartServer(proxyHTTP, false, nil); err != nil {
				log.Fatal(err)
			}
		}()
	}
	if clientsHTTPS != "" {
		go func() {
			wg.Add(1)
			tlsConfig := &server.TLSConfig{
				CertFile: clientsTLSCert,
				KeyFile:  clientsTLSKey,
			}
			// Start the HTTPS proxy server
			log.Printf("Starting HTTPS proxy server at %s", clientsHTTPS)
			if err := serv.StartServer(clientsHTTPS, true, tlsConfig); err != nil {
				log.Fatal(err)
			}
		}()
	}
	if clientsHTTP != "" {
		go func() {
			wg.Add(1)
			// Start the HTTP server
			log.Printf("Starting HTTP proxy server at %s", clientsHTTP)
			if err := serv.StartServer(clientsHTTP, true, nil); err != nil {
				log.Fatal(err)
			}
		}()
	}
	wg.Wait()
}

func main() {
	app := cli.NewApp()
	app.Name = "skyproxy"
	app.Version = "0.1.0"
	app.Usage = "Reverse tunnel HTTP proxy"
	app.Commands = globalCommands()
	app.Run(os.Args)
}
