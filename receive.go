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
	n.PrintSendReceiveMsg("receieve", msg.SendID, msg.MsgType, msg.ANum, msg.AVal)

	switch recvMsgType := msg.MsgType; recvMsgType {
	case PREPARE:
		n.recvPrepare(msg, conn)
	case PROMISE:
		n.recvPromise(msg)
	case ACCEPT:
		n.recvAccept(msg, conn)
	case ACK:
		n.recvAck(msg)
	case COMMIT:
		n.recvCommit(msg)
		//provide clarity to user that user input is still available
		fmt.Printf("Please enter a Command: ")
	default:
		fmt.Println("ERROR: The recieved message type is not valid")
	}
	//Not clever
	//fmt.Printf("\nPlease enter a Command:")
	return
}

func (n *Node) recvPrepare(msg message, conn net.Conn) {
	//If the slot is not equal to the proposed message slot, return the AccNum & AccVal of proposed slot
	//Maybe have a different message type

	//If the message is the current value, but the entry is empty, send the empty val back to signal this information
	//	Send a message back with an empty entry object

	//THE ISSUE IS HERE
	if msg.Slot == n.SlotCounter && msg.ANum > n.MaxPrepare {
		n.MaxPrepare = msg.ANum
		//send a new message as a response
		recvID := msg.SendID
		msgN := message{n.Id, PROMISE, n.AccNum, n.AccVal, n.SlotCounter}
		n.Send(conn, recvID, msgN)
	} else if msg.Slot < n.SlotCounter {
		//DO NOT MOVE TO FIRST ELSE STATEMENT WITH AN &&
		//IT WILL CAUSE AN OUT OF INDEX RANGE ERROR
		if msg.ANum >= n.Log[msg.Slot].MaxPrepare {
			//Recovery case
			n.Log[msg.Slot].MaxPrepare = msg.ANum
			recvID := msg.SendID
			msgN := message{n.Id, PROMISE, n.Log[msg.Slot].MaxPrepare, n.Log[msg.Slot], msg.Slot}
			n.Send(conn, recvID, msgN)
		} else {
			fmt.Println("MaxPrepare is less than or equal to proposed n (LOWER LOG SLOT). No reponse is being returned")
		}
	} else {
		//Possible optimization: Send something back saying that the request failed
		fmt.Printf("MaxPrepare %d is less than or equal to proposed n(%d). No reponse is being returned\n",
			n.MaxPrepare, msg.ANum)

		//fmt.Printf("msg.Slot: %d  : n.SlotCounter %d \n", msg.Slot, n.SlotCounter)
	}
	//TO DO: Add another if statement that won't respond to the request if it already accepted another one
	return
}

func (n *Node) recvPromise(msg message) {
	//The setting of the accVal is only included for telling the recovering site
	//	that there is nothing at the location it's trying to read.
	//Otherwise, we can assume, the AccVal will be set to the actual value after receiving accept

	//FOR RECOVERY
	var emptyVal entry
	if msg.AVal == emptyVal || msg.ANum == emptyPropossal {
		n.RecoverStop = true
	}
	n.AccVal = msg.AVal
	n.RecvAcceptedPromise++
	return
}

func (n *Node) recvAccept(msg message, conn net.Conn) {
	//If it's a different slot, change the MaxPrepare, AccNum, and AccVal of the specified log slot
	if (msg.ANum >= n.MaxPrepare && msg.Slot == n.SlotCounter) || msg.ANum == leaderPropossal {
		n.MaxPrepare = msg.ANum
		n.AccNum = msg.ANum
		n.AccVal = msg.AVal

		//send a new message as a response
		recvID := msg.SendID
		msg := message{n.Id, ACK, n.AccNum, n.AccVal, n.SlotCounter}
		n.Send(conn, recvID, msg)
	} else if msg.ANum >= n.Log[msg.Slot].MaxPrepare && msg.Slot < n.SlotCounter {
		//Recovery case
		n.Log[msg.Slot].MaxPrepare = msg.ANum
		n.Log[msg.Slot].AccNum = msg.ANum

		recvID := msg.SendID
		msgN := message{n.Id, ACK, n.Log[msg.Slot].MaxPrepare, n.Log[msg.Slot], msg.Slot}
		n.Send(conn, recvID, msgN)
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
	//Check for if the propossal msg is of different log slot (instead of this if statement)
	// If you have recieved a commit for a log slot you have already committed, update the AccNum & AccVal at log location
	//if msg.AVal == n.AccVal
	if msg.Slot == n.SlotCounter {
		//Update the node values
		n.CommitNodeUpdate()
		//Add the value to the physical and virtual log
		n.Log = append(n.Log, msg.AVal)
		n.writeLog()

		//If the entry is a block or unblock, update the dictionary
		if msg.AVal.Event == INSERT {
			n.Blocks[msg.AVal.User][msg.AVal.Follower] = true
		} else if msg.AVal.Event == DELETE {
			delete(n.Blocks[msg.AVal.User], msg.AVal.Follower)
		}
	}
	return
}
