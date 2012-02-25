package rapid

import (
	"log"
	"os"
	"net"
	"fmt"
	"json"
	"bytes"
	"time"
)

const debug = true

type EventChan chan<- *Event

type Event struct {
	root      map[string]interface{}
	type_name string
	params    map[string]string
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

// loops until Shutdown() is called
// reconnects automatically if connection is lost
func (ctx *Ctx) ClientLoop() {
	var buf [4048]byte

	for !ctx.shutdown {
		// reconnect if we're not connected
		if !ctx.Connected || ctx.Conn == nil {
			ctx.Reconnect()
			continue
		}
		
		// read a chunk of data
		len, err := ctx.Conn.Read(buf[:])

		// failed to read
		if len == 0 || err != nil {
			if ctx.shutdown {
				// socket got closed while we were reading, whatever
				continue
			}

			// we should no longer consider ourselves connected
			ctx.Connected = false
			ctx.Conn.Close()

			if err != nil {
				Log("Error reading from connection: %v", err)
			}
		
			continue
		}

		// split JSON on NULL char
		sep := []byte{0}
		elements := bytes.Split(buf[0:4048], sep)

		for _, e := range elements {
			ctx.parseFragment(e, ctx.Out)
		}
	}
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
	}

	// wait 3 seconds before reconnecting
	time.Sleep(1000 * 1000 * 1000 * 3)

	ctx.InitiateConnection(ctx.ServerAddress)
}

func (ctx *Ctx) parseFragment(frag []byte, out EventChan) {
	var root map[string]interface{}

	if len(frag) == 0 {
		return
	}
	
	err := json.Unmarshal(frag, &root)
	if err != nil || root == nil {
		log.Printf("Error parsing JSON '%s': %v\n", frag, err)
		os.Exit(1)
		return
	}
	
	params := make(map[string]string)
	var type_name string

	switch nametype := root["type"].(type) {
	case nil:
		log.Printf("Got JSON message with no type field defined\n")
	case string:
		log.Printf("Got message with type_name=%v\n", nametype)
		type_name = nametype
	default:
		log.Printf("Got JSON message with unknown type for 'type' field\n")
	}

	evt := Event{root, type_name, params}
	out <- &evt
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
