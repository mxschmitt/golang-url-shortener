import React, { Component } from 'react'
import { Segment, Header, Form, Input, Card } from 'semantic-ui-react'

import CustomCard from '../Card/Card'

export default class LookupComponent extends Component {
    state = {
        links: []
    }
    handleURLChange = (e, { value }) => this.url = value
    handleURLSubmit = () => {
        let id = this.url.replace(window.location.origin + "/", "")
        fetch("/api/v1/protected/lookup", {
            method: "POST",
            body: JSON.stringify({
                ID: id
            }),
            headers: {
                'Authorization': window.localStorage.getItem('token'),
                'Content-Type': 'application/json'
            }
        }).then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(res => this.setState({
                links: [...this.state.links, [
                    res.URL,
                    this.url,
                    this.VisitCount,
                    res.CratedOn,
                    res.LastVisit
                ]]
            }))
    }
    render() {
        const { links } = this.state
        return (
            <div>
                <Segment raised>
                    <Header size='huge'>URL Lookup</Header>
                    <Form onSubmit={this.handleURLSubmit} autoComplete="off">
                        <Form.Field>
                            <Input required size='big' ref={input => this.urlInput = input} action={{ icon: 'arrow right', labelPosition: 'right', content: 'Lookup' }} type='url' onChange={this.handleURLChange} name='url' placeholder={window.location.origin + "/..."} />
                        </Form.Field>
                    </Form>
                </Segment>
                <Card.Group itemsPerRow="2">
                    {links.map((link, i) => <CustomCard key={i} header={new URL(link[0]).hostname} metaHeader={link[1]} description={link[0]} showInfoURL/>)}
                </Card.Group>
            </div>
        )
    }
};
