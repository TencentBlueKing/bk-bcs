/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/
import hljs from 'highlight.js/lib/highlight';
import yaml from 'highlight.js/lib/languages/yaml';
import 'highlight.js/styles/vs2015.css';
// const registerLanguages = [
//     // { name: 'css', path: require('highlight.js/lib/languages/css') },
//     // { name: 'javascript', path: require('highlight.js/lib/languages/javascript') },
//     // { name: 'bash', path: require('highlight.js/lib/languages/bash') },
//     // { name: 'python', path: require('highlight.js/lib/languages/python') },
//     // { name: 'yaml', path: require('highlight.js/lib/languages/yaml') }
// ]
// registerLanguages.forEach(lang => hljs.registerLanguage(lang.name, lang.path))

hljs.registerLanguage('yaml', yaml);
export default hljs;
