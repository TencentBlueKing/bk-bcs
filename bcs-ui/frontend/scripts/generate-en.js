const yamljs = require('js-yaml')
const fs = require('fs')
const Dot = require('dot-object');
const dot = new Dot('.');

const enMap = yamljs.load(fs.readFileSync('./src/i18n/en-US.yaml'))
const chData = yamljs.load(fs.readFileSync('./src/i18n/zh-CN.yaml'))

const data = dot.dot(chData)
Object.keys(data).forEach(path => {
  const k = data[path]
  dot.str(path, enMap[k], chData)
})
// console.log(chData)
fs.writeFileSync('./src/i18n/en-US.yaml', yamljs.dump(chData))