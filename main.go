package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aof.Close()

	aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handler(args)
	})

	api := NewAPI(aof)
	go api.Start()

	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	fmt.Println("Listening on port :6379")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}

		go handleConnection(conn, aof)
	}
}

func handleConnection(conn net.Conn, aof *Aof) {
	defer conn.Close()

	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			return
		}

		if value.typ != "array" {
			continue
		}

		if len(value.array) == 0 {
			continue
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		writer := NewWriter(conn)

		handler, ok := Handlers[command]
		if !ok {
			writer.Write(Value{typ: "string", str: ""})
			continue
		}

		switch command {
		case "SET", "HSET", "HDEL", "DEL", "INCR", "DECR", "INCRBY", "DECRBY", "APPEND", "LPOP", "RPOP", "LPUSH", "RPUSH":
			aof.Write(value)
		}
		result := handler(args)
		writer.Write(result)
	}
}
