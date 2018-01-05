import React, { Component } from 'react'
import { Container } from 'semantic-ui-react'
import Moment from 'react-moment';
import ReactTable from 'react-table'
import 'react-table/react-table.css'

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

        const columns = [{
            Header: 'Timestamp',
            accessor: 'Timestamp',
            Cell: props => <Moment>{props.value}</Moment>
        }, {
            Header: 'IP',
            accessor: "IP"
        }, {
            Header: 'User Agent',
            accessor: "UserAgent"
        }, {
            Header: 'Referer',
            accessor: "Referer"
        }, {
            Header: 'UTM (source, medium, campaign, content, term)',
            Cell: props => this.getUTMSource(props.original)
        }]

        return (
            <Container >
                {entry && <p>
                    Entry with id '{id}' was created at <Moment>{entry.CreatedOn}</Moment> and redirects to '{entry.URL}'. Currently it has {visitors.length} visits.
                </p>}
                <ReactTable data={visitors} columns={columns} />
            </Container>
        )
    }
}
