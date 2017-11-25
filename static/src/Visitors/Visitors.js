import React, { Component } from 'react'
import { Container, Table } from 'semantic-ui-react'
import moment from 'moment'
import toastr from 'toastr'

export default class VisitorComponent extends Component {
    state = {
        visitors: [],
        info: null
    }

    componentDidMount() {
        this.setState({ id: this.props.match.params.id })
        fetch("/api/v1/protected/lookup", {
            method: "POST",
            body: JSON.stringify({
                ID: this.props.match.params.id
            }),
            headers: {
                'Authorization': window.localStorage.getItem('token'),
                'Content-Type': 'application/json'
            }
        })
            .then(res => res.ok ? res.json() : Promise.reject(res.json()))
            .then(info => this.setState({ info }))
            .catch(e => {
                toastr.error(`Could not fetch lookup: ${e}`)
            })
        this.loop = setInterval(() => {
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
        }, 1000)
    }

    componentWillUnmount() {
        clearInterval(this.loop)
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
        const { visitors, id, info } = this.state
        return (
            <Container >
                {info && <p>
                    Entry with id {id} was created at {moment(info.CreatedOn).format('LLL')} and redirects to '{info.URL}'. Currently it has {visitors.length} visits.
                </p>}
                <Table celled>
                    <Table.Header>
                        <Table.Row>
                            <Table.HeaderCell>Timestamp</Table.HeaderCell>
                            <Table.HeaderCell>IP</Table.HeaderCell>
                            <Table.HeaderCell>User Agent</Table.HeaderCell>
                            <Table.HeaderCell>Referer</Table.HeaderCell>
                            <Table.HeaderCell>UTM (source, medium, campaign, content, term)</Table.HeaderCell>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {visitors && visitors.map((visit, idx) => <Table.Row key={idx}>
                            <Table.Cell>{moment(visit.Timestamp).format('LLL')}</Table.Cell>
                            <Table.Cell>{visit.IP}</Table.Cell>
                            <Table.Cell>{visit.UserAgent}</Table.Cell>
                            <Table.Cell>{visit.Referer}</Table.Cell>
                            <Table.Cell>{this.getUTMSource(visit)}</Table.Cell>
                        </Table.Row>)}
                    </Table.Body>
                </Table>
            </Container>
        )
    }
};
