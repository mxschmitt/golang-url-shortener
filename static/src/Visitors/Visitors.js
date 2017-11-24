import React, { Component } from 'react'
import { Container, Table } from 'semantic-ui-react'
import moment from 'moment'
import toastr from 'toastr'

export default class VisitorComponent extends Component {
    state = {
        id: "",
        visitors: null
    }

    componentDidMount() {
        this.setState({ id: this.props.match.params.id })
        fetch('/api/v1/protected/visitors', {
            method: 'POST',
            body: JSON.stringify({
                ID: this.props.match.params.id
            }),
            headers: {
                'Authorization': window.localStorage.getItem('token'),
                'Content-Type': 'application/json'
            }
        })
            .then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(visitors => this.setState({ visitors }))
            .catch(e => e.done(res => toastr.error(`Could not fetch visitors: ${res}`)))
    }

    // getUTMSource is a function which generates the output for the utm[...] table column
    getUTMSource(visit) {
        return [visit.UTMSource, visit.UTMMedium, visit.UTMCampaign, visit.UTMContent, visit.UTMTerm]
            .map(value => value ? value : null)
            .filter(v => v)
            .map((value, idx, data) => value + (idx !== data.length - 1 ? ", " : ""))
            .join("")
    }

    render() {
        const { visitors } = this.state
        return (
            <Container >
                <Table celled>
                    <Table.Header>
                        <Table.Row>
                            <Table.HeaderCell>Timestamp</Table.HeaderCell>
                            <Table.HeaderCell>IP</Table.HeaderCell>
                            <Table.HeaderCell>Referer</Table.HeaderCell>
                            <Table.HeaderCell>UTM (source, medium, campaign, content, term)</Table.HeaderCell>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {visitors && visitors.map((visit, idx) => <Table.Row key={idx}>
                            <Table.Cell>{moment(visit.Timestamp).format('LLL')}</Table.Cell>
                            <Table.Cell>{visit.IP}</Table.Cell>
                            <Table.Cell>{visit.Referer}</Table.Cell>
                            <Table.Cell>{this.getUTMSource(visit)}</Table.Cell>
                        </Table.Row>)}
                    </Table.Body>
                </Table>
            </Container>
        )
    }
};
