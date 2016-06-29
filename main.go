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

func userEnter(cs *ChatServer, userJoined User, bot *User) {
	cs.Input <- Message{
		//User: &userJoined,
		User: bot,
		Text: fmt.Sprintf("%s has joined\n", userJoined.Username),
		When: fmt.Sprintf("%s ", time.Now()),
	}
}

func userExit(cs *ChatServer, userLeft User, bot *User) {
	cs.Input <- Message{
		//User: &userLeft,
		User: bot,
		Text: fmt.Sprintf("%s has left", userLeft.Username),
		When: fmt.Sprintf("%s ", time.Now()),
	}
}

func (chatServer *ChatServer) start() {

	//create a special user that announces the ins and outs of the
	chatBot := &User{
		Username: "ChatBot",
		Output:   make(chan Message),
	}

	for {
		select {

		case userJoined := <-chatServer.Join:

			chatServer.Users[userJoined.Username] = userJoined

			go userEnter(chatServer, userJoined, chatBot)

		case userLeft := <-chatServer.Leave:

			delete(chatServer.Users, userLeft.Username)

			go userExit(chatServer, userLeft, chatBot)

		case messages := <-chatServer.Input:
			for _, users := range chatServer.Users {
				users.Output <- messages
			}

		}
	}
}

//handlerConnection is used when user types "telnet localhost 9000" to establish connection
func handlerConnection(connection net.Conn, chatServer *ChatServer) {
	io.WriteString(connection, "Enter your username: ")

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

	//user.Output is a channel of Message
	for message := range user.Output {
		io.WriteString(connection, message.When+" "+message.User.Username+": "+message.Text+"\n")

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
