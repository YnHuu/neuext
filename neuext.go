package neuext

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Extension struct {
	methods     map[string]reflect.Value
	conn        *websocket.Conn
	accessToken string
	mutex       sync.Mutex
	debug       bool
}

func (e *Extension) Debug() {
	e.debug = true
}

func (e *Extension) Register(rcvr any) {
	typ := reflect.TypeOf(rcvr)
	vof := reflect.ValueOf(rcvr)
	e.methods = make(map[string]reflect.Value)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mname := method.Name
		if !method.IsExported() {
			continue
		}
		name := strings.ToLower(mname[:1]) + mname[1:]
		// fmt.Println("name", name)
		e.methods[name] = vof.MethodByName(mname)
	}
}

func (e *Extension) WSDial(uri, accessToken string) error {
	e.accessToken = accessToken

	c, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		return err
	}
	defer c.Close()
	e.conn = c
	for {
		typ, msg, err := c.ReadMessage()
		if err != nil {
			return err
		}
		if typ != websocket.TextMessage {
			continue
		}
		if e.debug {
			log.Printf("recv: %s\n", msg)
		}

		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			continue
		}
		if _, ok := data["event"]; !ok {
			continue
		}
		if _, ok := e.methods[data["event"].(string)]; !ok {
			continue
		}
		if data["data"] == nil {
			e.methods[data["event"].(string)].Call(nil)
			continue
		}
		args := []reflect.Value{reflect.ValueOf(data["data"])}
		e.methods[data["event"].(string)].Call(args)
	}
}

func (e *Extension) Send(event, body any) {
	if e.conn == nil {
		if e.debug {
			log.Println("websocket is nil")
		}
		return
	}
	m := make(map[string]any)
	m["id"] = uuid.NewString()
	m["method"] = "app.broadcast"
	m["accessToken"] = e.accessToken
	m["data"] = map[string]any{"event": event, "data": body}
	e.mutex.Lock()
	defer e.mutex.Unlock()
	_ = e.conn.WriteJSON(m)
	if e.debug {
		log.Printf("send: %+v\n", m)
	}
}

func (e *Extension) Close() {
	_ = e.conn.Close()
}
