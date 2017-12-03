package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

//To Implement:
//  - Only use one log for storing information
//    - This means that the dictionary needs to be retrieved in a different way

const staticLog = "./localLog.json"
const emptyPropossal = -2
const leaderPropossal = -1

type Node struct {
	Id       int
	SiteName string

	MaxPrepare          int
	AccNum              int
	AccVal              entry
	ProposalVal         int
	MajorityVal         int
	RecvAcceptedPromise int
	RecvAcceptedAck     int
	CountSiteFailures   int
	SlotCounter         int
	LeaderID            int
	RecoverStop         bool

	OutputChannel chan string

	NodeMutex *sync.Mutex

	Log      []entry
	Timeline []entry

	Blocks map[int]map[int]bool

	ListenPort int
	IPtargets  map[int]string
	Names      map[int]string
}

func makeNode(inputfile string, inputID int) *Node {
	ret := new(Node)

	file, err := ioutil.ReadFile(inputfile)
	if err != nil {
		log.Fatal("Cannot Open file " + inputfile)
	}

	type startinfo struct {
		Names         map[string]string
		TotalNodes    int
		EntryLeaderID int
		IPs           map[string]string
	}

	var info startinfo //Deserialize the JSON
	if err := json.Unmarshal(file, &info); err != nil {
		log.Fatal(err)
	}

	ret.NodeMutex = &sync.Mutex{}
	ret.Id = inputID
	ret.SiteName = info.Names[strconv.Itoa(ret.Id)]
	//Set the initial values for these values? Or are they just null?
	ret.MaxPrepare = emptyPropossal
	ret.AccNum = emptyPropossal
	var itt entry
	ret.AccVal = itt
	ret.ProposalVal = inputID
	ret.MajorityVal = info.TotalNodes/2 + 1
	ret.SlotCounter = 0
	ret.LeaderID = info.EntryLeaderID
	ret.RecoverStop = false

	ret.OutputChannel = make(chan string)

	parts := strings.Split(info.IPs[strconv.Itoa(ret.Id)], ":")
	ret.ListenPort, err = strconv.Atoi(parts[1])
	if err != nil {
		log.Panicln("Failed to get IP address for local node")
		log.Fatalln(err)
	}

	//If log has not ben created before, prevent the recovery algorithm from running (used to prevent initial recovery)
	//	Assume that all sites start running
	//		If they don't, then comment this out (VERY IMPORTANT)
	firstRun := false
	//Create the log and load the entries from the static log
	ret.Log = make([]entry, 0, 10)
	if check, exists, err := ret.LoadEntries(staticLog); err != nil || check == false {
		//create file
		f, err := os.Create(staticLog)
		if err != nil {
			log.Fatal("cannot create log")
		}
		f.Close()
		firstRun = exists
	}

	//Make the dictionary
	ret.Blocks = make(map[int]map[int]bool)
	for i := 0; i < info.TotalNodes; i++ {
		ret.Blocks[i] = make(map[int]bool)
	}

	//Update the dictionary based on the log entries
	if err := ret.LoadDict(); err != nil {
		log.Fatal("Error with dictionary load")
	}

	//Populate the IPtargets
	ret.IPtargets = make(map[int]string)
	for keyValue, mapValue := range info.IPs {
		idInt, _ := strconv.Atoi(keyValue)
		//if idInt != ret.Id {
		ret.IPtargets[idInt] = mapValue
		//}
	}

	//Populate the Names
	ret.Names = make(map[int]string)
	for keyValue, mapValue := range info.Names {
		idInt, _ := strconv.Atoi(keyValue)
		//if idInt != ret.Id {
		ret.Names[idInt] = mapValue
		//}
	}

	//Checking if leader has already been elected to the node
	if len(ret.Log) != 0 {
		ret.LeaderID = ret.Log[len(ret.Log)-1].User
	}

	//Update the SlotCounter
	ret.SlotCounter = len(ret.Log)

	//Utilize Paxos to recover any missing log entries that might have been missed during site failure
	if firstRun == false { //Don't do recovery if this is the first time the site has run
		if recoverCount, err := ret.LearnMissingEntries(); err != nil {
			log.Println("Error has occured while recovering missing entries")
		} else {
			if recoverCount == 1 { // That pural
				fmt.Printf("Site has learned about %v missing entry during recovery\n", recoverCount)
			} else if recoverCount > 1 || recoverCount == 0 {
				fmt.Printf("Site has learned about %v missing entries during recovery\n", recoverCount)
			}
		}
	}

	ret.UpdateTimeline()

	return ret
}

func (n *Node) LoadEntries(filename string) (bool, bool, error) {
	_, err := os.Stat(staticLog)
	if os.IsNotExist(err) {
		log.Println("LOG FILE NOT YET CREATED")
		return false, true, nil
	}

	file, err := ioutil.ReadFile(staticLog)
	if err != nil {
		return false, false, err
	}

	if err := json.Unmarshal(file, &n.Log); err != nil {
		return false, false, err
	}

	return true, false, nil
}

//LoadDict - Get the dictionary information from the log
func (n *Node) LoadDict() error {
	organizedLog := OrganizeEntries(n.Log)
	for i := 0; i < len(organizedLog); i++ {
		if organizedLog[i].Event == INSERT {
			n.Blocks[organizedLog[i].User][organizedLog[i].Follower] = true
		} else if organizedLog[i].Event == DELETE {
			delete(n.Blocks[organizedLog[i].User], organizedLog[i].Follower)
		}
	}
	return nil
}

func (n *Node) writeLog() {
	logBytes, err := json.MarshalIndent(n.Log, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(staticLog, logBytes, 0644)
	if err != nil {
		log.Fatalln("Failed to write to staticlog")
	}
}

func (n *Node) LearnMissingEntries() (int, error) {
	//SECTION IS GOING TO BE COMPLETELY CHANGED
	// Removal of FAIL
	// changes in recieve.go to check to see what log slot was proposed
	//	current node will stay the same, but lower slot will act differently
	//  Propose an entry of a special event type that the propose in receive.go will recognize

	//TO DO: implement paxos here to load the missing entries from other sites

	//In this case, assume that there are no holes in the log, so start recovery from the current slot value

	fmt.Println("In recovery function...")

	initialSlotCounter := n.SlotCounter
	for {
		var etyEmpty entry
		foundEntry := n.RecoveryProposeHandler(etyEmpty, n.SlotCounter)
		if foundEntry == false {
			return n.SlotCounter - initialSlotCounter, nil
		}
		//otherwise, a new value was commited to a log slot and try to find more missing values
	}
}

func (n *Node) IncrementPropossalVal() {
	n.ProposalVal += len(n.IPtargets)
	return
}

func (n *Node) CommitNodeUpdate() {
	n.LeaderID = n.AccVal.User
	n.SlotCounter++
	n.MaxPrepare = emptyPropossal
	n.AccNum = emptyPropossal
	var itt entry
	n.AccVal = itt
	n.ProposalVal = n.Id
	n.UpdateTimeline()
	return
}

func (n *Node) AmLeader() bool {
	return n.Id == n.LeaderID
}

func (n *Node) getProposeValue() int {
	if n.AmLeader() {
		return leaderPropossal
	}
	return n.ProposalVal
}

func (n *Node) UpdateTimeline() {
	n.Timeline = make([]entry, 0)
	organizedLog := OrganizeEntries(n.Log)
	logReverse := reverse(organizedLog)
	for i := 0; i < len(logReverse); i++ {
		if logReverse[i].Event == TWEET && !n.checkBlock(logReverse[i].User, n.Id) {
			//why so many Itoas? fmt.Printf("%i", somevalue ) would work
			n.Timeline = append(n.Timeline, logReverse[i])
		}
	}
}
