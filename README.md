# Distributed Systems and Algorithms - Project 2
CSCI-4510/6510, Fall 2017

Implementation of distributed Twitter service using the Paxos algorithm to replicate tweet, block, and unblock events. More details located in Project Specifications folder. 

## Program Execution
* Must have golang installed (make sure you set the go path, ex: export GOPATH="/home/ubuntu/go")
* This program was tested with five AWS ec2 instances (feel free to create your own or test this on your local machine)
* Create directory in go for src/github.com
* Clone this repository and go into the directory
* Input command: "go build main.go listen.go entries.go input.go node.go send.go receive.go"
* Input command: "./main <*json entry data file*> <*user id for site*>"
