import React, { Component } from 'react'
import { Container, Table } from 'semantic-ui-react'
import toastr from 'toastr'
import moment from 'moment'

export default class RecentComponent extends Component {
    state = {
        recent: null
    }

    componentWillMount() {
        fetch('/api/v1/protected/recent', {
            method: 'POST',
            headers: {
                'Authorization': window.localStorage.getItem('token'),
            }
        })
            .then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(recent => this.setState({ recent: recent }))
            .catch(e => e.done(res => toastr.error(`Could not fetch recent shortened URLs: ${res}`)))
    }

    onRowClick(id) {
        this.props.history.push(`/visitors/${id}`)
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
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {recent && Object.keys(recent).map(key => <Table.Row key={key} title="Click to view visitor statistics" onClick={this.onRowClick.bind(this, key)}>
                            <Table.Cell>{recent[key].Public.URL}</Table.Cell>
                            <Table.Cell>{moment(recent[key].Public.CreatedOn).format('LLL')}</Table.Cell>
                            <Table.Cell>{`${window.location.origin}/${key}`}</Table.Cell>
                            <Table.Cell>{recent[key].Public.VisitCount}</Table.Cell>
                        </Table.Row>)}
                    </Table.Body>
                </Table>
            </Container>
        )
    }
};
