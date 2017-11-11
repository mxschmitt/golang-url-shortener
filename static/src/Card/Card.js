import React, { Component } from 'react'
import { Card, Icon, Button, Modal } from 'semantic-ui-react'
import { QRCode } from 'react-qr-svg';
import Clipboard from 'react-clipboard.js';

export default class CardComponent extends Component {
    render() {
        return (<Card key={this.key}>
            <Card.Content>
                <Card.Header>
                    {this.props.header}
                </Card.Header>
                <Card.Meta>
                    {this.props.metaHeader}
                </Card.Meta>
                <Card.Description>
                    {this.props.description}
                </Card.Description>
            </Card.Content>
            <Card.Content extra>
                {!this.props.showInfoURL ? <div className='ui two buttons'>
                    <Modal closeIcon trigger={<Button icon='qrcode' content='Show QR-Code' />}>
                        <Modal.Header className="ui center aligned">{this.props.description}</Modal.Header>
                        <Modal.Content style={{ textAlign: "center" }}>
                            <QRCode style={{ width: "75%" }} value={this.props.description} />
                        </Modal.Content>
                    </Modal>
                    <Clipboard component="button" className="ui button" data-clipboard-text={this.props.description} button-title="Copy the Shortened URL to the Clipboard">
                        <div>
                            <Icon name="clipboard" />
                            Copy to Clipboard
                        </div>
                    </Clipboard>
                </div> : <div className='ui two buttons'>
                        <Button icon='line chart' content='Show live tracking' />
                        <Button icon='clock' content='Show recent visitors' />
                    </div>}
            </Card.Content>
        </Card>)
    }
};