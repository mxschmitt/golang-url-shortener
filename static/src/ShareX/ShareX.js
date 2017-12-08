import React, { Component } from 'react'
import { Container, Image, Modal, Button, Icon } from 'semantic-ui-react'
import Clipboard from 'react-clipboard.js';
import 'prismjs'
import 'prismjs/components/prism-json'
import PrismCode from 'react-prism'
import 'prismjs/themes/prism.css';

import './ShareX.css'

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
            URL: "$json:URL$",
            DeletionURL: "$json:DeletionURL$"
        }, null, 4),
        currentStep: 0,
        availableSteps: [
            {
                image: "open-destination-settings.png",
                text: "To add the correct configuration for the custom URL shortener in ShareX, open first the destination settings."
            },
            {
                image: "import-url-shortener.png",
                text: "After you scroll down in the left sidebar, select custom uploaders. There you can import it from the clipboard."
            },
            {
                image: "set-default-shortener.png",
                text: "Now it is successfully added to ShareX. To set it as the default shortener, select Custom URL shortener in the dropdown menu."
            },
        ]
    }

    goBackwards = () => this.setState({ currentStep: this.state.currentStep - 1 })

    goForwards = () => this.setState({ currentStep: this.state.currentStep + 1 })

    render() {
        const { config, currentStep, availableSteps } = this.state
        return (
            <Container id='rootContainer' >
                <div>On this page you see information about the ShareX integration and how you configure it. If you haven't ShareX installed, you can download it from <a href='https://getsharex.com/'>here</a>.</div>
                <PrismCode component="pre" className="language-json">
                    {config}
                </PrismCode>
                <Modal closeIcon trigger={
                    <div className="ui center aligned segment">
                        <Clipboard data-clipboard-text={config} className='ui button' onClick={this.onClipboardButtonClick}>
                            Copy the configuration and start the ShareX setup
                        </Clipboard >
                    </div>
                }>
                    <Modal.Header>Setting up ShareX - Step {currentStep + 1}</Modal.Header>
                    <Modal.Content>
                        <p>{availableSteps[currentStep].text}</p>
                        <Image src={'images/setting-up-sharex/' + availableSteps[currentStep].image} rounded />
                    </Modal.Content>
                    <Modal.Actions>
                        {currentStep - 1 >= 0 &&
                            <Button onClick={this.goBackwards}>
                                <Icon name='step backward' /> Backwards
                            </Button>
                        }
                        {currentStep + 1 < availableSteps.length &&
                            <Button onClick={this.goForwards} >
                                <Icon name='step forward' />Forwards
                            </Button>
                        }
                    </Modal.Actions>
                </Modal>
            </Container>
        )
    }
}
