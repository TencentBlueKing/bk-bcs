

const merge = require('webpack-merge')
const prodEnv = require('./prod.env')

const NODE_ENV = JSON.stringify('development')

module.exports = merge(prodEnv, {
    'process.env': {
        'NODE_ENV': NODE_ENV
    },
    staticUrl: '/static',
    NODE_ENV: NODE_ENV,
    VERSION: JSON.stringify('ieod'),
    LOGIN_SERVICE_URL: JSON.stringify(''),
    SENTRY_URL: JSON.stringify('')
})
