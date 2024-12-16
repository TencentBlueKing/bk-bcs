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
import http, { streamTypes } from '@/api';
import { json2Query } from '@/common/util';

const methodsWithoutData = ['get', 'head', 'options', 'delete'];
const defaultConfig = { needRes: false };

// 转换URL为绝对路径
export const resolveUrlPrefix = (url, { domain = window.DEVOPS_BCS_API_URL, prefix = '' } = {}) => {
  if (/(http|https):\/\/([\w.]+\/?)\S*/.test(url) || url.indexOf(domain) === 0) {
    return url;
  }
  return `${process.env.NODE_ENV === 'development' ? '' : domain}${prefix}${url}`;
};

export const parseUrl = (reqMethod, url, body = {}) => {
  let params = JSON.parse(JSON.stringify(body));
  // 全局URL变量替换
  const variableData = {
    $projectId: window._project_id_,
    $projectCode: window._project_code_,
  };
  Object.keys(params).forEach((key) => {
    // 自定义url变量
    if (key.indexOf('$') === 0) {
      variableData[key] = params[key];
    }
  });
  let newUrl = url;
  Object.keys(variableData).forEach((key) => {
    if (!variableData[key]) {
      // console.warn(`路由变量未配置${key}`, url);
      // 去除后面的路径符号
      newUrl = newUrl.replace(`/${key}`, '');
    }
    // todo 重名问题
    newUrl = newUrl.replace(new RegExp(`\\${key}\\b`, 'g'), variableData[key]);
    // 删除URL上的参数
    delete params[key];
  });
  // 参数拼接在URL后面
  if (methodsWithoutData.includes(reqMethod)) {
    const query = json2Query(params, '');
    if (query) {
      newUrl += `?${query}`;
    }
    params = null;
  }
  return {
    url: newUrl,
    params,
  };
};

export const request = (method, url) => (params = {}, config = {}) => {
  const reqMethod = method.toLowerCase();
  const reqConfig = Object.assign({}, defaultConfig, config);

  let newUrl;
  let newParams;
  if (streamTypes.includes(config.responseType) || params instanceof FormData) {
    const result = parseUrl(reqMethod, resolveUrlPrefix(url), {});
    newUrl = result.url;
    newParams = params;
  } else {
    const result = parseUrl(reqMethod, resolveUrlPrefix(url), params);
    newUrl = result.url;
    newParams = result.params;
  }
  const req = http[reqMethod](newUrl, newParams, reqConfig);
  return req.then((res) => {
    if (reqConfig.needRes) return Promise.resolve(res);

    return Promise.resolve(res.data);
  }).catch((err) => {
    console.error('request error', err);
    return Promise.reject(err);
  });
};

export const createRequest = ({ domain, prefix = '' } = {}) => (method, url) => request(method, resolveUrlPrefix(url, { domain, prefix }));

export default request;
