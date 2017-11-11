import React, { Component } from 'react'
import { Input, Segment, Form, Header, Card } from 'semantic-ui-react'

import CustomCard from '../Card/Card'

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
          {links.map((link, i) => <CustomCard key={i} header={new URL(link[1]).hostname} metaHeader={link[1]} description={link[0]} />)}
        </Card.Group>
      </div >
    )
  }
};