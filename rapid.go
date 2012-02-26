package rapid

import (
	"log"
	"os"
	"net"
	"fmt"
	"json"
	"time"
)

const debug = true

type EventChan chan<- *Event

type Event struct {
	TypeName  string
	Params    map[string]interface{}
}

type Ctx struct {
	ServerAddress     string
	ReconnectTimeout  uint
	Conn              net.Conn
	Out               EventChan
	Connected         bool
	shutdown          bool
}

// out is a channel that receives Events
func NewContext(out EventChan) *Ctx {
	ctx := new(Ctx)
	ctx.Out = out

	return ctx
}

// dest can be a string containing a server address, or it can be
// an already-existing net.Conn
func (ctx *Ctx) InitiateConnection(dest interface{}) {
	ctx.shutdown = false
	
	switch destType := dest.(type) {
	case nil:
		panic("destination is required in rapid.Connect()")
	case string:
		// handed a server address to connect to
		ctx.Connected = false
		ctx.ServerAddress = destType
	case net.Conn:
		// handed an already-existing socket
		ctx.Connected = true
		ctx.Conn = destType
	}
	
	if !ctx.Connected {
		// connect to server
		Log("Connecting to %s", ctx.ServerAddress)
		var err os.Error
		ctx.Conn, err = net.Dial("tcp", ctx.ServerAddress)

		if err != nil {
			log.Printf("Error connecting: %s\n", err)
			return
		}
	}

	ctx.Connected = true
	Debug("Connected")
}

func (ctx *Ctx) Read(p []byte) (int, os.Error) {
	// read a chunk of data
	return ctx.Conn.Read(p[:])
}

// loops until Shutdown() is called
// reconnects automatically if connection is lost
func (ctx *Ctx) ClientLoop() {
	for !ctx.shutdown {
		// reconnect if we're not connected
		if !ctx.Connected || ctx.Conn == nil {
			ctx.Reconnect()
			continue
		}
		
		// read a chunk of data
		err := ctx.ReadEvents()

		// failed to read
		if err != nil {
			Debug("failed to read")
			
			if ctx.shutdown {
				// socket got closed while we were reading, whatever
				continue
			}

			// we should no longer consider ourselves connected
			ctx.Connected = false
			ctx.Conn.Close()

			Log("Error reading from connection: %v", err)
		
			continue
		}
	}
}

/*

 what I really want is for this to take a JSON key named "type" and stick it in Event.TypeName...
 
func (evt *Event) UnmarshalJSON(data []byte) os.Error {
	
	fmt.Printf("got data: '%s'\n", data)

	err := json.Unmarshal(data, evt)
	if err != nil {
		fmt.Printf("Error decoding '%s': %v\n", data, err)
	}

	fmt.Printf("Evt: %v", evt)

	return err
}
*/
func (ctx *Ctx) ReadEvents() os.Error {
	// we expect a slice of Events
	type Outer struct {
		Events []*Event
	}
	var outer Outer

	// read from stream, parse Events into outer
	decoder := json.NewDecoder(ctx)
	err := decoder.Decode(&outer)

	if err != nil {
		return err
	}

	for _, evt := range outer.Events {
		log.Printf("Decoded object: %v\n", evt)
		ctx.Out <- evt
	}

	return nil
}

func (ctx *Ctx) Shutdown() {
	Debug("shutdown")
	ctx.Connected = false
	ctx.shutdown = true

	if ctx.Conn != nil {
		ctx.Conn.Close()
	}
}

func (ctx *Ctx) Reconnect() {
	if ctx.shutdown {
		return
	}
	
	Debug("reconnect")

	// can we reconnect?
	if ctx.ServerAddress == "" {
		log.Println("ServerAddress is not defined, cannot reconnect")
		ctx.Shutdown()
		return
	}

	// wait 3 seconds before reconnecting
	time.Sleep(1000 * 1000 * 1000 * 3)

	ctx.InitiateConnection(ctx.ServerAddress)
}

func Log(format string, v ...interface{}) {
	if !debug {
		return
	}

	ret := fmt.Sprintf(format, v)
	Debug(ret)
}

func Debug(ln string) {
	if !debug {
		return
	}
	log.Println(ln)
}
