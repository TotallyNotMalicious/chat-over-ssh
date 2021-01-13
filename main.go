package main

import (
	"fmt"
	"io/ioutil"
	"net"

	"golang.org/x/crypto/ssh"
)

func handler(conn net.Conn, config *ssh.ServerConfig, chatRoom *room) {
	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		fmt.Println("Bad handshake with most recent connection attempt")
		return
	}
	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "UnknownChannelType")
			return
		}
		channel, requests, err := newChan.Accept()
		if err != nil {
			fmt.Println("Could not accept channel")
			return
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				switch req.Type {
				case "pty-req":
					req.Reply(true, nil)
					continue
				case "shell":
					req.Reply(true, nil)
					continue
				}
				req.Reply(false, nil)
			}
		}(requests)

		chatRoom.HandleChannel <- channel
	}
}

func main() {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "chat" && string(pass) == "server" { // change to something more secure if you will actually be using this
				return nil, nil
			}
			return nil, fmt.Errorf("Failed login attempt: %q", c.User())
		},
	}
	// ssh-keygen -t rsa
	privKey, err := ioutil.ReadFile("./id_rsa")
	if err != nil {
		fmt.Println("Failed to load private key")
		return
	}

	private, err := ssh.ParsePrivateKey(privKey)
	if err != nil {
		fmt.Println("Failed to parse private key")
		return
	}

	config.AddHostKey(private)

	chatRoom := newRoom()
	go chatRoom.run()
	listener, err := net.Listen("tcp", fmt.Sprintf(":2222"))
	if err != nil {
		fmt.Println("Error: %s", err)
		return
	}
	for {
		newConnection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error: %s", err)
			return
		}
		go handler(newConnection, config, chatRoom)
	}
}
