import React, { Component } from 'react'
import { Segment, Header, Form, Input, Card, Button } from 'semantic-ui-react'

import util from '../util/util'
import CustomCard from '../Card/Card'

export default class LookupComponent extends Component {
    state = {
        links: []
    }
    handleURLChange = (e, { value }) => this.url = value
    handleURLSubmit = () => {
        let id = this.url.replace(window.location.origin + "/", "")
        util.lookupEntry(id, res => this.setState({
            links: [...this.state.links, [
                res.URL,
                this.url,
                this.VisitCount,
                res.CratedOn,
                res.LastVisit,
                res.Expiration
            ]]
        }))
    }
    render() {
        const { links } = this.state
        return (
            <div>
                <Segment raised>
                    <Header size='huge'>URL Lookup</Header>
                    <Form onSubmit={this.handleURLSubmit}>
                        <Form.Field>
                            <Input required size='big' ref={input => this.urlInput = input} action={{ icon: 'arrow right', labelPosition: 'right', content: 'Lookup' }} type='url' onChange={this.handleURLChange} name='url' placeholder={window.location.origin + "/..."} autoComplete="off" />
                        </Form.Field>
                    </Form>
                </Segment>
                <Card.Group itemsPerRow="2">
                    {links.map((link, i) => <CustomCard key={i} header={new URL(link[0]).hostname} metaHeader={link[1]} description={link[0]} expireDate={link[5]} customExtraContent={<div className='ui two buttons'>
                        <Button icon='clock' content='Show recent visitors' />
                        <Button icon='line chart' content='Delete Entry' />
                    </div>} />)}
                </Card.Group>
            </div>
        )
    }
};
