package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

//Propose a new propossal number n
func (n *Node) ProposePhase(ety entry) bool {
	//exitBool := false
	//timeout := time.After(10 * time.Second)
	for {
		n.RecvAcceptedPromise = 0
		n.CountSiteFailures = 0
		msg := message{n.Id, PREPARE, n.ProposalVal, ety}
		n.BroadCast(msg)
		/*
			fmt.Println("Waiting on responses...")
			select {
			//Wait to hear back from everyone
			case <-timeout:
				exitBool = true
				fmt.Println("Timeout")
			default:
				//This section may not be necessary
				if n.RecvAcceptedPromise+n.RecvNotAcceptedPromise+n.CountSiteFailures == len(n.IPtargets) {
					break
				}
			}
		*/
		fmt.Println("Number of Responses Recieved:", n.RecvAcceptedPromise)
		if n.RecvAcceptedPromise >= n.MajorityVal {
			//if the number is achieved, exit, goal complete
			return true
		} else if n.CountSiteFailures >= n.MajorityVal {
			fmt.Println("Majority of sites have failed, Event Propossal impossible")
			return false
			/*} else if exitBool == true {
			//REMOVE THIS BLOCK ONCE COMMUNICATION WORKS
			//if timeout achieved or the propossal was a failure, try another value
			break
			*/
		} else {
			n.IncrementPropossalVal()
		}
	}
}

//Propose a new propossal number n
func (n *Node) AcceptPhase(ety entry) bool {
	n.RecvAcceptedAck = 0
	n.CountSiteFailures = 0
	msg := message{n.Id, ACCEPT, n.ProposalVal, ety}
	n.BroadCast(msg)
	fmt.Println("Number of Responses Recieved:", n.RecvAcceptedAck)
	//If receive ack from a majority, send commit(v)
	if n.RecvAcceptedAck >= n.MajorityVal {
		fmt.Println("Received ack from a majority of sites, sending commit")
		msg := message{n.Id, COMMIT, n.ProposalVal, ety}
		n.BroadCast(msg)
		n.IncrementPropossalVal()
		return true
	}
	return false
}

//Generate message for send and possible recieve
func (n *Node) BroadCast(msg message) {
	//n.NodeMutex.Lock()
	//defer n.NodeMutex.Unlock()
	for i, ip := range n.IPtargets {
		//fmt.Println(i, ip)

		//Don't bother sending it to another location if the current acceptor is at this process (Remove this code, or don't)
		/*
			if i == n.Id {
				if msg.MsgType == PREPARE {
					go n.recvPromise(msg)
				}
				continue
			}*/
		/*
			// Dictionary block code (to be re-added once paxos implementation is complete, or not, I don't know yet)
			if ok := n.Blocks[n.Id][i]; ok {
				log.Println("ID ", i, " is blocked, not sending to location")
				continue
			}
		*/
		//Possible improvement: create new go thread for each one (may lead to errors)
		n.HandleSendAndRecieve(ip, i, msg)
	}
	return
}

func (n *Node) HandleSendAndRecieve(ip string, k int, msg message) {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		//log.Println("Failed to connect to ", ip, "  ", err)
		log.Println("Failed to connect to ", ip, "  -   Acceptor is not alive at this location")
		n.CountSiteFailures++
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
	//log.Println("Successfully sent message to ", k)
	return
}
