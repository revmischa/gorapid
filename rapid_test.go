package rapid

import (
	"log"
	"testing"
	"net"
)

func TestConnection (t *testing.T) {
	server, client := net.Pipe();
	net.Pipe();
			
	events := make(chan *Event)
	done := make(chan bool)

	go func () {
		Connect(events, done, client)
	}()

	server.Write([]byte("{ \"type\":\"test_type\", \"params\": { \"x\": 1, \"y\": 2 } }\x00"));
	
	for {
		select {
		case i := <-events:
			if (i.type_name == "test_type") {
				// success
				return
			} else {
				t.Errorf("Got unexpected event %#v.", i.type_name)
			}
		case <-done:
			log.Println("Done")
			return
		}
	}
}
