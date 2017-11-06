import registerServiceWorker from './registerServiceWorker';
import React, { Component } from 'react'
import ReactDOM from 'react-dom';
import {
    HashRouter,
    Route,
    Link
} from 'react-router-dom'
import { Menu, Container } from 'semantic-ui-react'

import About from './About/About'
import Home from './Home/Home'

import 'semantic-ui-css/semantic.min.css';

export default class BaseComponent extends Component {
    state = {}

    handleItemClick = (e, { name }) => this.setState({ activeItem: name })

    render() {
        const { activeItem } = this.state

        return (
            <HashRouter>
                <Container style={{ "margin-top": "15px" }}>
                    <Menu stackable>
                        <Menu.Item to="/">
                            <img src='https://react.semantic-ui.com/logo.png' alt='user profile picture' />
                        </Menu.Item>

                        <Menu.Item name='features' active={activeItem === 'features'} onClick={this.handleItemClick} as={Link} to="/">
                            Shorten
                        </Menu.Item>

                        <Menu.Item name='testimonials' active={activeItem === 'testimonials'} onClick={this.handleItemClick} as={Link} to="/about">
                            About
                        </Menu.Item>

                        <Menu.Item name='sign-in' active={activeItem === 'sign-in'} onClick={this.handleItemClick}>
                            Sign-in
                        </Menu.Item>
                        <Menu.Menu position='right'>
                            <Menu.Item name='logout' active={activeItem === 'logout'} onClick={this.handleItemClick} />
                        </Menu.Menu>
                    </Menu>
                    <Route exact path="/" component={Home} />
                    <Route path="/about" component={About} />
                </Container>
            </HashRouter>
        )
    }
}

ReactDOM.render((
    <BaseComponent />
), document.getElementById('root'))

registerServiceWorker();