package main

import (
	"encoding/json"
	"log"
	"time"
)

const ( //iota is reset to 0
	TWEET  = iota // TWEET=0
	INSERT = iota // INSERT=1
	DELETE = iota // DELETE=2
)

const ( //iota is reset to 0
	PROPOSE = iota
	PROMISE = iota
	ACCEPT  = iota
	ACK     = iota
	COMMIT  = iota
)

type entry struct {
	Message  string
	User     int
	Follower int
	Clock    time.Time
	EntryVal int
	Event    int
	//event is the type (tweet,insert,delete)
}

//struct for propose, promise, accept, ack, and commit messages
type message struct {
	UserID  int
	MsgType int
	ANum    int
	AVal    entry
}

func (n entry) getTimestamp() time.Time {
	return n.Clock
}

func (n entry) getUser() int {
	return n.User
}

func (n entry) getJSON() []byte {
	ret, err := json.Marshal(n)
	if err != nil {
		log.Printf("Failed to create JSON")
		return nil
	}
	return ret
}

func getEntries(msg []byte) ([]entry, error) {
	var ret []entry
	err := json.Unmarshal(msg, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
