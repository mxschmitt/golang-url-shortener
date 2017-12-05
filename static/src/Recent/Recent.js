import React, { Component } from 'react'
import { Container, Table, Button, Icon } from 'semantic-ui-react'
import toastr from 'toastr'
import Moment from 'react-moment';
import util from '../util/util'
export default class RecentComponent extends Component {
    state = {
        recent: null
    }

    componentDidMount() {
        this.loadRecentURLs()
    }

    loadRecentURLs() {
        fetch('/api/v1/protected/recent', {
            method: 'POST',
            headers: {
                'Authorization': window.localStorage.getItem('token'),
            }
        })
            .then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(recent => this.setState({ recent: recent }))
            .catch(e => e instanceof Promise ? e.then(error => toastr.error(`Could load recent URLs: ${error.error}`)) : null)
    }

    onRowClick(id) {
        this.props.history.push(`/visitors/${id}`)
    }

    onEntryDeletion(entry) {
        util.deleteEntry(entry.DeletionURL)
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
                        {recent && Object.keys(recent).map(key => <Table.Row key={key} title="Click to view visitor statistics">
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
};
