package main

import (
	"fmt"
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/gorilla/websocket"
)

type Handler func(*Client, interface{})

//Upgrader hijacks connection and switch protocol from http to websockets
//add initial property/field values- Read Write CheckOrigin-Server
//determines origin policy(allow any connection)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

//Router Struct go map is a dictionary of key value pairs added as field in our struct
//keys are strings and values are handler functions
type Router struct {
	rules   map[string]Handler
	session *r.Session
}

//NewRouter map needs to be initialized before it can be used, no constructors in Go. typical pattern
//is to create a func that creates and returns an object.
//NewRouter with *Router pointer return type. inside function return a pointer to the new &Router
//inside braces initialize new rules map
func NewRouter(session *r.Session) *Router {
	return &Router{
		rules:   make(map[string]Handler),
		session: session,
	}
}

//Handle recieves pointer to router, attaching function to Router struct
//what to do with msgName and handler function? store rule in router for later use
//add provided rule, event name, and function to handle it to map with r.rules[msgName] = handler
func (r *Router) Handle(msgName string, handler Handler) {
	r.rules[msgName] = handler
}

//FindHandler optional second return value when getting value from a map. boolean indicating if
//value stored for provided key
func (r *Router) FindHandler(msgName string) (Handler, bool) {
	handler, found := r.rules[msgName]
	return handler, found
}

//ServeHTTP return http 500 if error
//Client will need websocket connection to read and write to it so pass socket to client
//Write uses own goroutine,read method uses the goroutine created when ServeHTTP is called
func (e *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	client := NewClient(socket, e.FindHandler, e.session)
	defer client.Close()
	go client.Write()
	client.Read()

}
