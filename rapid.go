package rapid

import (
	"log";
	"os";
	"net";
	"fmt";
	"time";
	"json";
	"strings";
)

const debug = true

type Event struct {
	root map[string]interface{}
	type_name string
	params map[string]string
}

func Connect(out chan<- *Event, done chan<- bool) (err os.Error) {
	server := "localhost:6000"
	connected := false
	var c net.Conn
	
	for ! connected {
		Log("Connecting to %s", server)
		c, err = net.Dial("tcp", server)

		if err != nil {
			log.Printf("Error connecting: %s\n", err)
			connected = false;
			time.Sleep(1000 * 1000 * 1000 * 3)
			continue
		} else {
			connected = true
		}

		Debug("Connected")
		var buf [4048]byte
		len, err := c.Read(buf[0:4047])

		if len == 0 || err != nil {
			c.Close()
			connected = false;
		}

		var elements []string = strings.Split(buf[0:4047], "\x00")

		for i, element := range elements {
			parseFragment(element)
		}
	}
}

func parseFragment(frag string) {
	var root map[string]interface{}
	
	err := json.Unmarshal(frag, &root)
	if err != nil || root == nil {
		log.Printf("Error parsing JSON '%s': %v\n", buf, err)
		return
	} else {
		params := make(map[string]string)
		var type_name string

		switch type_name_type := root["type"].(type) {
		case nil:
			log.Printf("Got JSON message with no type field defined\n")
		case string:
			log.Printf("Got message with type_name=%v\n")
			type_name = type_name_type
		default:
			log.Printf("Got JSON message with unknown type for 'type' field\n")
		}

		evt := Event{root, type_name, params}
		out <- &evt
	}
}

func Log(format string, v ...interface{}) {
	if ! debug {
		return
	}

	ret := fmt.Sprintf(format, v);
	Debug(ret)
}

func Debug(ln string) {
	if ! debug {
		return
	}
	log.Println(ln);
}
