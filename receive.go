package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

func (n *Node) receive(conn net.Conn) {

	recvBuf := make([]byte, 4096*10)
	_, err := conn.Read(recvBuf)
	recvBuf = bytes.Trim(recvBuf, "\x00")
	if err != nil {
		fmt.Println("Receive Function: No message was found")
		//log.Println(err)
		return
	}
	var msg message
	err = json.Unmarshal(recvBuf, &msg)
	if err != nil {
		log.Println(err)
		return
	}

	//log.Println("Received message from ", msg.SendID)

	switch recvMsgType := msg.MsgType; recvMsgType {
	case PREPARE:
		fmt.Printf("Received message from %v - Propose(%v)\n", msg.SendID, msg.ANum)
		n.recvPrepare(msg, conn)
	case PROMISE:
		fmt.Printf("Received message from %v - Promise(%v,%v)\n", msg.SendID, msg.ANum, msg.AVal)
		n.recvPromise(msg)
	case ACCEPT:
		fmt.Printf("Received message from %v - Accept(%v,%v)\n", msg.SendID, msg.ANum, msg.AVal)
		n.recvAccept(msg, conn)
	case ACK:
		fmt.Printf("Received message from %v - Ack(%v,%v)\n", msg.SendID, msg.ANum, msg.AVal)
		n.recvAck(msg)
	case COMMIT:
		fmt.Printf("Received message from %v - Commit(%v)\n", msg.SendID, msg.AVal)
		n.recvCommit(msg)
	default:
		fmt.Println("ERROR: The recieved message type is not valid")
	}

	//provide clarity to user that user input is still available
	//fmt.Printf("Please enter a Command: ")
}

func (n *Node) recvPrepare(msg message, conn net.Conn) {
	if msg.ANum > n.MaxPrepare {
		n.MaxPrepare = msg.ANum

		//send a new message as a response
		recvID := msg.SendID
		msg := message{n.Id, PROMISE, n.AccNum, n.AccVal}
		n.Send(conn, recvID, msg)
	} else {
		//Possible optimization: Send something back saying that the request failed
		fmt.Println("MaxPrepare is less than or equal to proposed n. No reponse is being returned")
	}
	//TO DO: Add another if statement that won't respond to the request if it already accepted another one
	return
}

func (n *Node) recvPromise(msg message) {
	n.RecvAcceptedPromise++
	return
}

func (n *Node) recvAccept(msg message, conn net.Conn) {
	if msg.ANum >= n.MaxPrepare {
		n.MaxPrepare = msg.ANum
		n.AccNum = msg.ANum
		n.AccVal = msg.AVal

		//send a new message as a response
		recvID := msg.SendID
		msg := message{n.Id, ACK, n.AccNum, n.AccVal}
		n.Send(conn, recvID, msg)
	} else {
		fmt.Println("MaxPrepare is less than accept value of n. No reponse is being returned")
	}
	return
}

func (n *Node) recvAck(msg message) {
	n.RecvAcceptedAck++
	return
}

func (n *Node) recvCommit(msg message) {
	//Update the node values
	n.SlotCounter++
	//n.IncrementPropossalVal()
	//Add the value to the physical and virtual log
	n.Log = append(n.Log, msg.AVal)
	n.writeLog()

	//If the entry is a block or unblock, update the dictionary
	if msg.AVal.Event == INSERT {
		n.Blocks[msg.AVal.User][msg.AVal.Follower] = true
	} else if msg.AVal.Event == DELETE {
		delete(n.Blocks[msg.AVal.User], msg.AVal.Follower)
		//n.Blocks[msg.AVal.User][msg.AVal.Follower] = false
	}
	return
}
