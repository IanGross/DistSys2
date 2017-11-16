package main

import (
	"encoding/json"
	"log"
	"net"
)

//Propose a new propossal number n
func (n *Node) Propose(ety entry) bool {
	return false
}

//Generate message for send
func (n *Node) BroadCast(msgTypeVar int, aNumVar int, aValVar entry) {
	for i, ip := range n.IPtargets {
		/*
			// Bad code that got us points taken off last time, bad
			if ok := n.Blocks[n.Id][i]; ok {
				log.Println("ID ", i, " is blocked, not sending to location")
				continue
			}
		*/
		conn, err := net.Dial("tcp", ip)
		if err != nil {
			log.Println("Failed to connect to ", ip, "  ", err)
			continue
		}
		var msg message
		msg.SendID = n.Id
		msg.MsgType = msgTypeVar
		msg.ANum = aNumVar
		msg.AVal = aValVar

		n.Send(conn, i, msg)
	}
	return
}

//Send the message the other ip targets
func (n *Node) Send(conn net.Conn, k int, msg message) {
	defer conn.Close()

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
