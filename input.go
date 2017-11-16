package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

//TO DO: Call the propose, accept, etc.
// For simplicity, maybe combine the three events into one function, and pass the event type to distinguish between the three

//Prints all events stored in the log
func (localN *Node) PrintLog() {
	//fmt.Println(localN.Log)
	for i := 0; i < len(localN.Log); i++ {
		fmt.Printf(time.Time.String(localN.Log[i].Clock) + " - ")
		fmt.Printf("User " + strconv.Itoa(localN.Log[i].User) + ", ")
		if localN.Log[i].Event == 0 {
			fmt.Printf("TWEET: " + localN.Log[i].Message)
		} else if localN.Log[i].Event == 1 {
			fmt.Printf("BLOCK: Follower " + strconv.Itoa(localN.Log[i].Follower))
		} else if localN.Log[i].Event == 2 {
			fmt.Printf("UNBLOCK: Follower " + strconv.Itoa(localN.Log[i].Follower))
		}
		fmt.Println("")
	}
}

//Prints all events stored in the log (for pu)
func (localN *Node) PrintDictionary() {
	//fmt.Println(localN.Blocks)
	for k, v := range localN.Blocks {
		fmt.Println("Dictionary at site", k, "-", len(v), "site(s) blocked")
		for kVal, vVal := range localN.Blocks[k] {
			fmt.Println("- Site", kVal, ", ", vVal)
		}
	}
}

func reverse(logArray []entry) []entry {
	for i, j := 0, len(logArray)-1; i < j; i, j = i+1, j-1 {
		logArray[i], logArray[j] = logArray[j], logArray[i]
	}
	return logArray
}

func OrganizeEntries(logContent []entry) []entry {
	var combineList []entry
	for i := 0; i < len(logContent); i++ {
		if logContent[i].Event != 0 {
			continue
		}
		if len(combineList) == 0 {
			combineList = append(combineList, logContent[i])
		} else {
			for k := 0; k < len(combineList); k++ {
				if logContent[i].Clock.Before(combineList[k].Clock) {
					var combineListBefore []entry
					combineListCopy := append([]entry(nil), combineList...)
					combineListBefore = append(combineList[:k], logContent[i])
					combineList = append(combineListBefore, combineListCopy[k:]...)
					break
				} else if k+1 == len(combineList) {
					combineList = append(combineList, logContent[i])
					break
				}
			}
		}
	}
	return combineList
}

func (localN *Node) ViewTweets() {
	fmt.Println("Current events in log:")
	organizedLog := OrganizeEntries(localN.Log)
	logReverse := reverse(organizedLog)
	for i := 0; i < len(logReverse); i++ {
		if logReverse[i].Event == 0 && localN.Blocks[localN.Id][logReverse[i].User] == false {
			fmt.Printf(time.Time.String(logReverse[i].Clock) + " - ")
			fmt.Printf("User " + strconv.Itoa(logReverse[i].User) + ": ")
			fmt.Printf(logReverse[i].Message)
			fmt.Println("")
		}
	}
}

func (localN *Node) TweetEvent(message string) {
	//Update the counter
	localN.NodeMutex.Lock()
	defer localN.NodeMutex.Unlock()
	ety := entry{message, localN.Id, localN.Id, time.Now().UTC(), -500000, 0}
	//Send Execution (Put in a new file)
	//	Call: Propose
	//	Wait for a Promise from the Majority of Users
	//		If False (competing propossal or no majority), stop execution and tell user that proposal has failed (DON'T ADD TO LOG)
	//		If True, send accept to all, wait for ack message, then send commit
	//			On commit, save the entry to the log

	retVal := localN.Propose(ety)
	if retVal == true {
		fmt.Println("Propossal was successful")
	} else if retVal == false {
		fmt.Println("Error: Propossal was Unsuccessful")
	}
	//update the tweet in memory
	//localN.Log = append(localN.Log, ety)
	//update the tweet in the physical log
	//localN.writeLog()
}

func (localN *Node) InvalidBlock(username string, blockType int) bool {
	userID, err := strconv.Atoi(username)
	if err != nil {
		return true
	}
	//SAFTEY CHECKS:
	// - User calls block/unblock on another user that exists (and string is a number)
	//		- Assume that the id is always from 0 to len-1
	if userID < 0 || userID > len(localN.IPtargets) {
		return true
	}
	// - User calls block on themself
	if localN.Id == userID {
		return true
	}
	// - User calls block on a user that is already blocked
	// - User calls unblock on a user that is not in the dictionary
	if ok := localN.Blocks[localN.Id][userID]; ok {
		if blockType == 1 {
			return true
		}
	} else if blockType == 2 {
		return true
	}

	return false
}

func (localN *Node) BlockUser(username string) {
	if localN.InvalidBlock(username, 1) == true {
		log.Println("Invalid Block Call")
		return
	}
	localN.NodeMutex.Lock()
	defer localN.NodeMutex.Unlock()
	//userID, _ := strconv.Atoi(username)
	//etyBlock := entry{"", localN.Id, userID, time.Now().UTC(), -5000000, 1}
	//Send Execution

	//localN.Log[localN.Id] = append(localN.Log[localN.Id], etyBlock)
	//localN.Blocks[localN.Id][userID] = true
	//localN.writeLog()
	//localN.addToDict()
}

func (localN *Node) UnblockUser(username string) {
	if localN.InvalidBlock(username, 2) == true {
		log.Println("Invalid Block Call")
		return
	}
	localN.NodeMutex.Lock()
	defer localN.NodeMutex.Unlock()
	//userID, _ := strconv.Atoi(username)
	//etyUnblock := entry{"", localN.Id, userID, time.Now().UTC(), -500000, 2}
	//Send Execution

	//localN.Log[localN.Id] = append(localN.Log[localN.Id], twtUnblock)
	//delete(localN.Blocks[localN.Id], userID)
	//localN.Blocks[localN.Id][userID] = false
	//localN.writeLog()
	//localN.addToDict()
}

func InputHandler(local *Node) {
	reader := bufio.NewReader(os.Stdin)
	for true {
		fmt.Printf("Please enter a Command: ")
		inputTmp, _ := reader.ReadString('\n')
		input := strings.Replace(inputTmp, "\r", "", -1)

		if i := strings.Index(input, "tweet"); i == 0 {
			message := input[6 : len(input)-1]
			fmt.Println("Tweet Called")
			local.TweetEvent(message)
		} else if i := strings.Index(input, "view"); i == 0 {
			fmt.Printf("View called\n")
			local.ViewTweets()
		} else if i := strings.Index(input, "block"); i == 0 {
			username := input[6:7]
			fmt.Printf("Block called on %s\n", username)
			local.BlockUser(username)
		} else if i := strings.Index(input, "unblock"); i == 0 {
			username := input[8:9]
			fmt.Printf("Unblock called on %s\n", username)
			local.UnblockUser(username)
		} else if i := strings.Index(input, "print log"); i == 0 {
			fmt.Printf("Print Log called\n")
			local.PrintLog()
		} else if i := strings.Index(input, "print dict"); i == 0 {
			fmt.Printf("Print Dictionary called\n")
			local.PrintDictionary()
		} else if i := strings.Index(input, "exit"); i == 0 {
			fmt.Printf("Exit called, exiting...\n")
			break
		} else {
			fmt.Printf("Command not recognized\n")
		}
	}
}
