package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit
var store = make(map[string]string)
var expirations = make(map[string]int64)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// O que fazer quando isPrefix for True?
		line, err := reader.ReadString('\n')
		if len(line) == 0 {
			continue
		}

		if err != nil {
			fmt.Println("Error reading command")
		}

		if string(line[0]) != "*" {
			fmt.Println("Invalid command. Must be array of bulk strings")
		}

		inputArrLen, _ := strconv.Atoi(strings.TrimSuffix(string(line[1:]), "\r\n"))

		command := make([]string, inputArrLen)
		for i := range inputArrLen {
			bulkStringDesc, _, _ := reader.ReadLine()
			if string(bulkStringDesc[0]) != "$" {
				fmt.Println("Invalid command. Must be array of bulk strings")
			}

			bulkStringVal, _, _ := reader.ReadLine()
			command[i] = string(bulkStringVal)
		}

		switch command[0] {
		case "ECHO":
			answer := fmt.Sprintf("$%d\r\n%v\r\n", len(command[1]), command[1])
			conn.Write([]byte(answer))
		case "SET":
			if len(command) > 3 && command[3] == "PX" {
				expirationMilis, _ := strconv.Atoi(command[4])
				expirations[command[1]] = time.Now().UnixMilli() + int64(expirationMilis)
			}

			store[command[1]] = command[2]
			conn.Write([]byte("+OK\r\n"))
		case "GET":
			if val, ok := store[command[1]]; ok {
				if exp, ok := expirations[command[1]]; !ok || exp > time.Now().UnixMilli() {
					conn.Write(fmt.Appendf(nil, "$%d\r\n%v\r\n", len(val), val))
					break
				}
			}
			conn.Write([]byte("$-1\r\n"))
		default:
			conn.Write([]byte("+PONG\r\n"))
		}
	}
}
