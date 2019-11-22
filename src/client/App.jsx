
import React from 'react'
import { Component } from 'react';
import './app.scss';
import ComponentSocket from './components/component-socket.jsx'
import SystemInfo from './components/systeminfo/systeminfo.jsx'
import Clock from './components/clock/clock';
export default class App extends Component {

state = {
  onmessage : (event) => {
    console.log("message here:"+event.data);
    const data = JSON.parse(event.data);
    var stateUpdate = {}
    stateUpdate[data.id]=data;
    this.setState(stateUpdate);
  }
};

componentWillmount() {
  this.setState({simple:"Not gotten"});
  }
  componentDidMount() {
  }

  render() {
    return (
      <div>
        <ComponentSocket url="ws://localhost:8080/ws" onmessage={this.state.onmessage}/>
        <Clock id="clock" message={this.state.clock}/>
        <SystemInfo id="systeminfo" message={this.state.systeminfo}/>
        <SystemInfo id="systeminfo2" message={this.state.systeminfo2}/>
      </div>
    );
  }
}
