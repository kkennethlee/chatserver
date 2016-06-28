package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

//User container user information
type User struct {
	Username string
	Output   chan Message
}

//Message contains who posted it, the message and the time of message
type Message struct {
	User *User
	Text string
	When string
}

//ChatServer contains behaviors of chat server
type ChatServer struct {
	//list of all users
	Users map[string]User

	//User has joined
	Join chan User
	//User has left
	Leave chan User

	//All the outputs of the users are input into the chatserver
	Input chan Message
}

func userEnter(cs *ChatServer, userJoined User) {
	cs.Input <- Message{
		User: &userJoined,
		Text: fmt.Sprintf("%s has joined\n", userJoined.Username),
		When: fmt.Sprintf("%s ", time.Now()),
	}
}

func userExit(cs *ChatServer, userLeft User) {
	cs.Input <- Message{
		User: &userLeft,
		Text: fmt.Sprintf("%s has left", userLeft.Username),
		When: fmt.Sprintf("%s ", time.Now()),
	}
}

func (chatServer *ChatServer) start() {
	for {
		select {

		case userJoined := <-chatServer.Join:

			chatServer.Users[userJoined.Username] = userJoined

			go userEnter(chatServer, userJoined)

		case userLeft := <-chatServer.Leave:

			delete(chatServer.Users, userLeft.Username)

			go userExit(chatServer, userLeft)

		case messages := <-chatServer.Input:
			for _, users := range chatServer.Users {
				users.Output <- messages
			}

		}
	}
}

//user
func handlerConnection(connection net.Conn, chatServer *ChatServer) {
	io.WriteString(connection, "Enter your username\n")

	scanner := bufio.NewScanner(connection)
	scanner.Scan()

	defer connection.Close()

	user := User{
		Username: scanner.Text(),
		Output:   make(chan Message),
	}

	chatServer.Join <- user
	defer func() {
		chatServer.Leave <- user
	}()

	//read
	go func() {
		scan := bufio.NewScanner(connection)
		for scan.Scan() {
			ln := scan.Text()
			time := time.Now()
			chatServer.Input <- Message{
				User: &user,
				Text: ln,
				When: time.String(),
			}
		}
	}()

	//Write all the outputs
	for message := range user.Output {
		io.WriteString(connection, message.When+" "+message.User.Username+": "+message.Text)
	}
}

func main() {

	// listening for request
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()

	//create chatroom
	chatRoom := &ChatServer{
		Users: make(map[string]User),
		Join:  make(chan User),
		Leave: make(chan User),
		Input: make(chan Message),
	}

	go chatRoom.start()

	for {
		connection, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		go handlerConnection(connection, chatRoom)
		//writing to connection
		//io.WriteString(connection, fmt.Sprint("Hello World\n", time.Now(), "\n"))

		//connection.Close()
	}
}
