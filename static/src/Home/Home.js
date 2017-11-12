import React, { Component } from 'react'
import { Input, Segment, Form, Header, Card, Button, Select, Icon } from 'semantic-ui-react'

import CustomCard from '../Card/Card'

export default class HomeComponent extends Component {
  handleURLChange = (e, { value }) => this.url = value
  handleCustomIDChange = (e, { value }) => {
    this.customID = value
    fetch("/api/v1/protected/lookup", {
      method: "POST",
      body: JSON.stringify({
        ID: value
      }),
      headers: {
        'Authorization': window.localStorage.getItem('token'),
        'Content-Type': 'application/json'
      }
    }).then(res => res.ok ? res.json() : Promise.reject(res.json()))
      .then(() => {
        this.setState({ showCustomIDError: true })
      })
      .catch(() => this.setState({ showCustomIDError: false }))
  }
  onSettingsChange = (e, { value }) => this.setState({ setOptions: value })

  state = {
    links: [],
    options: [
      { text: 'Custom URL', value: 'custom' },
      { text: 'Expiration', value: 'expire' }
    ],
    setOptions: [],
    showCustomIDError: false
  }
  componentDidMount() {
    this.urlInput.focus()
  }
  handleURLSubmit = () => {
    if (!this.state.showCustomIDError) {
      fetch('/api/v1/protected/create', {
        method: 'POST',
        body: JSON.stringify({
          URL: this.url,
          ID: this.customID
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
  }

  render() {
    const { links, options, setOptions, showCustomIDError } = this.state
    return (
      <div>
        <Segment raised>
          <Header size='huge'>Simplify your links</Header>
          <Form onSubmit={this.handleURLSubmit} autoComplete="off">
            <Form.Field>
              <Input required size='large' type='url' ref={input => this.urlInput = input} onChange={this.handleURLChange} placeholder='Paste a link to shorten it' action>
                <input />
                <Select options={options} placeholder='Settings' onChange={this.onSettingsChange} multiple />
                <Button type='submit'>Shorten<Icon name="arrow right" /></Button>
              </Input>
            </Form.Field>
            <Form.Group widths='equal'>
              {setOptions.indexOf("custom") > -1 && <Form.Field error={showCustomIDError}><Input label={window.location.origin + "/"} onChange={this.handleCustomIDChange} placeholder='my-shortened-url' />
              </Form.Field>
              }
            </Form.Group>

          </Form>
        </Segment>
        <Card.Group itemsPerRow="2">
          {links.map((link, i) => <CustomCard key={i} header={new URL(link[1]).hostname} metaHeader={link[1]} description={link[0]} />)}
        </Card.Group>
      </div >
    )
  }
};