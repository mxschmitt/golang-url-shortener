import React, { Component } from 'react'
import { Container, Table } from 'semantic-ui-react'
import moment from 'moment'

export default class AboutComponent extends Component {
    state = {
        info: null
    }

    componentDidMount() {
       this.setState({ info: this.props.location.state.info })
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
                            <Table.Cell><a href="https://github.com/maxibanki/golang-url-shortener" target="_blank" rel="noopener noreferrer">github.com/maxibanki/golang-url-shortener</a></Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Author</Table.Cell>
                            <Table.Cell><a href="https://github.com/maxibanki/" target="_blank" rel="noopener noreferrer">Max Schmitt</a></Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Compilation Time</Table.Cell>
                            <Table.Cell>{moment(info.compilationTime).fromNow()} ({info.compilationTime})</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Commit Hash</Table.Cell>
                            <Table.Cell><a href={"https://github.com/maxibanki/golang-url-shortener/commit/" + info.commit} target="_blank" rel="noopener noreferrer">{info.commit}</a></Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Go Version</Table.Cell>
                            <Table.Cell>{info.go.replace("go", "")}</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>React Version</Table.Cell>
                            <Table.Cell>{React.version}</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Node.js Version</Table.Cell>
                            <Table.Cell>{info.nodeJS.replace("v", "")}</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>Yarn Version</Table.Cell>
                            <Table.Cell>{info.yarn}</Table.Cell>
                        </Table.Row>
                        <Table.Row>
                            <Table.Cell>License</Table.Cell>
                            <Table.Cell><a href="https://github.com/maxibanki/golang-url-shortener/blob/master/LICENSE.md" target="_blank" rel="noopener noreferrer">MIT</a></Table.Cell>
                        </Table.Row>
                    </Table.Body>
                </Table>}
            </Container>
        )
    }
};
