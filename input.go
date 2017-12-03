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

//PrintLog Prints all events stored in the log
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

//PrintDictionary Prints all events stored in the log (for pu)
func (localN *Node) PrintDictionary() {
	//fmt.Println(localN.Blocks)
	for k, v := range localN.Blocks {
		fmt.Println("Dictionary at site", k, "-", len(v), "site(s) blocked")
		for kval, vval := range localN.Blocks[k] {
			fmt.Println("- Site", kval, ", ", vval)
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
		//if logContent[i].Event != 0 {
		//	continue
		//}
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
	localN.UpdateTimeline()
	for _, tweet := range localN.Timeline {
		//why so many Itoas? fmt.Printf("%i", somevalue ) would work
		fmt.Printf(time.Time.String(tweet.Clock) + " - ")
		fmt.Printf("Propossal value %d, ", tweet.AccNum)
		fmt.Printf("Slot %d - ", tweet.SlotNumber)
		fmt.Printf("User %d: ", tweet.User)
		fmt.Printf(tweet.Message)
		fmt.Println("")
	}
	/*organizedLog := OrganizeEntries(localN.Log)
	logReverse := reverse(organizedLog)
	for i := 0; i < len(logReverse); i++ {
		if logReverse[i].Event == TWEET && !localN.checkBlock(logReverse[i].User, localN.Id) {
			//why so many Itoas? fmt.Printf("%i", somevalue ) would work
			fmt.Printf(time.Time.String(logReverse[i].Clock) + " - ")
			fmt.Printf("Propossal value %d, ", logReverse[i].AccNum)
			fmt.Printf("Slot %d - ", logReverse[i].SlotNumber)
			fmt.Printf("User %d: ", logReverse[i].User)
			fmt.Printf(logReverse[i].Message)
			fmt.Println("")
		}
	}*/
}

// Do not print tweet if the tweeter has blocked me
// an ONLY if the tweeter has blocked me
func (localN *Node) checkBlock(tweeter int, self int) bool {
	val, ok := localN.Blocks[tweeter][self]
	if ok {
		return val
	}
	return false
}

func (localN *Node) ProposeHandler(ety entry, slotPropose int) {
	var emptyEty entry
	if localN.AmLeader() {
		localN.LeaderPropseHandler(ety, slotPropose)
		return
	}
	retVal1 := localN.ProposePhase(emptyEty, slotPropose)
	if retVal1 == true {
		fmt.Println("Propossal was successful") //add: of value _
		//Update the entry's accNum and MaxPrepare
		ety.AccNum = localN.ProposalVal
		ety.MaxPrepare = localN.ProposalVal
		retVal2 := localN.AcceptPhase(ety, slotPropose)

		if retVal2 == true {
			fmt.Println("Accept Phase and Commit was successful, proposed entry has been added to the log")
		} else if retVal2 == false {
			fmt.Println("Failure: Accept phase was unsuccessful")
		}
	} else if retVal1 == false {
		fmt.Println("Failure: Propossal was Unsuccessful")
	}
}

func (localN *Node) RecoveryProposeHandler(ety entry, slotPropose int) bool {
	var emptyEty entry
	retVal1 := localN.RecoveryProposePhase(emptyEty, slotPropose)
	//Add a separate return that checks to see if any sites have sent back a message that indicates you are all caught up
	if retVal1 == true {
		fmt.Println("Propossal was successful") //add: of value _
		//The recieved value is stored in accVal, so use that instead of your own
		retVal2 := localN.RecoveryAcceptPhase(localN.AccVal, slotPropose)

		if retVal2 == true {
			fmt.Println("Accept Phase and Commit was successful, proposed entry has been added to the log")
			return true
		} else if retVal2 == false {
			fmt.Println("Failure: Accept phase was unsuccessful")
			return false
		}
	} else if retVal1 == false {
		fmt.Println("Failure: Propossal was Unsuccessful")
		return false
	}
	return false
}

func (localN *Node) LeaderPropseHandler(ety entry, slotPropose int) {
	ety.AccNum = leaderPropossal
	check := localN.AcceptPhase(ety, slotPropose)
	if check == true {
		fmt.Println("Accept Phase and Commit was successful, proposed entry has been added to the log")
	} else if check == false {
		fmt.Println("Failure: Accept phase was unsuccessful")
	}
}

func (localN *Node) TweetEvent(message string) {
	//localN.NodeMutex.Lock()
	//defer localN.NodeMutex.Unlock()
	ety := entry{message, localN.Id, localN.Id, time.Now().UTC(), 0, localN.ProposalVal, localN.ProposalVal, localN.SlotCounter}
	localN.ProposeHandler(ety, localN.SlotCounter)
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
	//localN.NodeMutex.Lock()
	//defer localN.NodeMutex.Unlock()
	userID, _ := strconv.Atoi(username)
	etyBlock := entry{"", localN.Id, userID, time.Now().UTC(), 1, localN.ProposalVal, localN.ProposalVal, localN.SlotCounter}
	localN.ProposeHandler(etyBlock, localN.SlotCounter)
}

func (localN *Node) UnblockUser(username string) {
	if localN.InvalidBlock(username, 2) == true {
		log.Println("Invalid UnBlock Call")
		return
	}
	//localN.NodeMutex.Lock()
	//defer localN.NodeMutex.Unlock()
	userID, _ := strconv.Atoi(username)
	etyUnblock := entry{"", localN.Id, userID, time.Now().UTC(), 2, localN.ProposalVal, localN.ProposalVal, localN.SlotCounter}
	localN.ProposeHandler(etyUnblock, localN.SlotCounter)
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
