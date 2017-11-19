package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

//Propose a new propossal number n
func (n *Node) Propose(ety entry) bool {
	//Idea: use accepted val in node to track the number of messages that have accepted the request

	//Send a prepare message to all acceptors
	//Wait until a response is recieved from a majority
	//  If no majority is recieved, timeout and exit the function
	exitBool := false
	timeout := time.After(1 * time.Second)
	var emptyEntry entry
	n.RecvAcceptedPromise = 0
	msg := message{n.Id, PREPARE, n.ProposalVal, emptyEntry}
	n.BroadCast(msg)
	fmt.Println("ACCEPTED:", n.RecvAcceptedPromise)
	for {
		//if timeout {
		//	return false
		//}
		select {
		case <-timeout:
			exitBool = true
			fmt.Println("Timeout")
			//break
			//default:
			//change to accept var
			//if n.RecvAccepted >= n.MajorityVal {
			//	fmt.Println("break")
			//	break
			//}
		}
		fmt.Println("ACCEPTED:", n.RecvAcceptedPromise)
		if n.RecvAcceptedPromise >= n.MajorityVal {
			//if the number is achieved, exit, goal complete
			return true
		} else if exitBool == true {
			//REMOVE THIS BLOCK ONCE COMMUNICATION WORKS
			//if timeout achieved or the propossal was a failure, try another value
			break
		} else {
			n.IncrementPropossalVal()
		}
		//CORNER CASE: Check to make sure there are enough acceptors alive to send back a promise
	}
	fmt.Println("yay! Out of the the Propose!")

	return true
}

//Generate message for send and possible recieve
func (n *Node) BroadCast(msg message) {
	//n.NodeMutex.Lock()
	//defer n.NodeMutex.Unlock()
	for i, ip := range n.IPtargets {
		fmt.Println(i)
		//fmt.Println(ip)

		//Don't bother sending it to another location if the current acceptor is at this process
		//	Remove this code once communication works, or don't
		if i == n.Id {
			if msg.MsgType == PREPARE {
				go n.recvPromise(msg)
			}
			continue
		}
		/*
			// Bad code that got us points taken off last time, bad
			if ok := n.Blocks[n.Id][i]; ok {
				log.Println("ID ", i, " is blocked, not sending to location")
				continue
			}
		*/
		/*
			if i != 0 {
				continue
			}
		*/

		go n.HandleSendAndRecieve(ip, i, msg)
	}
	return
}

func (n *Node) HandleSendAndRecieve(ip string, k int, msg message) {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		log.Println("Failed to connect to ", ip, "  ", err)
		return
	}
	defer conn.Close()
	n.Send(conn, k, msg)
	if msg.MsgType == PREPARE || msg.MsgType == ACCEPT {
		n.recieve(conn)
	}
	return
}

//Send the message the other ip targets
func (n *Node) Send(conn net.Conn, k int, msg message) {
	//defer conn.Close()

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
	return
}
