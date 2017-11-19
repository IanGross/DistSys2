package main

import (
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
	//n.NodeMutex.Lock()
	//defer n.NodeMutex.Unlock()
	fmt.Println("Handling connection request...")
	defer conn.Close()
	serv.recieve(conn)
	//recieve may also send a response message, depending on the type of message
	return
}
