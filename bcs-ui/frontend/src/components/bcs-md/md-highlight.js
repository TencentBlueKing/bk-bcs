const hljs = require('highlight.js')
const registerLanguages = [
    { name: 'css', path: require('highlight.js/lib/languages/css') },
    { name: 'javascript', path: require('highlight.js/lib/languages/javascript') },
    { name: 'bash', path: require('highlight.js/lib/languages/bash') },
    { name: 'python', path: require('highlight.js/lib/languages/python') },
    { name: 'yaml', path: require('highlight.js/lib/languages/yaml') }
]
registerLanguages.forEach(lang => hljs.registerLanguage(lang.name, lang.path))
module.exports = hljs
