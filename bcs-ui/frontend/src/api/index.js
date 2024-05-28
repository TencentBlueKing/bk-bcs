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

import axios from 'axios';
import cookie from 'cookie';

// 登录弹窗
import { showLoginModal } from '@blueking/login-modal';

import { messageError } from '../common/bkmagic';
import { bus } from '../common/bus';

import CachedPromise from './cached-promise';
import RequestQueue from './request-queue';

import { random } from '@/common/util';

// axios 实例
const axiosInstance = axios.create({
  xsrfCookieName: 'bcs_csrftoken',
  xsrfHeaderName: 'X-CSRFToken',
  withCredentials: true,
});

/**
 * request interceptor
 */
axiosInstance.interceptors.request.use((config) => {
  const CSRFToken = cookie.parse(document.cookie).bcs_csrftoken;
  if (CSRFToken) {
    config.headers['X-CSRFToken'] = CSRFToken;
  }
  if (!window.BCS_CONFIG.disableTracing) {
    config.headers.Traceparent = `00-${random(32, 'abcdef0123456789')}-${random(16, 'abcdef0123456789')}-01`;
  }
  config.headers['X-Requested-With'] = 'XMLHttpRequest';
  return config;
}, error => Promise.reject(error));

const http = {
  queue: new RequestQueue(),
  cache: new CachedPromise(),
  cancelRequest: requestId => http.queue.cancel(requestId),
  cancelCache: requestId => http.cache.delete(requestId),
  cancel: (requestId) => {
    Promise.all([http.cancelRequest(requestId), http.cancelCache(requestId)]);
  },
};

const methodsWithoutData = ['get', 'head', 'options'];
const methodsWithData = ['delete', 'post', 'put', 'patch'];
const allMethods = [...methodsWithoutData, ...methodsWithData];

// 在自定义对象 http 上添加各请求方法
allMethods.forEach((method) => {
  Object.defineProperty(http, method, {
    get() {
      return getRequest(method);
    },
  });
});

/**
 * 获取 http 不同请求方式对应的函数
 *
 * @param {string} http method 与 axios 实例中的 method 保持一致
 *
 * @return {Function} 实际调用的请求函数
 */
function getRequest(method) {
  return (url, data, config) => getPromise(method, url, data, config);
}

/**
 * 实际发起 http 请求的函数，根据配置调用缓存的 promise 或者发起新的请求
 *
 * @param {string} method http method 与 axios 实例中的 method 保持一致
 * @param {string} url 请求地址
 * @param {Object} data 需要传递的数据, 仅 post/put/patch 三种请求方式可用
 * @param {Object} userConfig 用户配置，包含 axios 的配置与本系统自定义配置
 *
 * @return {Promise} 本次http请求的Promise
 */
async function getPromise(method, url, data, userConfig = {}) {
  const config = initConfig(method, url, userConfig);
  let promise;
  if (config.cancelPrevious) {
    await http.cancel(config.requestId);
  }

  if (config.clearCache) {
    http.cache.delete(config.requestId);
  } else {
    promise = http.cache.get(config.requestId);
  }

  if (config.fromCache && promise) {
    return promise;
  }

  // eslint-disable-next-line @typescript-eslint/no-misused-promises
  promise = new Promise(async (resolve, reject) => {
    const axiosRequest = methodsWithData.includes(method)
      ? axiosInstance[method](url, data, config)
      : axiosInstance[method](url, config);

    try {
      const response = await axiosRequest;
      Object.assign(config, response.config || {});
      handleResponse({ config, response, resolve, reject });
    } catch (httpError) {
      // http status 错误
      // 避免 cancel request 时出现 error message
      if (httpError?.message && httpError.message.type === 'cancel') {
        console.warn('请求被取消：', url);
        return;
      }

      Object.assign(config, httpError.config);
      reject(httpError);
    }
  }).catch(codeError =>
    // code 错误
    handleReject(codeError, config))
    .finally(() => {
      http.queue.delete(config.requestId);
    });

  // 添加请求队列
  http.queue.set(config);
  // 添加请求缓存
  http.cache.set(config.requestId, promise);

  return promise;
}

/**
 * 处理 http 请求成功结果
 *
 * @param {Object} 请求配置
 * @param {Object} cgi 原始返回数据
 * @param {Function} promise 完成函数
 * @param {Function} promise 拒绝函数
 */
function handleResponse({ config, response, resolve, reject }) {
  // 容器服务 -> 配置 -> heml 模板集 helm/getQuestionsMD 请求 response 是一个 string 类型 markdown 文档内容
  if (typeof response === 'string') {
    resolve(response, config);
    return;
  }

  if (response.data?.code !== 0 && config.globalError) {
    reject({ response });
    return;
  }

  if (config.originalResponse) {
    resolve(response);
    return;
  }

  resolve(response.data);
}

// 不弹tips的特殊状态码
const CUSTOM_HANDLE_CODE = [4005, 40300, 40403, 4005002, 4005003, 4005005];
/**
 * 处理 http 请求失败结果
 *
 * @param {Object} Error 对象
 * @param {config} 请求配置
 *
 * @return {Promise} promise 对象
 */
function handleReject(error, config) {
  if (axios.isCancel(error) || error?.message === 'Request aborted') {
    return Promise.reject(error);
  }

  if (error.response) {
    const { status, data } = error.response;
    let message = error?.message || data?.message;
    if (status === 401) {
      // 登录弹窗
      // eslint-disable-next-line camelcase
      let loginUrl;
      const successUrl = `${location.origin}/login_success.html`;
      if (process.env.NODE_ENV === 'development') {
        loginUrl = `${window.LOGIN_FULL}plain/?size=big&c_url=${encodeURIComponent(successUrl)}`;
      } else {
        loginUrl = `${data.data.login_url.simple}?c_url=${encodeURIComponent(successUrl)}`;
      }
      // 传入最终的登录地址，弹出登录窗口，更多选项参考 Options
      showLoginModal({ loginUrl });

      return;
    } if (status === 500) {
      message = window.i18n.t('generic.msg.error.system');
    } else if (status === 403) {
      message = window.i18n.t('generic.msg.warning.403');
    } else if ([4005, 40300].includes(data?.code)) {
      bus.$emit('show-apply-perm-modal', data?.data);
    } else if (data?.code === 40403) {
      bus.$emit('show-apply-perm-modal', data?.web_annotations);
    }

    // eslint-disable-next-line camelcase
    if (data?.code === 500 && data?.request_id) { // code为500的请求需要带上request_id
      // eslint-disable-next-line camelcase
      message = `${message}( ${data?.request_id.slice(0, 8)} )`;
    }

    if (config.globalError && !CUSTOM_HANDLE_CODE.includes(data?.code)) {
      messageError({
        code: data?.code || '',
        overview: message,
        suggestion: '',
        assistant: window.BCS_CONFIG?.contact,
        type: 'key-value',
        details: {
          traceparent: error.response?.config?.headers?.Traceparent || '--',
          requestId: error.response?.headers?.['x-request-id'] || data?.requestID || data?.request_id || '--',
          message: data.message || '--',
          URL: error.response?.config?.url || '--',
        },
      });
    }
  } else if (error.message === 'Network Error') {
    messageError(window.i18n.t('generic.msg.error.network'));
  } else if (error.message && config.globalError) {
    messageError(error.message);
  }
  return Promise.reject(error);
}

/**
 * 初始化本系统 http 请求的各项配置
 *
 * @param {string} http method 与 axios 实例中的 method 保持一致
 * @param {string} 请求地址, 结合 method 生成 requestId
 * @param {Object} 用户配置，包含 axios 的配置与本系统自定义配置
 *
 * @return {Promise} 本次 http 请求的 Promise
 */
function initConfig(method, url, userConfig) {
  const defaultConfig = {
    ...getCancelToken(),
    // http 请求默认 id
    requestId: `${method}_${url}`,
    // 是否全局捕获异常
    globalError: true,
    // 是否直接复用缓存的请求
    fromCache: false,
    // 是否在请求发起前清楚缓存
    clearCache: false,
    // 响应结果是否返回原始数据
    originalResponse: false,
    // 当路由变更时取消请求
    cancelWhenRouteChange: true,
    // 取消上次请求
    cancelPrevious: false,
  };
  return Object.assign(defaultConfig, userConfig);
}

/**
 * 生成 http 请求的 cancelToken，用于取消尚未完成的请求
 *
 * @return {Object} {cancelToken: axios 实例使用的 cancelToken, cancelExcutor: 取消http请求的可执行函数}
 */
function getCancelToken() {
  let cancelExcutor;
  const cancelToken = new axios.CancelToken((excutor) => {
    cancelExcutor = excutor;
  });
  return {
    cancelToken,
    cancelExcutor,
  };
}

export default http;
