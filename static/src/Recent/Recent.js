import React, { Component } from 'react'
import { Container, Table, Button, Icon } from 'semantic-ui-react'
import Moment from 'react-moment';
import util from '../util/util'
export default class RecentComponent extends Component {
    state = {
        recent: {}
    }

    componentDidMount() {
        this.loadRecentURLs()
    }

    loadRecentURLs = () => {
        util.getRecentURLs(recent => this.setState({ recent }))
    }

    onRowClick(id) {
        this.props.history.push(`/visitors/${id}`)
    }

    onEntryDeletion(entry) {
        util.deleteEntry(entry.DeletionURL, this.loadRecentURLs)
    }

    render() {
        const { recent } = this.state
        return (
            <Container>
                <Table celled selectable>
                    <Table.Header>
                        <Table.Row>
                            <Table.HeaderCell>Original URL</Table.HeaderCell>
                            <Table.HeaderCell>Created</Table.HeaderCell>
                            <Table.HeaderCell>Short URL</Table.HeaderCell>
                            <Table.HeaderCell>All Clicks</Table.HeaderCell>
                            <Table.HeaderCell>Delete</Table.HeaderCell>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {Object.keys(recent).map(key => <Table.Row key={key} title="Click to view visitor statistics">
                            <Table.Cell onClick={this.onRowClick.bind(this, key)}>{recent[key].Public.URL}</Table.Cell>
                            <Table.Cell onClick={this.onRowClick.bind(this, key)}><Moment>{recent[key].Public.CreatedOn}</Moment></Table.Cell>
                            <Table.Cell>{`${window.location.origin}/${key}`}</Table.Cell>
                            <Table.Cell onClick={this.onRowClick.bind(this, key)}>{recent[key].Public.VisitCount}</Table.Cell>
                            <Table.Cell><Button animated='vertical' onClick={this.onEntryDeletion.bind(this, recent[key])}>
                                <Button.Content hidden>Delete</Button.Content>
                                <Button.Content visible>
                                    <Icon name='trash' />
                                </Button.Content>
                            </Button></Table.Cell>
                        </Table.Row>)}
                    </Table.Body>
                </Table>
            </Container>
        )
    }
}
