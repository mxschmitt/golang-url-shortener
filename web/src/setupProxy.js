const proxy = require('http-proxy-middleware')
const pkg = require("../package.json")

module.exports = app => {
  app.use(proxy('/api', {
    target: pkg.proxy
  }))

  app.use(proxy('/d', {
    target: pkg.proxy
  }))
}