import React, { Component } from 'react'
import { Input, Segment, Form } from 'semantic-ui-react'

import './Home.css';

export default class HomeComponent extends Component {
  handleURLChange = (e, { value }) => this.url = value
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
      .then(r => alert(r.URL))
  }

  render() {
    return (
      <Segment raised>
        <p>Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa strong. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede link mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi.</p>
        <Form onSubmit={this.handleURLSubmit}>
          <Form.Field>
            <Input size='big' action={{ icon: 'arrow right', labelPosition: 'right', content: 'Shorten' }} type='url' onChange={this.handleURLChange} name='url' placeholder='Paste a link to shorten it' />
          </Form.Field>
        </Form>
      </Segment>
    )
  }
};