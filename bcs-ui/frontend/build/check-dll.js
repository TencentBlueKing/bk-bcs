

const path = require('path')
const fse = require('fs-extra')
const npm = require('npm')

const manifestExist = fse.pathExistsSync(path.resolve(__dirname, '..', 'static', 'lib-manifest.json'))
const bundleExist = fse.pathExistsSync(path.resolve(__dirname, '..', 'static', 'lib.bundle.js'))

if (!(manifestExist & bundleExist)) {
    npm.load({}, () => {
        npm.run('dll', err => {
            if (err) {
                throw err
            }
        })
    })
}
