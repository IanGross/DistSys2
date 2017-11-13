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
const staticDict = "./localDict.json"
const staticTArray = "./localTArray.json"

type Node struct {
	Id int
	Ci int

	NodeMutex *sync.Mutex

	UserName string

	Log [][]tweet

	TimeArray [][]int

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
		Names      map[string]string
		TotalNodes int
		IPs        map[string]string
	}

	var info startinfo //Deserialize the JSON
	if err := json.Unmarshal(file, &info); err != nil {
		log.Fatal(err)
	}

	ret.NodeMutex = &sync.Mutex{}
	ret.Id = inputID
	ret.Ci = 0
	ret.UserName = info.Names[strconv.Itoa(ret.Id)]

	parts := strings.Split(info.IPs[strconv.Itoa(ret.Id)], ":")
	ret.ListenPort, err = strconv.Atoi(parts[1])
	if err != nil {
		log.Panicln("Failed to get IP address for local node")
		log.Fatalln(err)
	}

	ret.Log = make([][]tweet, info.TotalNodes)
	for i := 0; i < info.TotalNodes; i++ {
		ret.Log[i] = make([]tweet, 0, 10)
	}

	if check, err := ret.LoadTweets(staticLog); err != nil || check == false {
		//create file
		f, err := os.Create(staticLog)
		if err != nil {
			log.Fatal("cannot create log")
		}
		f.Close()
	}

	ret.TimeArray = make([][]int, info.TotalNodes)
	for i := range ret.TimeArray {
		ret.TimeArray[i] = make([]int, info.TotalNodes)
		for z := range ret.TimeArray {
			ret.TimeArray[i][z] = 0
		}
	}
	if check, err := ret.LoadTArray(); err != nil || check == false {
		f, err := os.Create(staticTArray)
		if err != nil {
			log.Fatalln("cannot created staticTarray")
		}
		f.Close()
	}

	ret.Ci = ret.TimeArray[ret.Id][ret.Id]

	ret.Blocks = make(map[int]map[int]bool)

	for i := 0; i < info.TotalNodes; i++ {
		ret.Blocks[i] = make(map[int]bool)
	}

	check, err := ret.LoadDict()
	if err != nil || check == false {
		//create static dict
		f, err := os.Create(staticDict)
		if err != nil {
			log.Fatal("Failed to create staticDict")
		}
		f.Close()
	}

	ret.IPtargets = make(map[int]string)
	//Populate the IPtargets
	for keyValue, mapValue := range info.IPs {
		idInt, _ := strconv.Atoi(keyValue)
		if idInt != ret.Id {
			ret.IPtargets[idInt] = mapValue
		}
	}

	ret.Names = make(map[int]string)
	//Populate the IPtargets
	for keyValue, mapValue := range info.Names {
		idInt, _ := strconv.Atoi(keyValue)
		if idInt != ret.Id {
			ret.Names[idInt] = mapValue
		}
	}

	return ret
}

func (n *Node) LoadTweets(filename string) (bool, error) {
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
	n.Ci = len(n.Log[n.Id])

	return true, nil
}

func (n *Node) LoadDict() (bool, error) {
	_, err := os.Stat(staticDict)
	if os.IsNotExist(err) {
		log.Println("DICT FILE NOT YET CREATED")
		return false, nil
	}

	file, err := ioutil.ReadFile(staticDict)
	if err != nil {
		return false, err
	}
	cheat := make(map[string]map[string]bool)

	if err := json.Unmarshal(file, &cheat); err != nil {
		return false, err
	}

	for i := range cheat {
		for j := range cheat[i] {
			key1, erra := strconv.Atoi(i)
			key2, errb := strconv.Atoi(j)
			if erra != nil || errb != nil {
				log.Fatalln("Failed to read staticDict", erra, errb)
			}
			n.Blocks[key1][key2] = cheat[i][j]
		}
	}

	return true, nil
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

func (n *Node) writeDict() {
	cheat := make(map[string]map[string]bool)

	for i := range n.Blocks {
		cheat[strconv.Itoa(i)] = make(map[string]bool)
		for j := range n.Blocks[i] {
			cheat[strconv.Itoa(i)][strconv.Itoa(j)] = n.Blocks[i][j]
		}
	}

	writeBytes, err := json.MarshalIndent(cheat, "", "  ")
	if err != nil {
		log.Println(err)
		log.Fatalf("failed to Marshel to static Dict\n")
	}
	err = ioutil.WriteFile(staticDict, writeBytes, 0644)
	if err != nil {
		log.Fatalln("Failed to write to static dict")
	}
}

func (n *Node) writeTArray() {
	writeBytes, err := json.MarshalIndent(n.TimeArray, "", "  ")
	if err != nil {
		log.Fatalf("failed to Marshel to static Dict\n")
	}
	err = ioutil.WriteFile(staticTArray, writeBytes, 0644)
	if err != nil {
		log.Fatalln("Failed to write to static dict")
	}
}

func (n *Node) LoadTArray() (bool, error) {
	_, err := os.Stat(staticTArray)
	if os.IsNotExist(err) {
		log.Println("TARRAY FILE NOT YET CREATED")
		return false, nil
	}

	file, err := ioutil.ReadFile(staticDict)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(file, &n.TimeArray); err != nil {
		return false, err
	}
	return true, nil
}

func (n *Node) hasRec(msg tweet, k int) bool {
	ret := n.TimeArray[k][msg.User] >= msg.Counter
	return ret
}

func (n *Node) updateLocalTimeArray() {
	var tArr []int
	for i := 0; i < len(n.Log); i++ {
		tArr = append(tArr, len(n.Log[i]))
	}
	n.TimeArray[n.Id] = tArr
}

func (n *Node) incrementClock() {
	n.Ci = n.Ci + 1
	n.TimeArray[n.Id][n.Id] = n.Ci
}

func (n *Node) CleanDict() {
	for i := range n.Blocks {
		for j := range n.Blocks[i] {
			if false == n.Blocks[i][j] {
				n.truncate(i, j)
			}
		}
	}
}

func (n *Node) truncate(user int, follower int) {
	// check everyone has recieved delete command
	var t tweet

	for i := len(n.Log[user]) - 1; i >= 0; i-- {
		if n.Log[user][i].Follower == follower && n.Log[user][i].Event == DELETE {
			t = n.Log[user][i]
			break
		}
	}
	// if any node hasn't recieved the unblock, don't remove it.
	for i := range n.TimeArray {
		if false == n.hasRec(t, i) {
			return
		}
	}
	i := 0
	for _, x := range n.Log[user] {
		if !rmcheck(&x, user, follower) {
			n.Log[user][i] = x
			i++
		}
	}
	n.Log[user] = n.Log[user][:i]
	//Since the record has been removed, we can delete
	delete(n.Blocks[user], follower)
}

func rmcheck(t *tweet, use int, fol int) bool {
	return (t.Follower == fol && t.User == use && (t.Event == DELETE || t.Event == INSERT))
}
