import React, { Component } from 'react'
import { Container } from 'semantic-ui-react'
import ClipboardButton from 'react-clipboard.js';
import 'prismjs'
import 'prismjs/components/prism-json'
import PrismCode from 'react-prism'
import 'prismjs/themes/prism.css';

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
                <PrismCode component="pre" className="language-json">
                    {config}
                </PrismCode>
                <ClipboardButton data-clipboard-text={config} className='ui button'>
                    Copy the configuration to the clipboard
                </ClipboardButton>
            </Container>
        )
    }
};
