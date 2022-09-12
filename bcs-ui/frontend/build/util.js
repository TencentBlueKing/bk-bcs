

const path = require('path')
const os = require('os')
const config = require('./config')

const env = process.env.NODE_ENV

exports.assetsPath = _path => {
    return path.posix.join(config[env === 'production' ? 'build' : 'dev'].assetsSubDirectory, _path)
}

exports.getIP = () => {
    const ifaces = os.networkInterfaces()
    const defultAddress = '127.0.0.1'
    let ip = defultAddress

    /* eslint-disable fecs-use-for-of, no-loop-func */
    for (const dev in ifaces) {
        if (ifaces.hasOwnProperty(dev)) {
            /* jshint loopfunc: true */
            ifaces[dev].forEach(details => {
                if (ip === defultAddress && details.family === 'IPv4') {
                    ip = details.address
                }
            })
        }
    }
    /* eslint-enable fecs-use-for-of, no-loop-func */
    return ip
}

exports.resolve = dir => {
    return path.join(__dirname, '..', dir)
}
