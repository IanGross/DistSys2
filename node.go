package main

import (
	"encoding/json"
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
	SlotCounter         int
	LeaderID            int
	//TO DO: Add an acceptor bool that is set to false initially
	//	sets to true when it sets its MaxPrepare after a propose, to only accept from that MaxPrepare value
	//	sets back to false when it commits to the log

	NodeMutex *sync.Mutex

	Log []entry

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
	ret.MaxPrepare = -1
	ret.AccNum = -1
	var itt entry
	ret.AccVal = itt
	ret.ProposalVal = inputID
	ret.MajorityVal = info.TotalNodes/2 + 1
	ret.SlotCounter = 0
	ret.LeaderID = info.EntryLeaderID

	parts := strings.Split(info.IPs[strconv.Itoa(ret.Id)], ":")
	ret.ListenPort, err = strconv.Atoi(parts[1])
	if err != nil {
		log.Panicln("Failed to get IP address for local node")
		log.Fatalln(err)
	}

	/*
		ret.Log = make([][]entry, info.TotalNodes)
		for i := 0; i < info.TotalNodes; i++ {
			ret.Log[i] = make([]entry, 0, 10)
		}
	*/
	//changed for
	ret.Log = make([]entry, 0, 10)

	if check, err := ret.LoadEntries(staticLog); err != nil || check == false {
		//create file
		f, err := os.Create(staticLog)
		if err != nil {
			log.Fatal("cannot create log")
		}
		f.Close()
	}

	//TO DO: implement paxos here to load the missing entries from other sites

	ret.Blocks = make(map[int]map[int]bool)

	for i := 0; i < info.TotalNodes; i++ {
		ret.Blocks[i] = make(map[int]bool)
	}

	//TO DO: Update the dictionary based on the log entries
	//err := ret.LoadDict()

	ret.IPtargets = make(map[int]string)
	//Populate the IPtargets
	for keyValue, mapValue := range info.IPs {
		idInt, _ := strconv.Atoi(keyValue)
		//if idInt != ret.Id {
		ret.IPtargets[idInt] = mapValue
		//}
	}

	ret.Names = make(map[int]string)
	//Populate the IPtargets
	for keyValue, mapValue := range info.Names {
		idInt, _ := strconv.Atoi(keyValue)
		//if idInt != ret.Id {
		ret.Names[idInt] = mapValue
		//}
	}

	return ret
}

func (n *Node) IncrementPropossalVal() {
	n.ProposalVal += len(n.IPtargets)
	return
}

func (n *Node) LoadEntries(filename string) (bool, error) {
	_, err := os.Stat(staticLog)
	if os.IsNotExist(err) {
		log.Println("LOG FILE NOT YET CREATED")
		return false, nil
	}

	file, err := ioutil.ReadFile(staticLog)
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(file, &n.Log); err != nil {
		return false, err
	}

	//UPDATE NEEDED HERE:
	// Get and update MaxPrepare, AccNum and AccVal

	return true, nil
}

func (n *Node) LoadDict() error {
	//NEW FUNCTION REQUIRED:
	// Because we are no longer storing a separate log for dictionary values, we need to load the information separately
	// Get the dictionary from the log

	// Pseudocode:
	//		Run through the sorted list of log entries
	//		If the entry is block or unblock, perform the update to local dictionary without sending it to the other locations

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
