package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var NetworkType, SocketAddr string

var client = websocket.Dialer{
	NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial(NetworkType, SocketAddr)
	},
}

var cmdScan = &cli.Command{
	Name:  "scan",
	Usage: "trigger a scan on hypersonic instance (must keep hsctl running)",
	Action: func(ctx context.Context, c *cli.Command) error {
		conn, _, err := client.DialContext(ctx, "ws://localhost/library/scan", nil)
		if err != nil {
			return fmt.Errorf("dial admin socket: %w", err)
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					break
				}
				return fmt.Errorf("read message: %w", err)
			}
			logrus.Info(string(msg))
		}

		return nil
	},
}

var cmdCreateUser = &cli.Command{
	Name:  "create-user",
	Usage: "create a user on the hypersonic instance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "username",
			Aliases:  []string{"u"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"p"},
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		username := c.String("username")
		password := c.String("password")

		if password == "" {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(0)
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			fmt.Println()

			password = string(passwordBytes)
		}

		fmt.Println(username, password)

		return nil
	},
}

func main() {
	cmd := &cli.Command{
		Name:  "register",
		Usage: "create a new user",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "socket address of hypersonic admin server",
				Aliases:     []string{"a"},
				Value:       "/var/run/hypersonic/admin.sock",
				Destination: &SocketAddr,
			},
			&cli.StringFlag{
				Name:        "tcp",
				Usage:       "network type (tcp/unix)",
				Aliases:     []string{"n"},
				Value:       "unix",
				Destination: &NetworkType,
				Validator: func(s string) error {
					if s == "unix" || s == "tcp" {
						return nil
					}
					return fmt.Errorf("must be either tcp or unix")
				},
			},
		},
		Commands: []*cli.Command{cmdScan, cmdCreateUser},
	}

	cmd.Run(context.Background(), os.Args)
}
