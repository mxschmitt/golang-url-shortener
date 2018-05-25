import React, { Component } from 'react'
import { Container, Button, Icon } from 'semantic-ui-react'
import Moment from 'react-moment';
import ReactTable from 'react-table'
import 'react-table/react-table.css'

import util from '../util/util'

export default class RecentComponent extends Component {
    state = {
        recent: []
    }

    componentDidMount() {
        this.loadRecentURLs()
    }

    loadRecentURLs = () => {
        util.getRecentURLs(recent => {
            let parsed = [];
            for (let key in recent) {
                recent[key].ID = key;
                parsed.push(recent[key]);
            }
            this.setState({ recent: parsed })
        })
    }

    onRowClick(id) {
        this.props.history.push(`/visitors/${id}`)
    }

    onEntryDeletion(deletionURL) {
        util.deleteEntry(deletionURL, this.loadRecentURLs)
    }

    render() {
        const { recent } = this.state

        const columns = [{
            Header: 'Original URL',
            accessor: "Public.URL"
        }, {
            Header: 'Created',
            accessor: 'Public.CreatedOn',
            Cell: props => <Moment fromNow>{props.value}</Moment>
        }, {
            Header: 'Short URL',
            accessor: "ID",
            Cell: props => `${window.location.origin}/${props.value}`
        }, {
            Header: 'Visitor count',
            accessor: "Public.VisitCount"

        }, {
            Header: 'Delete',
            accessor: 'DeletionURL',
            Cell: props => <Button animated='vertical' onClick={this.onEntryDeletion.bind(this, props.value)}>
                <Button.Content hidden>Delete</Button.Content>
                <Button.Content visible>
                    <Icon name='trash' />
                </Button.Content>
            </Button>,
            style: { textAlign: "center" }
        }]

        return (
            <Container>
                <ReactTable data={recent} columns={columns} getTdProps={(state, rowInfo, column, instance) => {
                    return {
                        onClick: (e, handleOriginal) => {
                            if (handleOriginal) {
                                handleOriginal()
                            }
                            if (!rowInfo) {
                                return
                            }
                            if (column.id === "DeletionURL") {
                                return
                            }
                            this.onRowClick(rowInfo.row.ID)
                        }
                    }
                }} />
            </Container>
        )
    }
}
