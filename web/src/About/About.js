import React, { Component } from 'react'
import { Container, Table } from 'semantic-ui-react'
import Moment from 'react-moment';

export default class AboutComponent extends Component {
    state = {
        info: null
    }

    componentWillMount() {
        this.setState({ info: this.props.info })
    }

    render() {
        const { info } = this.state
        return (
            <Container >
                {info && <Table celled>
                    <Table.Header>
                        <Table.Row>
                            <Table.HeaderCell>Property</Table.HeaderCell>
                            <Table.HeaderCell>Info</Table.HeaderCell>
                        </Table.Row>
                    </Table.Header>

                    <Table.Body>
                        <Table.Row>
                            <Table.Cell>Source Code</Table.Cell>
                            <Table.Cell><a href="https://github.com/mxschmitt/golang-url-shortener" target="_blank" rel="noopener noreferrer">github.com/mxschmitt/golang-url-shortener</a></Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Author</Table.Cell>
                            <Table.Cell><a href="https://github.com/mxschmitt/" target="_blank" rel="noopener noreferrer">Max Schmitt</a></Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Compilation Time</Table.Cell>
                            <Table.Cell><Moment fromNow>{info.compilationTime}</Moment> - <Moment>{info.compilationTime}</Moment></Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Commit Hash</Table.Cell>
                            <Table.Cell><a href={"https://github.com/mxschmitt/golang-url-shortener/commit/" + info.commit} target="_blank" rel="noopener noreferrer">{info.commit}</a></Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Go Version</Table.Cell>
                            <Table.Cell>{info.go}</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>React Version</Table.Cell>
                            <Table.Cell>{React.version}</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Node.js Version</Table.Cell>
                            <Table.Cell>{info.nodeJS}</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Yarn Version</Table.Cell>
                            <Table.Cell>{info.yarn}</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>License</Table.Cell>
                            <Table.Cell><a href="https://github.com/mxschmitt/golang-url-shortener/blob/master/LICENSE.md" target="_blank" rel="noopener noreferrer">MIT</a></Table.Cell>
                        </Table.Row>
                    </Table.Body>
                </Table>}
            </Container>
        )
    }
}
