import React, { Component } from 'react'
import ReactDOM from 'react-dom';
import { HashRouter, Route, Link } from 'react-router-dom'
import { Menu, Container, Modal, Button, Image } from 'semantic-ui-react'
import 'semantic-ui-css/semantic.min.css';

import About from './About/About'
import Home from './Home/Home'
import ShareX from './ShareX/ShareX'
import Lookup from './Lookup/Lookup'

export default class BaseComponent extends Component {
    state = {
        open: true,
        userData: {},
        authorized: false,
        activeItem: ""
    }

    onOAuthClose() {
        this.setState({ open: true })
    }

    handleItemClick = (e, { name }) => this.setState({ activeItem: name })

    componentWillMount() {
        this.checkAuth()
    }

    checkAuth = () => {
        const that = this,
            token = window.localStorage.getItem('token');
        if (token) {
            fetch('/api/v1/check', {
                method: 'POST',
                body: JSON.stringify({
                    Token: token
                }),
                headers: {
                    'Content-Type': 'application/json'
                }
            }).then(res => res.ok ? res.json() : Promise.reject(res.json())) // Check if the request was StatusOK, otherwise reject Promise
                .then(d => {
                    that.setState({ userData: d })
                    that.setState({ authorized: true })
                })
                .catch(e => {
                    window.localStorage.removeItem('token');
                    that.setState({ authorized: false })
                })
        }
    }

    onAuthCallback = data => {
        if (data.isTrusted) {
            // clear the old event listener, so that the event can only emitted be once
            window.removeEventListener('message', this.onAuthCallback);
            window.localStorage.setItem('token', data.data);
            this.checkAuth();
        }
    }
    onAuthClick = () => {
        window.addEventListener('message', this.onAuthCallback, false);
        // Open the oAuth window that is it centered in the middle of the screen
        var wwidth = 400,
            wHeight = 500;
        var wLeft = (window.screen.width / 2) - (wwidth / 2);
        var wTop = (window.screen.height / 2) - (wHeight / 2);
        window.open('/api/v1/login', '', `width=${wwidth}, height=${wHeight}, top=${wTop}, left=${wLeft}`)
    }

    handleLogout = () => {
        window.localStorage.removeItem("token")
        this.setState({ authorized: false })
    }

    render() {
        const { open, authorized, activeItem, userData } = this.state
        if (!authorized) {
            return (
                <Modal size='tiny' open={open} onClose={this.onOAuthClose}>
                    <Modal.Header>
                        Authentication
                    </Modal.Header>
                    <Modal.Content>
                        <p>Currently you are only able to use Google as authentication service:</p>
                        <div className='ui center aligned segment'>
                            <Button className='ui google plus button' onClick={this.onAuthClick}>
                                <i className='google icon'></i>
                                Login with Google
                            </Button>
                        </div>
                    </Modal.Content>
                </Modal>
            )
        }
        return (
            <HashRouter>
                <Container style={{ paddingTop: "15px" }}>
                    <Menu stackable>
                        <Menu.Item as={Link} to="/" name='shorten' onClick={this.handleItemClick} >
                            <Image src={userData.Picture} alt='user profile' circular size='mini' />
                        </Menu.Item>
                        <Menu.Item name='shorten' active={activeItem === 'shorten'} onClick={this.handleItemClick} as={Link} to="/">
                            Shorten
                        </Menu.Item>
                        <Menu.Item name='ShareX' active={activeItem === 'ShareX'} onClick={this.handleItemClick} as={Link} to="/sharex">
                            ShareX
                        </Menu.Item>
                        <Menu.Item name='recent' active={activeItem === 'recent'} onClick={this.handleItemClick} as={Link} to="/recent">
                            Recent URLs
                        </Menu.Item>
                        <Menu.Item name='lookup' active={activeItem === 'lookup'} onClick={this.handleItemClick} as={Link} to="/lookup">
                            Lookup
                        </Menu.Item>
                        <Menu.Item name='about' active={activeItem === 'about'} onClick={this.handleItemClick} as={Link} to="/about">
                            About
                        </Menu.Item>
                        <Menu.Menu position='right'>
                            <Menu.Item onClick={this.handleLogout}>Logout</Menu.Item>
                        </Menu.Menu>
                    </Menu>
                    <Route exact path="/" component={Home} />
                    <Route path="/about" component={About} />
                    <Route path="/ShareX" component={ShareX} />
                    <Route path="/Lookup" component={Lookup} />
                </Container>
            </HashRouter>
        )
    }
}

ReactDOM.render((
    <BaseComponent />
), document.getElementById('root'))
