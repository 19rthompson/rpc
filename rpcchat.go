package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	MsgRegister = iota
	MsgList
	MsgCheckMessages
	MsgTell
	MsgSay
	MsgQuit
	MsgShutdown
)

var mutex sync.Mutex
var messages map[string][]string
var shutdown chan struct{}

func main() {
	log.SetFlags(log.Ltime)

	var listenAddress string
	var serverAddress string
	var username string

	switch len(os.Args) {
	case 2:
		listenAddress = net.JoinHostPort("", os.Args[1])
	case 3:
		serverAddress = os.Args[1]
		if strings.HasPrefix(serverAddress, ":") {
			serverAddress = "localhost" + serverAddress
		}
		username = strings.TrimSpace(os.Args[2])
		if username == "" {
			log.Fatal("empty user name")
		}
	default:
		log.Fatalf("Usage: %s <port>   OR   %s <server> <user>",
			os.Args[0], os.Args[0])
	}

	if len(listenAddress) > 0 {
		server(listenAddress)
	} else {
		fmt.Println("client function not implemented yet")
		// client(serverAddress, username)
	}
}

func server(listenAddress string) {
	shutdown = make(chan struct{})
	messages = make(map[string][]string)

	// set up network listen and accept loop here
	// to receive RPC requests and dispatch each
	// in its own goroutine

	// wait for a shutdown request
	<-shutdown
	time.Sleep(100 * time.Millisecond)
}

func dispatch(conn net.Conn) {
	// handle a single incomming request:
	len := make([]byte, 2)
	// 1. Read the length (uint16)
	_, err := conn.Read(len)
	if err != nil {
		fmt.Printf("Read error 1 - %s\n", err)
	}
	var length int = int(binary.BigEndian.Uint32(len))

	// 2. Read the entire message into a []byte
	message := make([]byte, length)
	_, err3 := conn.Read(message)
	if err3 != nil {
		fmt.Printf("Read error 3 - %s\n", err)

	}
	messageType, message, err2 := ReadUint16(message)
	if err2 != nil {
		fmt.Printf("Read error 2 - %s\n", err)
	}
	messageString := string(message)
	switch messageType {
	case MsgRegister:
		err := Register(messageString[2:])
		if err != nil {
			fmt.Println("Name error!", err)
		}
	case MsgList:
		users := List()                        // returns a slice of users []users
		foo := make([]byte, 2)                 // makes empty slice of bytes, length 2 []byte
		pickle := WriteStringSlice(foo, users) // returns a slice of bytes []byte
		_, err := conn.Write(pickle)           // returns integer
		if err != nil {
			fmt.Println("Listing error!", err)
		}
	case MsgTell:
		user := strings.Split(messageString, " ")[0]
		target := strings.Split(messageString, " ")[1]
		message := strings.TrimLeft(messageString, user+" "+target+" ")
		Tell(user, target, message)
	case MsgSay:
		user := strings.Split(messageString, " ")[0]
		message := strings.TrimLeft(messageString, user+" ")
		Say(user, message)
	case MsgShutdown:
		Shutdown()
	default:
		fmt.Println("You need help!")
	}
	// 3. From the message, parse the message type (uint16)

	// 4. Call the appropriate server stub, giving it the
	//    remainder of the request []byte and collecting
	//    the response []byte
	// 5. Write the message length (uint16)
	// 6. Write the message []byte
	// 7. Close the connection
	//
	// On any error, be sure to close the connection, log a
	// message, and return (a request error should not kill
	// the entire server)
}

func Register(user string) error {
	if len(user) < 1 || len(user) > 20 {
		return fmt.Errorf("Register: user must be between 1 and 20 letters")
	}
	for _, r := range user {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("Register: user must only contain letters and digits")
		}
	}
	mutex.Lock()
	defer mutex.Unlock()

	msg := fmt.Sprintf("*** %s has logged in", user)
	log.Printf(msg)
	for target, queue := range messages {
		messages[target] = append(queue, msg)
	}
	messages[user] = nil

	return nil
}

func List() []string {
	mutex.Lock()
	defer mutex.Unlock()

	var users []string
	for target := range messages {
		users = append(users, target)
	}
	sort.Strings(users)

	return users
}

func CheckMessages(user string) []string {
	mutex.Lock()
	defer mutex.Unlock()

	if queue, present := messages[user]; present {
		messages[user] = nil
		return queue
	} else {
		return []string{"*** You are not logged in, " + user}
	}
}

func Tell(user, target, message string) {
	mutex.Lock()
	defer mutex.Unlock()

	msg := fmt.Sprintf("%s tells you %s", user, message)
	if queue, present := messages[target]; present {
		messages[target] = append(queue, msg)
	} else if queue, present := messages[user]; present {
		messages[user] = append(queue, "*** No such user: "+target)
	}
}

func Say(user, message string) {
	mutex.Lock()
	defer mutex.Unlock()

	msg := fmt.Sprintf("%s says %s", user, message)
	for target, queue := range messages {
		messages[target] = append(queue, msg)
	}
}

func Quit(user string) {
	mutex.Lock()
	defer mutex.Unlock()

	msg := fmt.Sprintf("*** %s has logged out", user)
	log.Print(msg)
	for target, queue := range messages {
		messages[target] = append(queue, msg)
	}
	delete(messages, user)
}

func Shutdown() {
	shutdown <- struct{}{}
}

func WriteUint16(buf []byte, n uint16) []byte {
	return append(buf, byte(n>>8), byte(n))
}

func ReadUint16(buf []byte) (uint16, []byte, error) {
	if len(buf) < 2 {
		return 0, buf, fmt.Errorf("buffer too short")
	}
	value := uint16(buf[0])<<8 + uint16(buf[1])
	return value, buf[2:], nil
}

func WriteString(buf []byte, s string) []byte {
	buf = WriteUint16(buf, uint16(len(s)))
	return append(buf, s...)
}

func ReadString(buf []byte) (string, []byte, error) {
	length, rest, err := ReadUint16(buf)
	if err != nil {
		return "", rest, err
	}
	if len(rest) < int(length) {
		return "", rest, fmt.Errorf("buffer too short for string")
	}
	return string(rest[:length]), rest[length:], nil
}

func WriteStringSlice(buf []byte, s []string) []byte {
	buf = WriteUint16(buf, uint16(len(s))) // Write the slice length
	for _, str := range s {
		buf = WriteString(buf, str) // Write each string
	}
	return buf
}

func ReadStringSlice(buf []byte) ([]string, []byte, error) {
	length, rest, err := ReadUint16(buf) // Slice length
	if err != nil {
		return nil, rest, err
	}
	var strs []string
	for i := 0; i < int(length); i++ {
		var str string
		str, rest, err = ReadString(rest)
		if err != nil {
			return strs, rest, err
		}
		strs = append(strs, str)
	}
	return strs, rest, nil
}
