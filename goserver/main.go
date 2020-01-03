package main

//r before gorethink package adds alias
import (
	"log"
	"net/http"

	r "github.com/dancannon/gorethink"
)

//where to connect(session) to database when starting server?
//session must be accessible to all handler functions
//create in Main and pass as field into router struct passing as param to
//NewRouter() and session field to passed session
//then pass from router to client. Add session as field to Client struct and
//add session as param to NewClient() and initial session field to passed session
//In Router's ServeHTTP method pass new session into NewClient() function call

//Handling 'channel unsubscribe' message from browser
//channel unsubscribe message triggers call to unsubscribeChannel handler
//unsubscribeChannel should send stop signal across appropriate stop chan
//client *Client is passed to unsubscribeChannel and has all stop chans
//in it's stop chan map

//router.Handle('channel unsubscribe', unsubscribeChannel)

// changefeed result has two fields newVal and oldVal
// if update newVal it becomes oldVal. if deleted the newVal
// becomes null and old val is whatever it was before delete

func main() {
	//Connect to rethinkdb, specify server address with ConnectOpts
	//and default database to avoid referencing it in every call.
	//Connect returns both the connection to the db or session(in pool)
	//and err
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "chat",
	})

	if err != nil {
		log.Panic(err.Error())
	}

	// //create a router for websocket requests
	router := NewRouter(session)

	router.Handle("channel add", addChannel)
	router.Handle("channel subscribe", subscribeChannel)
	router.Handle("channel unsubscribe", unsubscribeChannel)

	router.Handle("user edit", editUser)
	router.Handle("user subscribe", subscribeUser)
	router.Handle("user unsubscribe", unsubscribeUser)

	router.Handle("message add", addChannelMessage)
	router.Handle("message subscribe", subscribeChannelMessage)
	router.Handle("message unsubscribe", unsubscribeChannelMessage)

	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}
