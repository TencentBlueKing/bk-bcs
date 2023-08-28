const { resolve } = require('path');
const fs = require('fs');
const yamljs = require('js-yaml')
const { extractI18NReport, readVueFiles, extractI18NItemsFromVueFiles, readLanguageFiles } = require('vue-i18n-extract');
const Dot = require('dot-object');

const dot = new Dot('.');
const config = {
  filePath: './src/**/*.?(js|vue|ts|tsx|jsx|html)',
  i18nPath: './src/i18n/**/*.?(yaml)'
}

const vueFiles = readVueFiles(resolve(process.cwd(), config.filePath));
const I18NItems = extractI18NItemsFromVueFiles(vueFiles);

const languageFiles = readLanguageFiles(resolve(process.cwd(), config.i18nPath));

Object.keys(languageFiles).forEach(index => {
  const item = languageFiles[index]
  const lang = item.fileName.replace(/^.*[\\\/]/, '')
  const content = JSON.parse(item.content)
  const flatContent = dot.dot(content)

  const data = I18NItems.reduce((pre, data) => {
    if (data.file.indexOf('src/views') > -1) {
      const filePath = data.file.split('/')
      const path = filePath.slice(1, filePath.length - 1).filter(path => !['src', 'views'].includes(path)).slice(0, 2).join('.').replace(/-/g, '')
      try {
        dot.str(`${path}.${data.path}`, flatContent[data.path], pre)  
      } catch(_) {
        debugger
      }
    } else {
      dot.str(`component.${data.path}`, flatContent[data.path], pre)
    }
    return pre
  }, {})

 fs.writeFileSync(`./lang-${index}.yaml`, yamljs.dump(data))
})
