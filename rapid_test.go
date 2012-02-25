package rapid

import (
	"testing"
	"net"
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
	
	go func () {
		ctx.ClientLoop()
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
				return
			}
		}
	}
}
