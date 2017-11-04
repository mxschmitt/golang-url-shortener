import React, { Component } from 'react'
import { Container, Input, Segment, Form, Modal, Button } from 'semantic-ui-react'
import './App.css';

class AppComponent extends Component {
  handleURLChange = (e, { value }) => this.url = value
  handleURLSubmit = () => {
    console.log("handle Submit", "URL:", this.url)
    fetch("/api/v1/protected/create", {
      method: "POST",
      body: JSON.stringify({
        URL: this.url
      }),
      headers: {
        "Authorization": window.localStorage.getItem("token"),
        'Content-Type': 'application/json'
      }
    }).then(res => res.ok ? res.json() : Promise.reject(res.json()))
      .then(d => console.log(d))
  }

  componentWillMount() {
    this.checkAuth()
  }

  state = {
    open: true,
    userData: {},
    authorized: false
  }

  onOAuthClose() {
    this.setState({ open: true })
  }
  checkAuth = () => {
    const that = this,
      token = window.localStorage.getItem("token");
    if (token) {
      fetch("/api/v1/check", {
        method: "POST",
        body: JSON.stringify({
          Token: token
        }),
        headers: {
          'Content-Type': 'application/json'
        }
      }).then(res => res.ok ? res.json() : Promise.reject(res.json())) // Check if the request was StatusOK, otherwise reject Promise
        .then(d => {
          that.setState({ userData: d })
          that.setState({ authorized: true })
        })
        .catch(e => {
          window.localStorage.removeItem("token");
          that.setState({ authorized: false })
        })
    }
  }
  onAuthCallback = data => {
    // clear the old event listener, so that the event can only emitted be once
    window.removeEventListener('onAuthCallback', this.onAuthCallback);
    window.localStorage.setItem("token", data.detail.token);
    this.checkAuth();
  }
  onAuthClick = () => {
    window.addEventListener('onAuthCallback', this.onAuthCallback, false);
    // Open the oAuth window that is it centered in the middle of the screen
    var wwidth = 400,
      wHeight = 500;
    var wLeft = (window.screen.width / 2) - (wwidth / 2);
    var wTop = (window.screen.height / 2) - (wHeight / 2);
    window.open("/api/v1/login", "", `width=${wwidth}, height=${wHeight}, top=${wTop}, left=${wLeft}`)
  }

  render() {
    const { open, authorized } = this.state
    if (authorized) {
      return (
        <Container id='rootContainer' >
          <Segment raised>
            <p>Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa strong. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede link mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi.</p>
            <Form onSubmit={this.handleURLSubmit}>
              <Form.Field>
                <Input size='big' action={{ icon: 'arrow right', labelPosition: 'right', content: 'Shorten' }} type='url' onChange={this.handleURLChange} name='url' placeholder='Paste a link to shorten it' />
              </Form.Field>
            </Form>
          </Segment>
        </Container>
      )
    } else {
      return (
        <Modal size="tiny" open={open} onClose={this.onOAuthClose}>
          <Modal.Header>
            Authentication
          </Modal.Header>
          <Modal.Content>
            <p>Currently you are only able to use Google as authentication service:</p>
            <div className="ui center aligned segment">
              <Button className="ui google plus button" onClick={this.onAuthClick}>
                <i className="google icon"></i>
                Login with Google
            </Button>
            </div>
          </Modal.Content>
        </Modal>
      )
    }
  }
}

export default AppComponent;
