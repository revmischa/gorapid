package rapid

import (
	"log"
	"os"
	"net"
	"fmt"
	"json"
	"bytes"
)

const debug = true

type EventChan chan<- *Event

type Event struct {
	root      map[string]interface{}
	type_name string
	params    map[string]string
}

func Connect(out EventChan, done chan<- bool, c net.Conn) (os.Error) {
	server := "localhost:6000"
	connected := c != nil
	
	if !connected {
		Log("Connecting to %s", server)
		var err os.Error
		c, err = net.Dial("tcp", server)

		if err != nil {
			log.Printf("Error connecting: %s\n", err)
			return err
		}
	}

	Debug("Connected")
	var buf [4048]byte
	len, err := c.Read(buf[:])

	if len == 0 || err != nil {
		c.Close()
		return err
	}

	sep := []byte{0}
	elements := bytes.Split(buf[0:4048], sep)

	for _, e := range elements {
		parseFragment(e, out)
	}

	return nil
}

func parseFragment(frag []byte, out EventChan) {
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
