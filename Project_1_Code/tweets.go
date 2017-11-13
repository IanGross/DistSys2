package main

import (
	"encoding/json"
	"log"
	"time"
)

type message struct {
	Ti     [][]int
	Events [][]tweet
	SendID int
}

const ( //iota is reset to 0
	TWEET  = iota
	INSERT = iota // INSERT=0
	DELETE = iota // DELETE=1
)



type tweet struct {
	Message  string
	User     int
	Follower int
	Clock    time.Time
	Counter  int
	Event    int
	//event is the type (tweet,insert,delete)
}

func (n tweet) getTimestamp() time.Time {
	return n.Clock
}

func (n tweet) getUser() int {
	return n.User
}

func (n tweet) getJSON() []byte {
	ret, err := json.Marshal(n)
	if err != nil {
		log.Printf("Failed to create JSON")
		return nil
	}
	return ret
}

func getTweets(msg []byte) ([]tweet, error) {
	var ret []tweet
	err := json.Unmarshal(msg, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func maxInt(x int, y int) int {
	if x > y {
		return x
	}
	return y
}
