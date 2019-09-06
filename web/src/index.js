import React, { Component } from 'react'
import ReactDOM from 'react-dom';
import { HashRouter, Route, Link } from 'react-router-dom'
import { Menu, Container, Modal, Button, Image, Icon } from 'semantic-ui-react'
import toastr from 'toastr'
import 'semantic-ui-css/semantic.min.css';
import 'toastr/build/toastr.css';

import About from './About/About'
import Home from './Home/Home'
import ShareX from './ShareX/ShareX'
import Lookup from './Lookup/Lookup'
import Recent from './Recent/Recent'
import Visitors from './Visitors/Visitors'

import util from './util/util'
export default class BaseComponent extends Component {
    state = {
        authPopupOpened: true,
        userData: {},
        authorized: false,
        activeItem: "",
        info: {}
    }

    handleItemClick = (e, { name }) => this.setState({ activeItem: name })

    onOAuthClose = () => {
        this.setState({ authPopupOpened: true })
    }

    componentDidMount() {
        fetch('/api/v1/info', {credentials: 'include'})
            .then(d => d.json())
            .then(info => this.setState({ info }))
            .then(() => this.checkAuth())
            .catch(e => util._reportError(e, "info"))
    }

    checkAuth = () => {
        const token = window.localStorage.getItem('token');
        if (token) {
            fetch('/api/v1/auth/check', {
                method: 'POST',
                credentials: 'include',
                body: JSON.stringify({
                    Token: token
                }),
                headers: {
                    'Content-Type': 'application/json'
                }
            })
                .then(res => res.ok ? res.json() : Promise.reject(`incorrect response status code: ${res.status}; text: ${res.statusText}`))
                .then(d => this.setState({
                    userData: d,
                    authorized: true
                }))
                .catch(e => {
                    toastr.error(`Could not fetch check: ${e}`)
                    window.localStorage.removeItem('token');
                    this.setState({ authorized: false })
                })
        }
    }

    onOAuthCallback = data => {
        if (data.isTrusted) {
            // clear the old event listener, so that the event can only emitted be once
            window.removeEventListener('message', this.onOAuthCallback);
            window.localStorage.setItem('token', data.data);
            this.checkAuth();
            this._oAuthPopup = null;
        }
    }

    onOAuthClick = provider => {
        window.addEventListener('message', this.onOAuthCallback, false);
        var url = `${window.location.origin}/api/v1/auth/${provider}/login`;
        if (!this._oAuthPopup || this._oAuthPopup.closed) {
            // Open the oAuth window that is it centered in the middle of the screen
            var wwidth = 400,
                wHeight = 500;
            this._oAuthPopup = window.open(url, 'Authenticate with OAuth 2.0', `width=${wwidth}, height=${wHeight}, top=${(window.screen.height / 2) - (wHeight / 2)}, left=${(window.screen.width / 2) - (wwidth / 2)}`)
        } else {
            this._oAuthPopup.location = url;
        }
    }

    onProxyAuthOpen = () => {
      // the token contents don't matter for proxy auth, but
      // checkAuth() needs it to be set to something
      window.localStorage.setItem('token', {"lorem": "ipsum"});
      this.checkAuth();
      this.setState({ authPopupOpened: false })
    }

    handleLogout = () => {
        window.localStorage.removeItem("token")
        this.setState({ authorized: false })
    }

    render() {
        const { authPopupOpened, authorized, activeItem, userData, info } = this.state
        if (!authorized) {
          if (Array.isArray(info.providers) && info.providers.includes("proxy")) {
            // window.localStorage.setItem('token', {"lorem": "ipsum"});
            // this.checkAuth();
            return (
              <Modal size='tiny' open={authPopupOpened} onMount={this.onProxyAuthOpen}>
                <Modal.Header>Authentication</Modal.Header>
                <Modal.Content><p>If you are seeing this, you have not successfully authenticated to the proxy.</p></Modal.Content>
              </Modal>
            )
          } else if (Array.isArray(info.providers)) {
            return (
                <Modal size='tiny' open={authPopupOpened} onClose={this.onOAuthClose}>
                    <Modal.Header>
                        Authentication
                    </Modal.Header>
                    <Modal.Content>
                        <p>The following authentication services are currently available:</p>
                        {info && <div className='ui center aligned segment'>
                            {info.providers.length === 0 && <p>There are currently no correct oAuth credentials maintained.</p>}
                            {info.providers.includes("google") && <div>
                                <Button className='ui google plus button' onClick={this.onOAuthClick.bind(this, "google")}>
                                    <Icon name='google' /> Login with Google
                                </Button>
                                {info.providers.includes("github") && <div className="ui divider"></div>}
                            </div>}
                            {info.providers.includes("github") && <div>
                                <Button style={{ backgroundColor: "#333", color: "white" }} onClick={this.onOAuthClick.bind(this, "github")}>
                                    <Icon name='github' /> Login with GitHub
                                </Button>
                            {info.providers.includes("okta") && <div className="ui divider"></div>}
                            </div>}
                            {info.providers.includes("okta") && <div>
                                <Button style={{ color: "#007dc1" }} onClick={this.onOAuthClick.bind(this, "okta")}>
                                    <Image src='/images/okta_logo.png' style={{ width: "16px", height: "16px", marginBottom: ".15em" }}  avatar /> Login with Okta
                                </Button>
                            {info.providers.includes("generic_oidc") && <div className="ui divider"></div>}
                            </div>}
                            {info.providers.includes("generic_oidc") && <div>
                                <Button style={{ color: "#007dc1" }} onClick={this.onOAuthClick.bind(this, "generic_oidc")}>
                                    <Image src='/images/generic_oidc_logo.png' style={{ width: "16px", height: "16px", marginBottom: ".15em" }}  avatar /> Login with OpenID Connect
                                </Button>
                            </div>}
                            {info.providers.includes("microsoft") && <div>
                                <div className="ui divider"></div>
                                <Button style={{ backgroundColor: "#0067b8", color: "white" }} onClick={this.onOAuthClick.bind(this, "microsoft")}>
                                    <Icon name='windows' /> Login with Microsoft
                                </Button>
                            </div>}
                        </div>}
                    </Modal.Content>
                </Modal >
            )
          }
        }
        return (
            <HashRouter>
                <Container style={{ padding: "15px 0" }}>
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
                        <Menu.Item name='about' active={activeItem === 'about'} onClick={this.handleItemClick} as={Link} to={{
                            pathname: "/about",
                            state: { info }
                        }}>
                            About
                        </Menu.Item>
                          <Menu.Menu position='right'>
                            {userData.Name && <Menu.Item>{userData.Name}</Menu.Item>}
                            {Array.isArray(info.providers) && !info.providers.includes("proxy") &&
                              <Menu.Item onClick={this.handleLogout}>Logout</Menu.Item>}
                          </Menu.Menu>
                    </Menu>
                    <Route exact path="/" component={Home} />
                    <Route path="/about" render={() => <About info={info} />} />
                    <Route path="/ShareX" component={ShareX} />
                    <Route path="/Lookup" component={Lookup} />
                    <Route path="/recent" component={Recent} />
                    <Route path="/visitors/:id" component={Visitors} />
                </Container>
            </HashRouter>
        )
    }
}

ReactDOM.render((
    <BaseComponent />
), document.getElementById('root'))
