// 版本一：根据i18n的分类规则，eg：$t('button.测试') 会自动更新到button.json目录下面，但是改造比较麻烦，每个翻译都加前缀
// const { resolve } = require('path');
// const fs = require('fs');
// const Dot = require('dot-object');
// const { extractI18NReport, readVueFiles, extractI18NItemsFromVueFiles, readLanguageFiles } = require('vue-i18n-extract');

// const vueFiles = readVueFiles(resolve(process.cwd(), './src/**/*.?(js|vue|ts|tsx|jsx)'));
// const I18NItems = extractI18NItemsFromVueFiles(vueFiles);

// const dot = new Dot('.');
// const classifys = [];
// const languageFiles = readLanguageFiles(resolve(process.cwd(), './src/i18n/**/*.?(json)')).map(item => {
//   const paths = item.fileName.split('/');
//   const lang = paths[paths.length - 2];// 语言名
//   const classify = paths[paths.length - 1].split('.')[0]; // 分类名称
//   if (!classifys.includes(classify)) {
//     classifys.push(classify)
//   }
//   return {
//     ...item,
//     lang,
//     classify
//   }
// });
// const I18NLanguage = languageFiles.reduce((pre, item) => {
//   if (!pre[item.lang]) {
//     pre[item.lang] = []
//   }
//   const flatContent = dot.dot(JSON.parse(item.content));
//   Object.keys(flatContent).forEach(key => {
//     pre[item.lang].push({
//       path: item.classify && item.classify !== 'common' ? `${item.classify}.${key}` : key,
//       file: item.fileName
//     })
//   })
//   return pre;
// }, {});

// // 原始数据
// const report = extractI18NReport(I18NItems, I18NLanguage);
// // 分类数据
// const newReport = {
//   missingKeys: {},
//   unusedKeys: {},
//   maybeDynamicKeys: {}
// }
// Object.keys(report).forEach(key => {
//   if (!newReport[key]) {
//     newReport[key] = {};
//   }
//   const data = report[key];
//   const newData = data.reduce((pre, item) => {
//     const { path } = item;
//     const classify = path.split('.')[0];
//     if (classifys.includes(classify)) {
//       if (!pre[classify]) {
//         pre[classify] = []
//       };
//       pre[classify].push({
//         ...item,
//         // 去除分类前缀
//         path: path.slice(classify.length + 1)
//       });
//     } else {
//       if (!pre['common']) {
//         pre['common'] = []
//       };
//       pre['common'].push({
//         ...item
//       });
//     }
//     return pre;
//   }, {});
//   newReport[key] = newData;
// })

// // 添加未翻译文案
// languageFiles.forEach(languageFile => {
//   const content = JSON.parse(languageFile.content);
//   const classify = languageFile.classify;
//   if (newReport.missingKeys && newReport.missingKeys[classify]) {
//     newReport.missingKeys[classify].forEach(item => {
//       if (item.language && languageFile.lang === item.language) {
//         const defaultValue = languageFile.lang === 'zh-CN' ? item.path : ''
//         if (item.path.indexOf('.') === -1) {
//           dot.str(item.path, defaultValue, content);
//         } else {
//           content[item.path] = defaultValue
//         }
//       }
//     })
//   }
//   languageFile.content = JSON.stringify(content)
//   fs.writeFileSync(languageFile.path, JSON.stringify(content, null, 2))
// });
// // 删除未使用文案
// languageFiles.forEach(languageFile => {
//   const content = JSON.parse(languageFile.content);
//   const classify = languageFile.classify;
//   if (newReport.unusedKeys && newReport.unusedKeys[classify]) {
//     newReport.unusedKeys[classify].forEach(item => {
//       if (item.language && languageFile.lang === item.language) {
//         if (item.path.indexOf('.') === -1) {
//           dot.delete(item.path, content); 
//         } else {
//           delete content[item.path]
//         } 
//       }
//     })
//   }
//   fs.writeFileSync(languageFile.path, JSON.stringify(content, null, 2))
// });


// 版本二：分类仅作为参考用，无需要加上前缀，迭代过程把unknow.json文案放入对应分类即可
const { resolve } = require('path');
const fs = require('fs');
const Dot = require('dot-object');
const { extractI18NReport, readVueFiles, extractI18NItemsFromVueFiles, readLanguageFiles } = require('vue-i18n-extract');

const vueFiles = readVueFiles(resolve(process.cwd(), './src/**/*.?(js|vue|ts|tsx|jsx|html)'));
const I18NItems = extractI18NItemsFromVueFiles(vueFiles);

const dot = new Dot('.');
const classifys = [];
const languageFiles = readLanguageFiles(resolve(process.cwd(), './src/i18n/**/*.?(json)')).map(item => {
  const paths = item.fileName.split('/');
  const lang = paths[paths.length - 2];// 语言名
  const classify = paths[paths.length - 1].split('.')[0]; // 分类名称
  if (!classifys.includes(classify)) {
    classifys.push(classify)
  }
  return {
    ...item,
    lang,
    classify
  }
});
const checkMap = {}
const cacheMap = {}
const I18NLanguage = languageFiles.reduce((pre, item) => {
  if (!pre[item.lang]) {
    pre[item.lang] = []
  }
  const flatContent = dot.dot(JSON.parse(item.content));
  Object.keys(flatContent).forEach(key => {
    // 检测key是否重复
    if (checkMap?.[item.lang]?.[key]) {
      console.error(`\"${key}\"重复, 文件: ${item.fileName}`)
    } else {
      if (!checkMap[item.lang]) {
        checkMap[item.lang] = {}
      }
      checkMap[item.lang][key] = true
    }
    // 缓存分类数据
    if (!cacheMap[item.lang]) {
      cacheMap[item.lang] = {}
    }
    if (!cacheMap[item.lang][item.classify]) {
      cacheMap[item.lang][item.classify] = {}
    }
    cacheMap[item.lang][item.classify][key] = true

    pre[item.lang].push({
      path: key,
      file: item.fileName
    })
  })
  return pre;
}, {});
const findPathClassify = (path, lang) => {
  if (!cacheMap[lang]) return 'unknow'
  return Object.keys(cacheMap[lang]).find(key => cacheMap[lang][key].hasOwnProperty(path)) || 'unknow'
}
// 原始数据
const report = extractI18NReport(I18NItems, I18NLanguage);
// 分类数据
const newReport = {
  missingKeys: {},
  unusedKeys: {},
  maybeDynamicKeys: {}
}
Object.keys(report).forEach(key => {
  if (!newReport[key]) {
    newReport[key] = {};
  }
  const data = report[key];
  const newData = data.reduce((pre, item) => {
    const { path } = item;
    const classify = findPathClassify(path, item.language);
    if (!pre[classify]) {
      pre[classify] = []
    };
    pre[classify].push(item);
    return pre;
  }, {});
  newReport[key] = newData;
})

// 添加未翻译文案
languageFiles.forEach(languageFile => {
  const content = JSON.parse(languageFile.content);
  const classify = languageFile.classify;
  if (newReport.missingKeys && newReport.missingKeys[classify]) {
    newReport.missingKeys[classify].forEach(item => {
      if (item.language && languageFile.lang === item.language) {
        const defaultValue = languageFile.lang === 'zh-CN' ? item.path : ''
        if (item.path.indexOf('.') === -1) {
          dot.str(item.path, defaultValue, content);
        } else {
          content[item.path] = defaultValue
        }
      }
    })
  }
  languageFile.content = JSON.stringify(content)
  fs.writeFileSync(languageFile.path, JSON.stringify(content, null, 2))
});
// 删除未使用文案
languageFiles.forEach(languageFile => {
  const content = JSON.parse(languageFile.content);
  const classify = languageFile.classify;
  if (newReport.unusedKeys && newReport.unusedKeys[classify]) {
    newReport.unusedKeys[classify].forEach(item => {
      if (item.language && languageFile.lang === item.language) {
        if (item.path.indexOf('.') === -1) {
          dot.delete(item.path, content); 
        } else {
          delete content[item.path]
        } 
      }
    })
  }
  fs.writeFileSync(languageFile.path, JSON.stringify(content, null, 2))
});