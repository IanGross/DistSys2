package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
)

func listen(serv *Node) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(serv.ListenPort))
	if err != nil {
		log.Fatalln("Failed to connect on port, shutting down ", err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go handleConn(conn, serv)
	}
}

func handleConn(conn net.Conn, serv *Node) {
	fmt.Println("Handling connection request...")
	defer conn.Close()

	//NOTE: The buffer is currently statically allocated
	// If the message becomes too large, it will cause errors
	//Future To Do: send byte size before sending the log
	recvBuf := make([]byte, 4096*10)
	_, err := conn.Read(recvBuf)
	recvBuf = bytes.Trim(recvBuf, "\x00")
	if err != nil {
		log.Println(err)
		return
	}
	var msg message
	err = json.Unmarshal(recvBuf, &msg)
	if err != nil {
		log.Println(err)
		return
	}

	//NOTE: UPDATE THE RECIEVE FUNCTION BEFORE UNCOMMENTING
	//serv.receive(msg)
	return
}
