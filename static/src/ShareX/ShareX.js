import React, { Component } from 'react'
import { Container } from 'semantic-ui-react'
import { Highlight } from 'react-fast-highlight';
import ClipboardButton from 'react-clipboard.js';
import 'highlight.js/styles/github.css'

export default class ShareXComponent extends Component {
    state = {
        config: JSON.stringify({
            Name: "Golang URL Shortener",
            DestinationType: "URLShortener",
            RequestType: "POST",
            RequestURL: window.location.origin + "/api/v1/protected/create",
            Arguments: {
                URL: "$input$"
            },
            Headers: {
                Authorization: window.localStorage.getItem('token')
            },
            ResponseType: "Text",
            URL: "$json:URL$"
        }, null, 4)
    }

    render() {
        const { config } = this.state
        return (
            <Container id='rootContainer' >
                <div>ShareX</div>
                <Highlight languages={['json']}>
                    {config}
                </Highlight>
                <ClipboardButton data-clipboard-text={config} className='ui button'>
                    Copy the configuration to the clipboard
                </ClipboardButton>
            </Container>
        )
    }
};
