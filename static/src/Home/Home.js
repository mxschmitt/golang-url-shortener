import React, { Component } from 'react'
import { Input, Segment, Form, Header, Card, Icon, Image, Button, Modal } from 'semantic-ui-react'
import { QRCode } from 'react-qr-svg';
import Clipboard from 'react-clipboard.js';

export default class HomeComponent extends Component {
  handleURLChange = (e, { value }) => this.url = value
  state = {
    links: []
  }
  componentDidMount() {
    this.urlInput.focus()
  }
  handleURLSubmit = () => {
    fetch('/api/v1/protected/create', {
      method: 'POST',
      body: JSON.stringify({
        URL: this.url
      }),
      headers: {
        'Authorization': window.localStorage.getItem('token'),
        'Content-Type': 'application/json'
      }
    }).then(res => res.ok ? res.json() : Promise.reject(res.json()))
      .then(r => this.setState({
        links: [...this.state.links, [
          r.URL,
          this.url
        ]]
      }))
  }

  render() {
    const { links } = this.state
    return (
      <div>
        <Segment raised>
          <Header size='huge'>Simplify your links</Header>
          <Form onSubmit={this.handleURLSubmit} autoComplete="off">
            <Form.Field>
              <Input required size='big' ref={input => this.urlInput = input} action={{ icon: 'arrow right', labelPosition: 'right', content: 'Shorten' }} type='url' onChange={this.handleURLChange} name='url' placeholder='Paste a link to shorten it' />
            </Form.Field>
          </Form>
        </Segment>
        <Card.Group itemsPerRow="2">
          {links.map((link, i) => <Card key={i}>
            <Card.Content>
              <Image floated='right' size='mini' src='/assets/images/avatar/large/steve.jpg' />
              <Card.Header>
                {new URL(link[1]).hostname}
              </Card.Header>
              <Card.Meta>
                {link[1]}
              </Card.Meta>
              <Card.Description>
                {link[0]}
              </Card.Description>
            </Card.Content>
            <Card.Content extra>
              <div className='ui two buttons'>
                <Modal closeIcon trigger={<Button icon='qrcode' content='Show QR-Code' />}>
                  <Modal.Header className="ui center aligned">{link[0]}</Modal.Header>
                  <Modal.Content style={{ textAlign: "center" }}>
                    <QRCode style={{ width: "75%" }} value={link[0]} />
                  </Modal.Content>
                </Modal>
                <Clipboard component="button" className="ui button" data-clipboard-text={link[0]} button-title="Copy the Shortened URL to Clipboard">
                  <Icon name="clipboard" />
                  Copy to Clipboard
                </Clipboard>
              </div>
            </Card.Content>
          </Card>)}
        </Card.Group>
      </div >
    )
  }
};