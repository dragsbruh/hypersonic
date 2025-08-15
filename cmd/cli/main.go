package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var Network = "unix"
var Address = "/var/run/hypersonic/admin.sock"

var client = http.Client{
	Transport: &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial(Network, Address)
		},
	},
}
var wsClient = websocket.Dialer{
	NetDial: func(network, addr string) (net.Conn, error) {
		return net.Dial(Network, Address)
	},
}

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		logrus.Fatalf("expected command")
	}

	command := os.Args[1]

	switch command {
	case "scan":
		scan(ctx)
	case "create-user":
		createUser()
	}
}

func scan(ctx context.Context) {
	conn, _, err := wsClient.DialContext(ctx, "ws://localhost/library/scan", nil)
	if err != nil {
		logrus.Fatalf("failed to dial unix socket: %v", err)
	}
	defer conn.Close()

	for {
		var json = map[string]any{}
		if err := conn.ReadJSON(&json); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				break
			}
			logrus.Fatalf("error reading message: %v", err)
		}

		logrus.Println(json["Status"], json["Count"], json["FileHash"], json["Path"])
	}
}

func createUser() {
	var username, password string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Username").Value(&username),
			huh.NewInput().Title("Password").Value(&password),
		),
	).WithTheme(huh.ThemeBase16()).Run()
	if err != nil {
		logrus.Fatalf("error running form: %v", err)
	}

	buf := bytes.NewBuffer([]byte{})

	json.NewEncoder(buf).Encode(map[string]any{
		"Username": username,
		"Password": password,
	})

	resp, err := client.Post("http://unix/users/create", "application/json", buf)
	if err != nil {
		logrus.Fatalf("error sending request: %v", err)
	}

	var data struct {
		Status string `json:"Status"`
		Error  string `json:"Error"`
		UserID string `json:"UserID"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	if data.Status != "Ok" {
		logrus.Fatalf(data.Status, data.Error)
	}
	logrus.Infof("Created user with user id %s", data.UserID)
}
