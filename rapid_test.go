package main

import (
    "fmt"
    "./rapid"
	"flag";
	"log";
)

var debug = flag.Bool("v", false, "verbose")

func main() {
	events := make(chan *rapid.Event)
	done := make(chan bool)

	go rapid.Connect(events, done)
	
	for {
		select {
		case i := <-events:
			fmt.Printf("i: %#v\n", i)
		case <-done:
			log.Println("Done")
			return
		}
	}
}
