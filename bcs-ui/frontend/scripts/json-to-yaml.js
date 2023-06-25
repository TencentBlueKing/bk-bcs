const yaml = require('js-yaml')
const fs = require('fs')

const data = require('../src/i18n/zh-CN/unknow.json')
fs.writeFileSync('./src/i18n/zh-CN.yaml', yaml.dump(data))