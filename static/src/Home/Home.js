import React, { Component } from 'react'
import { Input, Segment, Form, Header, Card, Button, Select, Icon } from 'semantic-ui-react'
import DatePicker from 'react-datepicker';
import moment from 'moment';
import MediaQuery from 'react-responsive';
import 'react-datepicker/dist/react-datepicker.css';
import toastr from 'toastr'

import CustomCard from '../Card/Card'
import './Home.css'

export default class HomeComponent extends Component {
  handleURLChange = (e, { value }) => this.url = value
  handleCustomExpirationChange = expire => this.setState({ expiration: expire })
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
    })
      .then(res => res.ok ? res.json() : Promise.reject(res.json()))
      .then(() => {
        this.setState({ showCustomIDError: true })
      })
      .catch(e => {
        this.setState({ showCustomIDError: false })
        toastr.error(`Could not fetch lookup: ${e}`)
      })
  }
  onSettingsChange = (e, { value }) => this.setState({ setOptions: value })

  state = {
    links: [],
    options: [
      { text: 'Custom URL', value: 'custom' },
      { text: 'Expiration', value: 'expire' }
    ],
    setOptions: [],
    showCustomIDError: false,
    expiration: moment()
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
          ID: this.customID,
          Expiration: this.state.setOptions.indexOf("expire") > -1 ? this.state.expiration.toISOString() : undefined
        }),
        headers: {
          'Authorization': window.localStorage.getItem('token'),
          'Content-Type': 'application/json'
        }
      })
        .then(res => res.ok ? res.json() : Promise.reject(res.json()))
        .then(r => this.setState({
          links: [...this.state.links, [
            r.URL,
            this.url,
            this.state.setOptions.indexOf("expire") > -1 ? this.state.expiration.toISOString() : undefined,
            r.DeletionURL
          ]]
        }))
        .catch(e => toastr.error(`Could not fetch create: ${e}`))
    }
  }

  render() {
    const { links, options, setOptions, showCustomIDError, expiration } = this.state
    return (
      <div>
        <Segment raised>
          <Header size='huge'>Simplify your links</Header>
          <Form onSubmit={this.handleURLSubmit} autoComplete="off">
            <Form.Field>
              <Input required size='large' type='url' ref={input => this.urlInput = input} onChange={this.handleURLChange} placeholder='Paste a link to shorten it' action>
                <input />
                <MediaQuery query="(min-width: 768px)">
                  <Select options={options} placeholder='Settings' onChange={this.onSettingsChange} multiple />
                </MediaQuery>
                <Button type='submit'>Shorten<Icon name="arrow right" /></Button>
              </Input>
            </Form.Field>
            <MediaQuery query="(max-width: 767px)">
              <Form.Field>
                <Select options={options} placeholder='Settings' onChange={this.onSettingsChange} multiple fluid />
              </Form.Field>
            </MediaQuery>
            <Form.Group widths='equal'>
              {setOptions.indexOf("custom") > -1 && <Form.Field error={showCustomIDError}>
                <Input label={window.location.origin + "/"} onChange={this.handleCustomIDChange} placeholder='my-shortened-url' />
              </Form.Field>}
              {setOptions.indexOf("expire") > -1 && <Form.Field>
                <DatePicker showTimeSelect
                  timeFormat="HH:mm"
                  timeIntervals={15}
                  dateFormat="LLL" onChange={this.handleCustomExpirationChange} selected={expiration} customInput={<Input label="Expiration" />} minDate={moment()} />
              </Form.Field>}
            </Form.Group>
          </Form>
        </Segment>
        <Card.Group itemsPerRow="2" stackable style={{ marginTop: "1rem" }}>
          {links.map((link, i) => <CustomCard key={i} header={new URL(link[1]).hostname} expireDate={link[2]} metaHeader={link[1]} description={link[0]} deletionURL={link[3]} />)}
        </Card.Group>
      </div >
    )
  }
};