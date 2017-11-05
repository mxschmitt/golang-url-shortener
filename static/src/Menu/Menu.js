import React, { Component } from 'react'
import { Container, Menu, Input } from 'semantic-ui-react'
import Home from '../App/App'

import React from 'react'
import { render } from 'react-dom'

// First we import some modules...
import { Router, Route, IndexRoute, Link, hashHistory } from 'react-router'

// Make a new component to render inside of Inbox
const Message = React.createClass({
    render() {
        return <h3>Message</h3>
    }
})

const Inbox = React.createClass({
    render() {
        return (
            <div>
                <h2>Inbox</h2>
            </div>
        )
    }
})



class MenuComponent extends Component {
    componentWillMount() {

    }

    state = {
        activeItem: 'home',
        history: null
    }

    handleItemClick = (e, { name }) => this.setState({ activeItem: name })

    render() {
        const { activeItem, history } = this.state
        let currentItem;
        switch (activeItem) {
            case 'home':
                currentItem = <Home />
                break;
        }
        return (
            // <Container id='rootContainer' >
            //     <Menu secondary >
            //         <Menu.Item name='home' active={activeItem === 'home'} onClick={this.handleItemClick} />
            //         <Menu.Item name='messages' active={activeItem === 'messages'} onClick={this.handleItemClick} />
            //         <Menu.Item name='friends' active={activeItem === 'friends'} onClick={this.handleItemClick} />
            //         <Menu.Menu position='right'>
            //             <Menu.Item>
            //                 <Input icon='search' placeholder='Search...' />
            //             </Menu.Item>
            //             <Menu.Item name='logout' active={activeItem === 'logout'} onClick={this.handleItemClick} />
            //         </Menu.Menu>
            //     </Menu>
            //     {{ currentItem }}
            // </Container>
            <Router history={history}>
                <Route path="/" component={App}>
                    <IndexRoute component={Home} />
                    <Route path="about" component={About} />
                    <Route path="inbox" component={Inbox}>
                        {/* add some nested routes where we want the UI to nest */}
                        {/* render the stats page when at `/inbox` */}
                        <IndexRoute component={InboxStats} />
                        {/* render the message component at /inbox/messages/123 */}
                        <Route path="messages/:id" component={Message} />
                    </Route>
                </Route>
            </Router>
        )
    }
}

export default MenuComponent;
