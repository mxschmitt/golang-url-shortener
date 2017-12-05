import React, { Component } from 'react'
import { Card, Icon, Button, Modal } from 'semantic-ui-react'
import { QRCode } from 'react-qr-svg';
import Clipboard from 'react-clipboard.js';
import toastr from 'toastr'
import Moment from 'react-moment';

export default class CardComponent extends Component {
    onDeletonLinkCopy() {
        toastr.info('Copied the deletion URL to the Clipboard')
    }
    onShortedURLSuccess() {
        toastr.info('Copied the shorted URL to the Clipboard')
    }
    render() {
        return (<Card key={this.key}>
            <Card.Content>
                {this.props.expireDate && <Card.Header style={{ float: "right", fontSize: "1.1em" }}>
                    Expires in <Moment fromNow>{this.props.expireDate}</Moment>
                </Card.Header>}
                <Card.Header>
                    {this.props.header}
                </Card.Header>
                <Card.Meta>
                    {this.props.metaHeader}
                </Card.Meta>
                <Card.Description>
                    {this.props.description} 
                    {this.props.deletionURL && <Clipboard component="i" className="trash link icon" style={{ float: "right" }} button-title="Copy the deletion URL to the clipboard" data-clipboard-text={this.props.deletionURL} onSuccess={this.onDeletonLinkCopy} />}
                </Card.Description>
            </Card.Content>
            <Card.Content extra>
                {!this.props.customExtraContent ? <div className='ui two buttons'>
                    <Modal closeIcon trigger={<Button icon='qrcode' content='Show QR-Code' />}>
                        <Modal.Header className='ui center aligned'>{this.props.description}</Modal.Header>
                        <Modal.Content style={{ textAlign: 'center' }}>
                            <QRCode style={{ width: '75%' }} value={this.props.description} />
                        </Modal.Content>
                    </Modal>
                    <Clipboard component='button' className='ui button' data-clipboard-text={this.props.description} onSuccess={this.onShortedURLSuccess} button-title='Copy the Shortened URL to the Clipboard'>
                        <div>
                            <Icon name='clipboard' />
                            Copy to Clipboard
                        </div>
                    </Clipboard>
                </div> : this.props.customExtraContent}
            </Card.Content>
        </Card>)
    }
};