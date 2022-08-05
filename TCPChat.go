package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

var (
	conns    = map[string]net.Conn{}
	messages string
	logo     []byte
	mutex    sync.Mutex
)

func main() {
	port := "8989"
	host := flag.String("h", "localhost", "Host")
	flag.Parse()
	if len(flag.Args()) > 1 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}
	if len(flag.Args()) > 0 {
		port = flag.Arg(0)
	}
	l, err := ioutil.ReadFile("logo.txt")
	if err != nil {
		log.Fatal("couldn't read the logotype")
	}
	logo = l
	startServer(host, port)
}

func startServer(host *string, port string) {
	addr := fmt.Sprintf("%s:%s", *host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Listening on the port :%s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection from client: %s", err)
		} else {
			go chat(conn)
		}
	}
}

func chat(guest net.Conn) {
	if len(conns) == 10 {
		guest.Write([]byte("sorry, the chat members can't exceed 10 people, try again later"))
		guest.Close()
		return
	}
	guest.Write(logo)
	var name, msg string
	guest.Write([]byte("[ENTER YOUR NAME]:"))
	scanner := bufio.NewScanner(guest)

	for scanner.Scan() {
		name = strings.TrimSpace(scanner.Text())
		if len(name) == 0 {
			guest.Write([]byte("[PLEASE, ENTER YOUR CORRECT NAME]:"))
		} else if _, ok := conns[name]; ok {
			guest.Write([]byte("[PLEASE, ENTER A DIFFERENT NAME]:"))
		} else {
			conns[name] = guest
			break
		}
	}

	guest.Write([]byte(messages))
	send(name, fmt.Sprintf("%s has joined our chat...\n", name), true)

	for scanner.Scan() {
		text := strings.Trim(scanner.Text(), "\n\t\r")
		if len(text) != 0 {
			mutex.Lock()
			msg = fmt.Sprintf("[%s][%s]:", getTime(), name) + text + "\n"
			send(name, msg, false)
			mutex.Unlock()
		}
	}
	send(name, fmt.Sprintf("%s has left our chat...\n", name), true)
	delete(conns, name)
	guest.Close()
}

func send(name, msg string, inform bool) {
	for guest := range conns {
		if guest != name {
			conns[guest].Write([]byte("\n" + msg))
		}
		conns[guest].Write([]byte(fmt.Sprintf("[%s][%s]:", getTime(), guest)))
	}
	if !inform {
		messages += msg
	}
}

func getTime() string {
	return time.Now().Format(TimeFormat)
}
