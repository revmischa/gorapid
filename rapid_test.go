package rapid

import (
	"testing"
	"net"
	"fmt"
)

func TestConnection (t *testing.T) {
	server, client := net.Pipe();
	net.Pipe();

	events := make(chan *Event)
	ctx := NewContext(events)
	ctx.InitiateConnection(client)

	if !ctx.Connected {
		t.Error("Failed to connect to local pipe")
		return
	}

	defer ctx.Shutdown()

	done := false
	
	go func () {
		ctx.ClientLoop()
		done = true
	}()

	server.Write([]byte("{ \"type\":\"test_type\", \"params\": { \"x\": 1, \"y\": 2 } }\x00{"));
	server.Write([]byte(`"type":"final_test_type", "extraJunk": 123}`))
	
	for !done {
		select {
		case i := <-events:
			if (i.TypeName == "final_test_type") {
				// success
				return
			} else {
				err := fmt.Sprintf("Got unexpected event %#v.", i.TypeName)
				fmt.Println(err)
				t.Error(err)
				return
			}
		}
	}
}
