import React, { Component } from 'react'
import { Container, Input, Segment, Form, Modal, Button } from 'semantic-ui-react'
import './App.css';

class ContainerExampleContainer extends Component {
  handleURLChange = (e, { value }) => this.url = value

  handleURLSubmit() {
    console.log(this.url)
  }

  componentWillMount() {
    console.log("componentWillMount")
  }
  componentDidMount = () => {
    console.log("componentDidMount")
  }

  state = {
    open: true,
    authorized: false
  }

  onOAuthClose = () => {
    this.setState({ open: false })
  }

  onAuthClick = () => {
    console.log("onAuthClick")
    window.open("/api/v1/login", "", "width=600,height=400")
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
                <Input size='big' action={{ icon: 'arrow right' }} type='email' onChange={this.handleURLChange} name='url' placeholder='Enter your long URL here' />
              </Form.Field>
            </Form>
          </Segment>
        </Container>
      )
    } else {
      return (
        <Modal size="tiny" open={open} onClose={this.onOAuthClose}>
          <Modal.Header>
            OAuth2 Authentication
          </Modal.Header>
          <Modal.Content>
            <p>Currently you are only able to use Google as authentification service:</p>
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

export default ContainerExampleContainer;
