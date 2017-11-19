package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

func (n *Node) recieve(conn net.Conn) {

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

	log.Println("Recieved message from ", msg.SendID)

	switch recvMsgType := msg.MsgType; recvMsgType {
	case PREPARE:
		fmt.Println("PREPARE RECIEVED")
		n.recvPrepare(msg, conn)
	case PROMISE:
		fmt.Println("PROMISE RECIEVED")
	case ACCEPT:
		fmt.Println("ACCEPT RECIEVED")
	case ACK:
		fmt.Println("ACK RECIEVED")
	case COMMIT:
		fmt.Println("COMMIT RECIEVED")
	default:
		fmt.Println("ERROR: The recieved message type is not valid")
	}

	//provide clarity to user that user input is still available
	fmt.Printf("Please enter a Command: ")
}

func (n *Node) recvPrepare(msg message, conn net.Conn) {
	if msg.ANum > n.MaxPrepare {
		n.MaxPrepare = msg.ANum

		//send a new message as a response
		recvID := msg.SendID
		msg := message{n.Id, PROMISE, n.AccNum, n.AccVal}
		n.Send(conn, recvID, msg)
	} else {
		fmt.Println("MaxPrepare is less than or equal to proposed n. No reponse is being returned")
	}
	//TO DO: Add another if statement that won't respond to the request if it already accepted another one
	return
}

func (n *Node) recvPromise(msg message) {
	if msg.ANum > n.MaxPrepare {
		n.MaxPrepare = msg.ANum
		n.RecvAcceptedPromise++
	}
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
	n.IncrementPropossalVal()
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