import React, { Component } from 'react'
import { Container, Table } from 'semantic-ui-react'
import Moment from 'react-moment';

import util from '../util/util'
export default class VisitorComponent extends Component {
    state = {
        id: "",
        entry: null,
        visitors: []
    }

    componentWillMount() {
        this.setState({ id: this.props.match.params.id })
        util.lookupEntry(this.props.match.params.id, entry => this.setState({ entry }))
        this.reloadVisitors()
        this.reloadInterval = setInterval(this.reloadVisitors, 1000)
    }

    componentWillUnmount() {
        clearInterval(this.reloadInterval)
    }

    reloadVisitors = () => {
        util.getVisitors(this.props.match.params.id, visitors => this.setState({ visitors }))
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
        const { visitors, id, entry } = this.state
        return (
            <Container >
                {entry && <p>
                    Entry with id '{id}' was created at <Moment>{entry.CreatedOn}</Moment> and redirects to '{entry.URL}'. Currently it has {visitors.length} visits.
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
                            <Table.Cell><Moment>{visit.Timestamp}</Moment></Table.Cell>
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
