package main

import (
	"encoding/json"
	"log"
	"net"
)

//Generate message for send
func (n *Node) BroadCast() {
	for i, ip := range n.IPtargets {
		if ok := n.Blocks[n.Id][i]; ok {
			log.Println("ID ", i, " is blocked, not sending to location")
			continue
		}
		conn, err := net.Dial("tcp", ip)
		if err != nil {
			log.Println("Failed to connect to ", ip, "  ", err)
			continue
		}
		n.Send(conn, i)
	}
	return
}

//Send the message the other ip targets
func (n *Node) Send(conn net.Conn, k int) {
	defer conn.Close()
	var msg message
	msg.Events = make([][]tweet, len(n.Log))
	msg.Ti = n.TimeArray
	msg.SendID = n.Id

	for i := range n.Log {
		for j := range n.Log[i] {
			if !n.hasRec(n.Log[i][j], k) {
				msg.Events[i] = append(msg.Events[i], n.Log[i][j])
			}
		}
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Println("failed to build message for ", k, "   ", err)
		return
	}

	check, err := conn.Write(bytes)
	if err != nil || check != len(bytes) {
		log.Println("Failed to send message to ", k, "  ", err)
		return
	}
	log.Println("Successfully sent message to ", k)
}
