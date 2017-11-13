package main

import (
	"fmt"
	"log"
)

/*
The logic behind this functiion starts with inserting and deleting things from
the blocks dict, but where does it go from there
The bit about removing things from the record is important
*/

func (n *Node) UpdateDict(events [][]tweet) {
	for _, stuff := range events {
		for _, record := range stuff {
			if record.Event == INSERT {
				n.Blocks[record.User][record.Follower] = true
			} else if record.Event == DELETE {
				//delete(n.Blocks[record.User], record.Follower)
				n.Blocks[record.User][record.Follower] = false
			}
		}
	}
}

//Put messages into log
func (n *Node) UpdateLog(events [][]tweet) {
	for i, noderec := range events {
		for _, record := range noderec {
			n.Log[i] = append(n.Log[i], record)
		}
	}
	n.writeLog()
}

func (n *Node) receive(msg message) {
	// Locks occur at this level
	// because if they occur any lower
	// one thread might change the clocks before another adds to the logs.
	// which would be ... very bad

	n.NodeMutex.Lock()
	defer n.NodeMutex.Unlock()

	log.Println("Recieved message from ", msg.SendID)

	//Figure which events are actually new
	newEvent := make([][]tweet, len(n.Log))
	for i := range n.Log {
		for j := range msg.Events[i] {
			if !(n.hasRec(msg.Events[i][j], n.Id)) {
				newEvent[i] = append(newEvent[i], msg.Events[i][j])
			}
		}
	}

	//update dictonary
	n.UpdateDict(newEvent)

	//update the time array
	for k := range n.TimeArray[n.Id] {
		n.TimeArray[n.Id][k] = maxInt(n.TimeArray[n.Id][k], msg.Ti[msg.SendID][k])
	}
	for i := range n.TimeArray {
		for j := range n.TimeArray[i] {
			n.TimeArray[i][j] = maxInt(n.TimeArray[i][j], msg.Ti[i][j])
		}
	}

	//update local log
	// add new events to log to verify that they are not in the log
	n.UpdateLog(newEvent)

	n.CleanDict()

	n.writeDict()
	n.writeLog()
	n.writeTArray()

	//provide clarity to user that user input is still available
	fmt.Printf("Please enter a Command: ")
}
