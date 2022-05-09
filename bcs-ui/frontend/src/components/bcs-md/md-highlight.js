import hljs from 'highlight.js/lib/highlight'
import yaml from 'highlight.js/lib/languages/yaml'
import 'highlight.js/styles/vs2015.css'
// const registerLanguages = [
//     // { name: 'css', path: require('highlight.js/lib/languages/css') },
//     // { name: 'javascript', path: require('highlight.js/lib/languages/javascript') },
//     // { name: 'bash', path: require('highlight.js/lib/languages/bash') },
//     // { name: 'python', path: require('highlight.js/lib/languages/python') },
//     // { name: 'yaml', path: require('highlight.js/lib/languages/yaml') }
// ]
// registerLanguages.forEach(lang => hljs.registerLanguage(lang.name, lang.path))

hljs.registerLanguage('yaml', yaml)
export default hljs
