package main

import (
	"log"

	r "github.com/dancannon/gorethink"
	"github.com/gorilla/websocket"
)

type FindHandler func(string) (Handler, bool)

// Client primary responsibility reading and writing messages to websocket
// when browser connects to server, create new client object
// when client reads message using router, it will find appropriate function to handle request and call
// when function needs to send messages back to browser use client.send channel
//client will also clean up when browser disconnects
//when client subscribes to channel changes a go routine is created that queries rethinkdb and blocks indefinitely
//waiting for channel related changes to be streamed back to browser. when user disconnects, we
//need to disconnect to avoid goroutine leak
//send is a channel that can pass Message{}
//socket field is pointer to websocket connection
//findHandler type is a function that takes string and returns a Handler and boolean
type Client struct {
	send         chan Message
	socket       *websocket.Conn
	findHandler  FindHandler
	session      *r.Session
	stopChannels map[int]chan bool
	id           string
	userName     string
}

//Golang Channel = communication pipe sharing data between go routines
func (c *Client) NewStopChannel(stopKey int) chan bool {
	c.StopForKey(stopKey)
	stop := make(chan bool)
	c.stopChannels[stopKey] = stop
	return stop
}

func (c *Client) StopForKey(key int) {
	if ch, found := c.stopChannels[key]; found {
		ch <- true
		delete(c.stopChannels, key)
	}
}

//<- pull single value out of channel
func (client *Client) Read() {
	var message Message
	for {
		// if error break out of for loop
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}
		//call a function to lookup and return
		//handler based on message name. findHandler returns two values. store them in variables
		//handler and found. if handler is found call it passing the client so function can send
		//response if appropriate and the data/message payload.
		if handler, found := client.findHandler(message.Name); found {
			handler(client, message.Data)
		}
	}
	client.socket.Close()
}

//<- pull single value out of channel
func (client *Client) Write() {
	// use range keyword to iterate over values sent through go channel. when no channels being sent for
	//block waits
	for msg := range client.send {
		if err := client.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	client.socket.Close()
}

//client cleans up after itself when browser disconnects by closing any
//changefeeds and exiting all goroutines
func (c *Client) Close() {
	for _, ch := range c.stopChannels {
		ch <- true
	}
	close(c.send)
	// delete user
	r.Table("user").Get(c.id).Delete().Exec(c.session)
}

//NewClient creates new obj and returns it. Returns pointer to newly instantiated Client
//explicitly references field send and assigns value after :  Ending comma required even as last field
func NewClient(socket *websocket.Conn, findHandler FindHandler,
	session *r.Session) *Client {
	var user User
	user.Name = "anonymous"
	res, err := r.Table("user").Insert(user).RunWrite(session)
	if err != nil {
		log.Println(err.Error())
	}
	var id string
	if len(res.GeneratedKeys) > 0 {
		id = res.GeneratedKeys[0]
	}
	return &Client{
		send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
		id:           id,
		userName:     user.Name,
	}
}
