import React, { Component } from 'react'
import { Container, Button, Icon } from 'semantic-ui-react'
import Moment from 'react-moment';
import ReactTable from 'react-table'
import 'react-table/react-table.css'

import util from '../util/util'

export default class AllEntriesComponent extends Component {
    state = {
        allEntries: [],
        displayURL: window.location.origin
    }

    componentDidMount() {
        this.getAllURLs()
        fetch("/displayurl")
        .then(response => response.json())
        .then(data => this.setState({displayURL: data}));
    }

    getAllURLs = () => {
        util.getAllURLs(allEntries => {
            let parsed = [];
            for (let key in allEntries) {
                if ({}.hasOwnProperty.call(allEntries, key)) {
                    allEntries[key].ID = key;
                    parsed.push(allEntries[key]);
                }
            }
            this.setState({ allEntries: parsed })
        })
    }

    onRowClick(id) {
        this.props.history.push(`/visitors/${id}`)
    }

    onEntryDeletion(deletionURL) {
        util.deleteEntry(deletionURL, this.getAllURLs)
    }

    render() {
        const { allEntries } = this.state

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
            Cell: props => `${this.state.displayURL}/${props.value}`
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
                <ReactTable data={allEntries} columns={columns} getTdProps={(state, rowInfo, column, instance) => {
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
