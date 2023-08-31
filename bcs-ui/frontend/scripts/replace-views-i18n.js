const { resolve } = require('path');
const fs = require('fs');
const yamljs = require('js-yaml')
const { extractI18NReport, readVueFiles, extractI18NItemsFromVueFiles, readLanguageFiles } = require('vue-i18n-extract');
const Dot = require('dot-object');

const dot = new Dot('.');
const config = {
  filePath: './src/**/*.?(js|vue|ts|tsx|jsx|html)',
  i18nPath: './src/i18n/zh-CN.yaml'
}

const vueFiles = readVueFiles(resolve(process.cwd(), config.filePath));
const I18NItems = extractI18NItemsFromVueFiles(vueFiles);

const languageFiles = readLanguageFiles(resolve(process.cwd(), config.i18nPath));

const content = JSON.parse(languageFiles[0].content)
const flatContent = dot.dot(content);
const replaceData = Object.keys(flatContent).filter(key => flatContent[key] !== key).map(key => ({
  path: flatContent[key],
  key
})).reduce((pre, item) => {
  pre[item.path] = item.key
  return pre
}, {})

I18NItems.filter(item => replaceData[item.path]).forEach((item) => {
  const fileData = fs.readFileSync(item.file)
  let str = fileData.toString()
  str = str.replace(item.path, replaceData[item.path])
  fs.writeFileSync(item.file, str)
})