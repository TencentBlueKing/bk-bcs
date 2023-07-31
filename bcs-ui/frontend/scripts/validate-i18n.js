const yaml = require('js-yaml')
const Dot = require('dot-object');
const fs = require('fs')
const { resolve } = require('path');
const { extractI18NItemsFromVueFiles, readVueFiles } = require('vue-i18n-extract');
const lodash = require('lodash')

// const data = require('../src/i18n/zh-CN/unknow.json')
// fs.writeFileSync('./src/i18n/zh-CN.yaml', yaml.dump(data))
const dot = new Dot('.');

function hasChinese(str) {
  var reg = /[\u4e00-\u9fa5]/g;
  return reg.test(str);
}

const yamlData = fs.readFileSync('./src/i18n/zh-CN.yaml')
const jsonData = yaml.load(yamlData)
const flatData = dot.dot(jsonData)
const repeatData = {}
const chineseData = []
const revertData = Object.keys(flatData).reduce((pre, key) => {
  if (hasChinese(key)) {
    chineseData.push(key)
  }
  if (pre[flatData[key]]) {
    // console.warn('重复文案', flatData[key])
    repeatData[flatData[key]] = key
  } else {
    pre[flatData[key]] = key
  }
  return pre
}, {})

console.log(`中文key数量: ${Object.keys(chineseData).length}`, chineseData)
console.log(`重复文案数量${Object.keys(repeatData).length}`, repeatData)

// const vueFiles = readVueFiles(resolve(process.cwd(), './src/**/*.?(js|vue|ts|tsx|jsx|html)'));
// const I18NItems = extractI18NItemsFromVueFiles(vueFiles);
// const unTranslation = I18NItems.reduce((pre, item) => {
//   if(!revertData[item.path]) {
//     pre[item.path] = item.file
//   }
//   return pre
// }, {})
// console.log(`未翻译文案数量${Object.keys(unTranslation).length}`, unTranslation)

// // 文件替换
// I18NItems.forEach((item) => {
//   const repeatDataxxx = {
//     '命名空间': 'k8s.namespace',
//     '集群列表': 'generic.label.clusterList',
//     '网络': 'k8s.networking',
//     '集群管理': 'iam.actionMap.cluster_manage',
//     '存储': 'generic.label.storage',
//     '完成': 'generic.status.done',
//     '日志采集': 'logCollector.text',
//     '资源视图': 'cluster.button.dashboard',
//     '云凭证': 'cluster.create.label.cloudToken',
//     '节点列表': 'cluster.nodeList.text',
//     '节点模板': 'cluster.nodeTemplate.text',
//     '变量管理': 'deploy.variable.env',
//     '组件库': 'plugin.tools.title',
//     '事件查询': 'projects.eventQuery.title',
//     '操作记录': 'projects.operateAudit.record',
//     '项目信息': 'projects.project.info',
//     '配置': 'dashboard.network.config',
//     '标准输出': 'logCollector.label.collectorType.stdout'
//   }
//   if (repeatDataxxx[item.path]) {
//     let fileString = fs.readFileSync(item.file).toString()
//     fileString = fileString.replace(`\'${item.path}\'`, `'${repeatDataxxx[item.path]}'`)
//     fileString = fileString.replace(`\"${item.path}\"`, `"${repeatDataxxx[item.path]}"`)
//     fs.writeFileSync(item.file, fileString)
//     return
//   }
//   if (repeatData[item.path] || !revertData[item.path]) return
//   let fileString = fs.readFileSync(item.file).toString()
//   fileString = fileString.replace(`\'${item.path}\'`, `'${revertData[item.path]}'`)
//   fileString = fileString.replace(`\"${item.path}\"`, `"${revertData[item.path]}"`)
//   fs.writeFileSync(item.file, fileString)
// })

const enJsonData = fs.readFileSync('./src/i18n/en-US.yaml')
const enFlatData = dot.dot(yaml.load(enJsonData))
const missingkeys = Object.keys(enFlatData).filter(key => {
  return !flatData[key]
})
console.log(`中英文key不一致, 数量: ${missingkeys.length}`, missingkeys)


const vueFiles = readVueFiles(resolve(process.cwd(), './src/**/*.?(js|vue|ts|tsx|jsx|html)'));
const I18NItems = extractI18NItemsFromVueFiles(vueFiles);
const unTranslationData = I18NItems.filter(item => {
  return !enFlatData[item.path] || !flatData[item.path]
}).map(item => item.path)
console.log(`未翻译文案, 数量: ${unTranslationData.length}`, unTranslationData)

