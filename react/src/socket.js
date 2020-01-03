import { EventEmitter } from 'events'
//using socket interface expose simpler interface to app to handle websocket messaging
//socket class handles creating and sending message
//EventEmitter is wrapper around websocket connection acting as event emitter between server and client
//ee included in nodejs. module provides pattern for event based message passing
class Socket {
  constructor(ws = new WebSocket(), ee = new EventEmitter()){
    this.ws = ws
    this.ee = ee
    ws.onmessage = this.message
    ws.onopen = this.open 
    ws.onclose = this.close 
  }

  on = (name,fn) => {
    this.ee.on(name, fn)
    this.ee.on('error', (err) => {
      console.log('there was an error')
    })
  }

  off = (name,fn) => {
    this.ee.removeListener(name, fn)
  }

  emit = (name,data) => {
    const message = JSON.stringify({name,data})
    this.ws.send(message)
  }

  //parse data and emit name and payload to call any event listeners added in our app component
  //if error parsing message, emit error event
  message = (e) => {
    try{
      const message = JSON.parse(e.data)
      console.log(message)
    //send message by passing in event name(ex.'channel add') and payload(ex.'hello')
      this.ee.emit(message.name, message.data)
    }
    catch(err) {
      this.ee.emit('error', err)
    }
  }
    
  open = () => {
    this.ee.emit('connect')
  }
  close = () => {
    this.ee.emit('disconnect')
  }
    
}

export default Socket