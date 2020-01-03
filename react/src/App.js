import React, {Component} from 'react'
import ChannelSection from './components/channels/ChannelSection'
import UserSection from './components/users/UserSection'
import MessageSection from './components/messages/MessageSection'
import Socket from './socket.js'

class App extends Component{
  constructor(props){
    super(props)
    this.state = {
      channels:[],
      users:[],
      messages:[],
      activeChannel:{},
      connected: false
    }
  }

  componentDidMount() {
    let ws = new WebSocket('ws://localhost:4000')
    //instantiate new socket to make connection
    let socket = this.socket = new Socket(ws)
    //to listen for server messages, pass in event name to listen for and function or event handler to call when message received
    socket.on('connect', this.onConnect)
    socket.on('disconnect', this.onDisconnect)
    socket.on('channel add', this.onAddChannel)
    socket.on('user add', this.onAddUser)
    socket.on('user edit',this.onEditUser)
    socket.on('user remove', this.onRemoveUser)
    socket.on('message add', this.onMessageAdd)
  }

  onMessageAdd = (message) => {
    let {messages} = this.state
    messages.push(message)
    this.setState({messages})
  }

  onRemoveUser = (removeUser) => {
    let {users} = this.state
    users = users.filter(user => {
      return user.id !== removeUser.id
    })
    this.setState({users})
  }

  onAddUser = (user) => {
    let {users} = this.state
    users.push(user)
    this.setState({users})
  }

  onEditUser = (editUser) => {
    let {users} = this.state
    users = users.map(user => {
      if(editUser.id === user.id){
        return editUser
      }
      return user
    })
    this.setState({users})
  }

  onConnect = () => {
    this.setState({connected:true})
    //to send message to server with event name
    this.socket.emit('channel subscribe')
    this.socket.emit('user subscribe')
  }

  onDisconnect = () => {
    this.setState({connected:false})
  }

  onAddChannel = (channel) => {
    let {channels} = this.state
    channels.push(channel)
    this.setState({channels})
  }

  addChannel = (name) => {
    //to send message to server with event name and payload
    this.socket.emit('channel add', {name})
  }

  setChannel = (activeChannel) => {
    this.setState({activeChannel})
    this.socket.emit('message unsubscribe')
    this.setState({messages:[]})
    this.socket.emit('message subscribe',
    {channelId: activeChannel.id})
  }

  setUserName = (name) => {
    this.socket.emit('user edit', {name})
  }

  addMessage = (body) => {
    let {activeChannel} = this.state
    this.socket.emit('message add',
    {channelId: activeChannel.id,body})
  }

  render(){
    return (
      <div className='app'>
        <div className="nav">
          <ChannelSection
            {...this.state}
            addChannel={this.addChannel}
            setChannel={this.setChannel}
          />
          <UserSection
            {...this.state}
            setUserName={this.setUserName}
          />
        </div>
          <MessageSection
            {...this.state}
            addMessage={this.addMessage}
          />
     </div>
    )
  }
}

export default App;
