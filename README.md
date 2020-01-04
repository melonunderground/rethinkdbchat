# chat

install rethinkdb (https://rethinkdb.com/docs/install/)

Start RethinkDB server in terminal

  $ rethinkdb
  
  open RethinkDB Administrator
  http://localhost:8080/
  
  Select Data Explorer tab
  
  create database
    r.dbCreate('chat')
  
  create channel table
    r.db('chat').tableCreate('channel')
  
  create user table
    r.db('chat').tableCreate('user')
  
  create message table
   r.db('chat').tableCreate('message')
  
  create index channelId to message table
    r.db('chat').table('message').
      indexCreate('channelId')
    
  create index createdAt to message table
    r.db('chat').table('message').
      indexCreate('createdAt')
  
open directory /goserver in terminal tab
  run go server
  $ go run *.go
  
open directory /react in terminal tab
  install npm
   $ npm install
  start react app
   $ npm start
 
Use React App Front End 

  Set your user name from react app 

  Add Channel

  Select Channel

  Add Message






