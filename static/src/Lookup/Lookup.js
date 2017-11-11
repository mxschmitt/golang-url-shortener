import React, { Component } from 'react'
import { Segment, Header, Form, Input } from 'semantic-ui-react'

export default class LookupComponent extends Component {
    render() {
        return (
            <Segment raised>
                <Header size='huge'>URL Lookup</Header>
                <Form onSubmit={this.handleURLSubmit} autoComplete="off">
                    <Form.Field>
                        <Input required size='big' ref={input => this.urlInput = input} action={{ icon: 'arrow right', labelPosition: 'right', content: 'Lookup' }} type='url' onChange={this.handleURLChange} name='url' placeholder={window.location.origin+"/..."} />
                    </Form.Field>
                </Form>
            </Segment>
        )
    }
};
