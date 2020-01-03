package main

import (
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/mitchellh/mapstructure"
)

//iota sets ChannelStop = 0, UserStop = 1, MessageStop = 2
const (
	ChannelStop = iota
	UserStop
	MessageStop
)

//Message add field tags to allow for lowercase. capitalization semantic in Go,capital N in Name
//indicates Public. n in name would indicate private to package.main
// interface{} specifies the behavior,dont know what
//data type is until we determine message or Name.Empty interface acts as placeholder
type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

//Channel struct, collection of fields and potentially methods. define struct and unmarshal/decode
//json bytearray into our struct
type Channel struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

//User go representation of user for new user record. saved in lcase in db
//with gorethink package field tags. don't encode Id if empty.
type User struct {
	Id   string `gorethink:"id,omitempty"`
	Name string `gorethink:"name"`
}

//ChannelMessage Struct
type ChannelMessage struct {
	Id        string    `gorethink:"id,omitempty"`
	ChannelId string    `gorethink:"channelId"`
	Body      string    `gorethink:"body"`
	Author    string    `gorethink:"author"`
	CreatedAt time.Time `gorethink:"createdAt"`
}

//
func editUser(client *Client, data interface{}) {
	var user User
	err := mapstructure.Decode(data, &user)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	client.userName = user.Name
	go func() {
		_, err := r.Table("user").
			Get(client.id).
			Update(user).
			RunWrite(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}
func subscribeUser(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(UserStop)
		cursor, err := r.Table("user").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)

		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "user", client.send, stop)
	}()
}

func unsubscribeUser(client *Client, data interface{}) {
	client.StopForKey(UserStop)
}

func addChannelMessage(client *Client, data interface{}) {
	var channelMessage ChannelMessage
	err := mapstructure.Decode(data, &channelMessage)
	if err != nil {
		client.send <- Message{"error", err.Error()}
	}
	go func() {
		channelMessage.CreatedAt = time.Now()
		channelMessage.Author = client.userName
		err := r.Table("message").
			Insert(channelMessage).
			Exec(client.session)

		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

func subscribeChannelMessage(client *Client, data interface{}) {
	go func() {
		eventData := data.(map[string]interface{})
		val, ok := eventData["channelId"]
		if !ok {
			return
		}
		channelId, ok := val.(string)
		if !ok {
			return
		}
		stop := client.NewStopChannel(MessageStop)
		cursor, err := r.Table("message").
			OrderBy(r.OrderByOpts{Index: r.Desc("createdAt")}).
			Filter(r.Row.Field("channelId").Eq(channelId)).
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)

		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "message", client.send, stop)
	}()
}

func unsubscribeChannelMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}

//addChannel what resources will function need to handle adding channel to database?
//will need pointer to Client for error handling to send(client.send) message back to browser
//will need data payload, an empty interface
// type Handler func(*Client, interface{})
func addChannel(client *Client, data interface{}) {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	go func() {
		err = r.Table("channel").
			Insert(channel).
			Exec(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

//subscribeChannel handles subscribe events/messages sent from browser. starts rethinkdb changefeed
//that sends new channels as they are added to browser to be displayed in channel list
//blocking operation so utilize goroutine
//changefeed returns r.ChangeResponse record with newVal and oldVal
func subscribeChannel(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(ChannelStop)
		cursor, err := r.Table("channel").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "channel", client.send, stop)
	}()
}

//unsubscribeChannel adds new method to Client struct that returns
// channel that returns a c
func unsubscribeChannel(client *Client, data interface{}) {
	client.StopForKey(ChannelStop)
}

// changefeed result has two fields newVal and oldVal
// if update newVal it becomes oldVal. if deleted the newVal
// becomes null and old val is whatever it was before delete
//stop <-chan bool indicates changeFeedHelper is receive only on stop channel
//go select allows wait on multiple channels concurrently with the first
//channel to provide a value triggers it's code block
func changeFeedHelper(cursor *r.Cursor, changeEventName string,
	send chan<- Message, stop <-chan bool) {
	change := make(chan r.ChangeResponse)
	cursor.Listen(change)
	for {
		eventName := ""
		var data interface{}
		select {
		case <-stop:
			//cursor.Close()stops changefeed
			cursor.Close()
			//return stops subscribe goroutine
			return
		//send changefeed result
		case val := <-change:
			if val.NewValue != nil && val.OldValue == nil {
				eventName = changeEventName + " add"
				data = val.NewValue
			} else if val.NewValue == nil && val.OldValue != nil {
				eventName = changeEventName + " remove"
				data = val.OldValue
			} else if val.NewValue != nil && val.OldValue != nil {
				eventName = changeEventName + " edit"
				data = val.NewValue
			}
			send <- Message{eventName, data}
		}
	}
}
