package main

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type room struct {
	Name          string
	guests        map[*guest]struct{}
	HandleChannel chan ssh.Channel
}

type guest struct {
	channel ssh.Channel
	Name string
}

func newRoom() *room {
	return &room{
		Name:          "Chat room 1",
		guests:        make(map[*guest]struct{}),
		HandleChannel: make(chan ssh.Channel),
	}
}

func (room *room) run() {
	fmt.Println("Chat launched on port 2222\nChat logs will be displayed below\n-------------------------------")
	for {
		select {
		case c := <-room.HandleChannel:
			go func() {
				guest := welcome(room, c)
				for {
					term := terminal.NewTerminal(c, "\n")
					term.SetPrompt("\r\n\033[1;36m" + guest.Name + "\033[37m: ")
					r, err := term.ReadLine()
					if err != nil {
						fmt.Println(err)
					}
					term.SetPrompt("")
					guest.sendMsg(room, string(r))
				}
			}()
		}
	}
}

func (guest *guest) self(text string) {
	io.WriteString(guest.channel, text)
}

func (guest *guest) sendMsg(room *room, text string) {
	for chat := range room.guests {
		if guest == chat {
			continue
		}
		io.Writer.Write(chat.channel, []byte("\r\n\033[1;1;36m"+guest.Name+"\033[37m: "+text+"\r\n"))
		io.Writer.Write(chat.channel, []byte("\r\n"))
		fmt.Println("\n" + guest.Name + ":" + text)
	}
}

func handle(name string, c ssh.Channel) *guest {
	chat := guest{channel: c, Name: name}
	return &chat
}

func welcome(room *room, c ssh.Channel) *guest {
	var byte bytes.Buffer
	byte.WriteString("\033[1;36mWelcome guest\r\n\nRoom name: " + room.Name + "\r\n\nPlease enter your alias\033[37m: ")
	io.Copy(c, &byte)
	term := terminal.NewTerminal(c, "")
	r, err := term.ReadLine()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("New user logged in: " + r)
	chatname := r
	guest := handle(string(chatname), c)
	room.guests[guest] = struct{}{}
	return guest
}
