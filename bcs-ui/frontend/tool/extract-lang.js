const fs = require('fs')
const { resolve } = require('path')

const lang = require('../src/common/lang')

const zh = {}
const en = {}
Object.keys(lang).forEach(key => {
    zh[key] = key
    en[key] = lang[key][0]
})

fs.writeFileSync(resolve(__dirname, '../src/i18n/zh.js'), JSON.stringify(zh, null, 4))
fs.writeFileSync(resolve(__dirname, '../src/i18n/en.js'), JSON.stringify(en, null, 4))
