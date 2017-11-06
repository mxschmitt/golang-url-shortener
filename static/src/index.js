import registerServiceWorker from './registerServiceWorker';
import React from 'react'
import ReactDOM from 'react-dom';
import {
    HashRouter,
    Route
} from 'react-router-dom'

import Header from './Header/Header'
import About from './About/About'
import Home from './Home/Home'

import 'semantic-ui-css/semantic.min.css';

ReactDOM.render((
    <HashRouter>
        <div>
            <Header />
            <Route exact path="/" component={Home} />
            <Route path="/about" component={About} />
        </div>
    </HashRouter>
), document.getElementById('root'))

registerServiceWorker();