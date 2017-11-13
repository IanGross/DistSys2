package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("First argument must be configuration file, Second must be the ID of the location")
	}
	starting := os.Args[1]
	id := os.Args[2]

	if _, err := os.Stat(starting); os.IsNotExist(err) {
		log.Fatalln("Unable to open configuration file\nPlease confirm it's a json")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Fatalln("Integer not entered for second argument (id)")
	}

	//local := makeNode(starting, idInt)
	//create thread at separate function to listen for incomming messages
	//go listen(local)
	//InputHandler(local)
	return
}
