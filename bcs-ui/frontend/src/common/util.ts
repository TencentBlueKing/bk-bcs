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
import moment from 'moment';

/**
 * 判断是否是对象
 *
 * @param {Object} obj 待判断的
 *
 * @return {boolean} 判断结果
 */
export function isObject(obj) {
  return obj !== null && typeof obj === 'object';
}

/**
 * 规范化参数
 *
 * @param {Object|string} type vuex type
 * @param {Object} payload vuex payload
 * @param {Object} options vuex options
 *
 * @return {Object} 规范化后的参数
 */
export function unifyObjectStyle(type, payload, options) {
  if (isObject(type) && type.type) {
    options = payload;
    payload = type;
    type = type.type;
  }

  if (process.env.NODE_ENV !== 'production') {
    if (typeof type !== 'string') {
      console.warn(`expects string as the type, but found ${typeof type}.`);
    }
  }

  return { type, payload, options };
}

/**
 * 异常处理
 *
 * @param {Object} err 错误对象
 * @param {Object} ctx 上下文对象，这里主要指当前的 Vue 组件
 */
export function catchErrorHandler(err, ctx) {
  const { data } = err;
  if (data) {
    if (!err.code || err.code === 404) {
      ctx.exceptionCode = {
        code: '404',
        msg: window.i18n.t('generic.msg.warning.404'),
      };
    } else if (err.code === 403) {
      ctx.exceptionCode = {
        code: '403',
        msg: window.i18n.t('generic.msg.warning.403'),
      };
    } else {
      console.error(err);
    }
  } else {
    // 其它像语法之类的错误不展示
    console.error(err);
  }
}

/**
 * 获取字符串长度，中文算两个，英文算一个
 *
 * @param {string} str 字符串
 *
 * @return {number} 结果
 */
export function getStringLen(str) {
  let len = 0;
  for (let i = 0; i < str.length; i++) {
    if (str.charCodeAt(i) > 127 || str.charCodeAt(i) === 94) {
      len += 2;
    } else {
      len += 1;
    }
  }
  return len;
}

/**
 * 转义特殊字符
 *
 * @param {string} str 待转义字符串
 *
 * @return {string} 结果
 */
export const escape = str => String(str).replace(/([.*+?^=!:${}()|[\]/\\])/g, '\\$1');

/**
 * 对象转为 url query 字符串
 *
 * @param {*} param 要转的参数
 * @param {string} key key
 *
 * @return {string} url query 字符串
 */
export function json2Query(param, key) {
  const mappingOperator = '=';
  const separator = '&';
  let paramStr = '';

  if (param instanceof String || typeof param === 'string'
            || param instanceof Number || typeof param === 'number'
            || param instanceof Boolean || typeof param === 'boolean'
  ) {
    paramStr += separator + key + mappingOperator + encodeURIComponent(param as any);
  } else if (typeof param === 'object') {
    Object.keys(param).forEach((p) => {
      const value = param[p];
      const k = (key === null || key === '' || key === undefined)
        ? p
        : key + (param instanceof Array ? `[${p}]` : `.${p}`);
      paramStr += separator + json2Query(value, k);
    });
  }
  return paramStr.substr(1);
}

/**
 *  获取元素相对于页面的高度
 *
 *  @param {Object} node 指定的 DOM 元素
 */
export function getActualTop(node) {
  let actualTop = node.offsetTop;
  let current = node.offsetParent;

  while (current !== null) {
    actualTop += current.offsetTop;
    current = current.offsetParent;
  }

  return actualTop;
}

/**
 *  获取元素相对于页面左侧的宽度
 *
 *  @param {Object} node 指定的 DOM 元素
 */
export function getActualLeft(node) {
  let actualLeft = node.offsetLeft;
  let current = node.offsetParent;

  while (current !== null) {
    actualLeft += current.offsetLeft;
    current = current.offsetParent;
  }

  return actualLeft;
}

/**
 * document 总高度
 *
 * @return {number} 总高度
 */
export function getScrollHeight() {
  let scrollHeight = 0;
  let bodyScrollHeight = 0;
  let documentScrollHeight = 0;

  if (document.body) {
    bodyScrollHeight = document.body.scrollHeight;
  }

  if (document.documentElement) {
    documentScrollHeight = document.documentElement.scrollHeight;
  }

  scrollHeight = (bodyScrollHeight - documentScrollHeight > 0) ? bodyScrollHeight : documentScrollHeight;

  return scrollHeight;
}

/**
 * 滚动条在 y 轴上的滚动距离
 *
 * @return {number} y 轴上的滚动距离
 */
export function getScrollTop() {
  let scrollTop = 0;
  let bodyScrollTop = 0;
  let documentScrollTop = 0;

  if (document.body) {
    bodyScrollTop = document.body.scrollTop;
  }

  if (document.documentElement) {
    documentScrollTop = document.documentElement.scrollTop;
  }

  scrollTop = (bodyScrollTop - documentScrollTop > 0) ? bodyScrollTop : documentScrollTop;

  return scrollTop;
}

/**
 * 浏览器视口的高度
 *
 * @return {number} 浏览器视口的高度
 */
export function getWindowHeight() {
  const windowHeight = document.compatMode === 'CSS1Compat'
    ? document.documentElement.clientHeight
    : document.body.clientHeight;

  return windowHeight;
}

/**
 * 在当前节点后面插入节点
 *
 * @param {Object} newElement 待插入 dom 节点
 * @param {Object} targetElement 当前节点
 */
export function insertAfter(newElement, targetElement) {
  const parent = targetElement.parentNode;
  if (parent.lastChild === targetElement) {
    parent.appendChild(newElement);
  } else {
    parent.insertBefore(newElement, targetElement.nextSibling);
  }
}

/**
 * 生成UUID
 *
 * @param  {number} len 位数
 * @param  {number} radix 进制
 * @return {string} uuid
 */
export function uuid(len, radix) {
  const chars = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz'.split('');
  const uuid: string[] = [];
  let i;
  radix = radix || chars.length;
  if (len) {
    for (i = 0; i < len; i++) {
      uuid[i] = chars[0 | Math.random() * radix];
    }
  } else {
    let r;
    // eslint-disable-next-line no-multi-assign
    uuid[8] = uuid[13] = uuid[18] = uuid[23] = '-';
    uuid[14] = '4';

    for (i = 0; i < 36; i++) {
      if (!uuid[i]) {
        r = 0 | Math.random() * 16;
        uuid[i] = chars[(i === 19) ? (r & 0x3) | 0x8 : r];
      }
    }
  }

  return uuid.join('');
}

/* 格式化日期
 *
 * @param  {string} date 日期
 * @param  {string} formatStr 格式
 * @return {str} 格式化后的日期
 */
export function formatDate(date, formatStr = 'YYYY-MM-DD hh:mm:ss') {
  if (!date) return '';

  const dateObj = new Date(date);
  const o = {
    'M+': dateObj.getMonth() + 1, // 月份
    'D+': dateObj.getDate(), // 日
    'h+': dateObj.getHours(), // 小时
    'm+': dateObj.getMinutes(), // 分
    's+': dateObj.getSeconds(), // 秒
    'q+': Math.floor((dateObj.getMonth() + 3) / 3), // 季度
    S: dateObj.getMilliseconds(), // 毫秒
  };
  if (/(Y+)/.test(formatStr)) {
    formatStr = formatStr.replace(RegExp.$1, (`${dateObj.getFullYear()}`).substr(4 - RegExp.$1.length));
  }
  for (const k in o) {
    if (new RegExp(`(${k})`).test(formatStr)) {
      formatStr = formatStr.replace(RegExp.$1, (RegExp.$1.length === 1) ? (o[k]) : ((`00${o[k]}`).substr((`${o[k]}`).length)));
    }
  }

  return formatStr;
}

/**
 * bytes 转换
 *
 * @param {Number} bytes 字节数
 * @param {Number} decimals 保留小数位
 *
 * @return {string} 转换后的值
 */
export function formatBytes(bytes, decimals = 2) {
  if (isNaN(bytes)) return bytes;
  if (parseFloat(bytes) === 0) {
    return '0 B';
  }
  const k = 1024;
  const dm = decimals;
  const sizes = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  if (i === -1) {
    return `${bytes} B`;
  }

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i] || ''}`;
}

/**
 * 判断是否为空
 * @param {Object} obj
 */
export function isEmpty(obj) {
  return typeof obj === 'undefined' || obj === null || obj === '';
}

/**
 * 清空对象的属性值
 * @param {*} obj
 */
export function clearObjValue(obj) {
  if (Object.prototype.toString.call(obj) !== '[object Object]') return;

  Object.keys(obj).forEach((key) => {
    if (Array.isArray(obj[key])) {
      obj[key] = [];
    } else if (Object.prototype.toString.call(obj[key]) === '[object Object]') {
      clearObjValue(obj[key]);
    } else {
      obj[key] = '';
    }
  });
}

/**
 * 获取对象key键下的值
 * @param {*} obj { a: { b: { c: '123' } } }
 * @param {*} key 'a.b.c'
 */
export const getObjectProps = (obj, key) => {
  if (!isObject(obj)) return obj[key];

  return String(key).split('.')
    .reduce((pre, k) => (pre?.[k] ? pre[k] : undefined), obj);
};

/**
 * 排序数组对象
 * 排序规则：1. 数字 => 2. 字母 => 3. 中文
 * @param {*} arr
 * @param {*} key
 */
export const sort = (arr, key, order = 'ascending', extraDataFn?): any[] => {
  if (!Array.isArray(arr)) return arr;
  const reg = /^[0-9a-zA-Z]/;
  const data = arr.sort((pre, next) => {
    if (isObject(pre) && isObject(next) && key) {
      // 合并extdata数据
      const preData = {
        ...pre,
        ...(extraDataFn?.(pre) || {}),
      };
      const nextData = {
        ...next,
        ...(extraDataFn?.(next) || {}),
      };
      const preStr = String(getObjectProps(preData, key));
      const nextStr = String(getObjectProps(nextData, key));
      if (reg.test(preStr) && !reg.test(nextStr)) {
        return -1;
      } if (!reg.test(preStr) && reg.test(nextStr)) {
        return 1;
      }
      return preStr.localeCompare(nextStr);
    }
    return (`${pre}`).toString().localeCompare((`${next}`));
  });
  return order === 'ascending' ? data : data.reverse();
};

// 格式化时间
export const formatTime = (timestamp, fmt = 'yyyy-MM-dd hh:mm:ss') => {
  const time = new Date(timestamp);
  const opt = {
    'M+': time.getMonth() + 1, // 月份
    'd+': time.getDate(), // 日
    'h+': time.getHours(), // 小时
    'm+': time.getMinutes(), // 分
    's+': time.getSeconds(), // 秒
    'q+': Math.floor((time.getMonth() + 3) / 3), // 季度
    S: time.getMilliseconds(), // 毫秒
  };
  if (/(y+)/.test(fmt)) {
    fmt = fmt.replace(RegExp.$1, (`${time.getFullYear()}`).substr(4 - RegExp.$1.length));
  }
  for (const k in opt) {
    if (new RegExp(`(${k})`).test(fmt)) {
      fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1)
        ? opt[k]
        : (`00${opt[k]}`).substr(String(opt[k]).length));
    }
  }
  return fmt;
};

export const copyText = (text) => {
  const textarea = document.createElement('textarea');
  document.body.appendChild(textarea);
  textarea.value = text;
  textarea.select();
  if (document.execCommand('copy')) {
    document.execCommand('copy');
  }
  document.body.removeChild(textarea);
};

export const deepEqual = (x, y) => {
  // 指向同一内存时
  if (x === y) {
    return true;
  } if ((typeof x === 'object' && x != null) && (typeof y === 'object' && y != null)) {
    if (Object.keys(x).length !== Object.keys(y).length) {
      return false;
    }
    // eslint-disable-next-line no-restricted-syntax
    for (const prop in x) {
      // eslint-disable-next-line no-prototype-builtins
      if (y.hasOwnProperty(prop)) {
        if (!deepEqual(x[prop], y[prop])) return false;
      } else {
        return false;
      }
    }
    return true;
  }
  return false;
};

export const timeZoneTransForm = (time, showTimeZone = true) => moment(time)
  .utcOffset(8 * 60 * 2)
  .format(showTimeZone ? 'YYYY-MM-DD HH:mm:ss ZZ' : 'YYYY-MM-DD HH:mm:ss');

export const timeFormat = (time, format = 'YYYY-MM-DD HH:mm:ss') => moment(time).format(format);


export const chainable = (obj, path, defaultValue = undefined) => {
  const travel = regexp => String.prototype.split
    .call(path, regexp)
    .filter(Boolean)
    .reduce((res, key) => (res !== null && res !== undefined ? res[key] : res), obj);
  const result = travel(/[,[\]]+?/) || travel(/[,[\].]+?/);
  return result === undefined || result === obj ? defaultValue : result;
};

export const timeDelta = (start, end) => {
  if (!start || !end) return;

  const time = (new Date(end).getTime() - new Date(start).getTime()) / 1000;
  if (time <= 0) return;

  const m = Math.floor(time / 60);
  const s = time - m * 60;
  return `${m ? `${m}m ` : ''}${s ? `${Math.ceil(s)}s ` : ''}`;
};

/**
 * 简单模板引擎
 * @param template
 * eg: user: '${username}' -> user: 'admin'
 */
export const renderTemplate = (template: string, params: Record<string, string> = {}) => {
  let str = template;
  Object.keys(params).forEach((key) => {
    str = str.replace(
      new RegExp(`\\$\\{${key}\\}`, 'g'),
      params[key],
    );
  });
  return str;
};

export const isRealObject = item => (item && typeof item === 'object' && !Array.isArray(item));

export const mergeDeep = (target, ...sources) => {
  if (!sources.length) return target;
  const source = sources.shift();

  if (isRealObject(target) && isRealObject(source)) {
    for (const key in source) {
      if (isRealObject(source[key])) {
        if (!target[key]) {
          Object.assign(target, {
            [key]: {},
          });
        }
        mergeDeep(target[key], source[key]);
      } else {
        Object.assign(target, {
          [key]: source[key],
        });
      }
    }
  }

  return mergeDeep(target, ...sources);
};

interface INode {
  id: string | number
  name: string
  isFolder?: boolean
  children: INode[]
}
export const path2Tree = (paths: string[]) => {
  if (!paths) return;
  const nodes: INode[] = [];
  paths.forEach((path) => {
    const tmpPaths = path.split('/');
    tmpPaths.reduce((pre, currentPath, index) => {
      const curID = tmpPaths.slice(0, index + 1).join('/');
      let node = pre.find(item => item.id === curID);
      if (!node) {
        node = {
          id: curID,
          name: currentPath,
          children: [],
          isFolder: index < (tmpPaths.length - 1),
        };
        pre.push(node);
      }
      return node.children;
    }, nodes);
  });
  return nodes;
};


export const validateIPv6 = function (a) {
  return new RegExp('(?!^(?:(?:.*(?:::.*::|:::).*)|::|[0:]+[01]|.*[^:]:|[0-9a-fA-F](?:.*:.*){8}[0-9a-fA-F]|(?:[0-9a-fA-F]:){1,6}[0-9a-fA-F])$)^(?:(::|[0-9a-fA-F]{1,4}:{1,2})([0-9a-fA-F]{1,4}:{1,2}){0,6}([0-9a-fA-F]{1,4}|::)?)$').test(a);
};
// 补全ipv6
export function padIPv6(simpeIpv6: string) {
  if (!validateIPv6(simpeIpv6)) return simpeIpv6;
  simpeIpv6 = simpeIpv6.toUpperCase();
  if (simpeIpv6 == '::') {
    return '0000:0000:0000:0000:0000:0000:0000:0000';
  }
  const arr = ['0000', '0000', '0000', '0000', '0000', '0000', '0000', '0000'];
  if (simpeIpv6.startsWith('::')) {
    const tmpArr = simpeIpv6.substring(2).split(':');
    for (let i = 0;i < tmpArr.length;i++) {
      arr[i + 8 - tmpArr.length] = (`0000${tmpArr[i]}`).slice(-4);
    }
  } else if (simpeIpv6.endsWith('::')) {
    const tmpArr = simpeIpv6.substring(0, simpeIpv6.length - 2).split(':');
    for (let i = 0;i < tmpArr.length;i++) {
      arr[i] = (`0000${tmpArr[i]}`).slice(-4);
    }
  } else if (simpeIpv6.indexOf('::') >= 0) {
    const tmpArr = simpeIpv6.split('::');
    const tmpArr0 = tmpArr[0].split(':');
    for (let i = 0;i < tmpArr0.length;i++) {
      arr[i] = (`0000${tmpArr0[i]}`).slice(-4);
    }
    const tmpArr1 = tmpArr[1].split(':');
    for (let i = 0;i < tmpArr1.length;i++) {
      arr[i + 8 - tmpArr1.length] = (`0000${tmpArr1[i]}`).slice(-4);
    }
  } else {
    const tmpArr = simpeIpv6.split(':');
    for (let i = 0;i < tmpArr.length;i++) {
      arr[i + 8 - tmpArr.length] = (`0000${tmpArr[i]}`).slice(-4);
    }
  }
  return arr.join(':');
};
export function throttle(fn, delay) {
  let timer;
  return function () {
    if (timer) {
      return;
    }
    timer = setTimeout(() => {
      fn.apply();
      timer = null; // 在delay后执行完fn之后清空timer，此时timer为假，throttle触发可以进入计时器
    }, delay);
  };
}

/**
 * 设置浏览器Cookie的函数
 * @param key Cookie的键
 * @param value Cookie的值
 * @param domain Cookie所适用的域名
 * @param expires Cookie的过期时间  Sat, 02 Aug 2025 07:02:43 GMT
 */
export function setCookie(key: string, value: string, domain?: string, expires?: string): void {
  const expiresStr = expires ? `; expires=${expires}` : '';

  // 构建Cookie字符串
  let cookieString = `${encodeURIComponent(key)}=${encodeURIComponent(value)}${expiresStr}; path=/`;

  // 如果提供了domain，则将其添加到Cookie字符串中
  if (domain) {
    cookieString += `; domain=${domain}`;
  }

  // 设置Cookie
  document.cookie = cookieString;
}

export function compareVersion(v1, v2) {
  const v1parts = v1.split('.');
  const v2parts = v2.split('.');

  for (let i = 0; i < v1parts.length; ++i) {
    if (v2parts.length === i) {
      return 1;
    }

    if (v1parts[i] === v2parts[i]) {
      continue;
    } else if (v1parts[i] > v2parts[i]) {
      return 1;
    } else {
      return -1;
    }
  }

  if (v1parts.length !== v2parts.length) {
    return -1;
  }

  return 0;
}

/**
 * 生成随机数
 * @param {Number} n
 * @param str,默认26位字母及数字
 */
export const random = (n, str = 'abcdefghijklmnopqrstuvwxyz0123456789') => {
  // 生成n位长度的字符串
  // const str = 'abcdefghijklmnopqrstuvwxyz0123456789' // 可以作为常量放到random外面
  let result = '';
  for (let i = 0; i < n; i++) {
    result += str[parseInt(String(Math.random() * str.length), 10)];
  }
  return result;
};

export const fullScreen = (ele) => {
  if (ele.requestFullscreen) {
    ele.requestFullscreen();
  } else if (ele.mozRequestFullScreen) {
    ele.mozRequestFullScreen();
  } else if (ele.webkitRequestFullscreen) {
    ele.webkitRequestFullscreen();
  } else if (ele.msRequestFullscreen) {
    ele.msRequestFullscreen();
  }
};

export const exitFullscreen = (element) => {
  if (document.exitFullscreen) {
    document.exitFullscreen();
  } if (element.msExitFullscreen) {
    element.msExitFullscreen();
  }
};

export function validateCIDR(cidr: string, isIPv6?: boolean) {
  if (!cidr) return true;
  let pattern = '';
  if (isIPv6) {
    pattern = '^([0-9a-f]{1,4}::?){1,7}[0-9a-f]{1,4}/[0-9]{1,3}$';
  } else {
    pattern = '^([0-9]{1,3}.){3}[0-9]{1,3}/([0-9]|[1-2][0-9]|3[0-2])$';
  }
  const regex = new RegExp(pattern);
  return regex.test(cidr);
}

export function countIPsInCIDR(cidr) {
  if (!validateCIDR(cidr)) return;
  const prefixLength = parseInt(cidr.split('/')[1]);
  const hostBits = 32 - prefixLength;
  const totalIPs = Math.pow(2, hostBits);
  return totalIPs;
}
export const getCidrIpNum = (cidr) => {
  const mask = Number(String(cidr).split('/')[1] || 0);
  if (mask <= 0) {
    return 0;
  }
  return Math.pow(2, 32 - mask);
};

const cidrToRange = (cidr: string): [string, string] => {
  const [ip, subnet] = cidr.split('/');
  const ipBinary = ipToBinary(ip);
  const subnetMask = parseInt(subnet, 10);
  const networkBinary = ipBinary.substring(0, subnetMask).padEnd(32, '0');
  const broadcastBinary = networkBinary.substring(0, subnetMask).padEnd(32, '1');

  return [binaryToIp(networkBinary), binaryToIp(broadcastBinary)];
};

const ipToBinary = (ip: string) => ip.split('.').map(octet => parseInt(octet, 10).toString(2)
  .padStart(8, '0'))
  .join('');

const binaryToIp = binary => Array.from({ length: 4 }, (_, i) => parseInt(binary.substring(i * 8, (i + 1) * 8), 2).toString(10)).join('.');

export const cidrContains = (cidrA: string, cidrB: string|string[]) => {
  const [aStart, aEnd] = cidrToRange(cidrA);
  const aStartBinary = parseInt(ipToBinary(aStart), 2);
  const aEndBinary = parseInt(ipToBinary(aEnd), 2);

  const cidrBList = Array.isArray(cidrB) ? cidrB : [cidrB];
  const cidrIpToBinary = cidrBList.reduce<number[]>((pre, cidr) => {
    const range = cidrToRange(cidr);
    range.forEach((ip) => {
      pre.push(parseInt(ipToBinary(ip), 2));
    });
    return pre;
  }, []).sort((a, b) => a - b);
  const rangeStartBinary = cidrIpToBinary[0];
  const rangeEndBinary = cidrIpToBinary[cidrIpToBinary.length - 1];

  return aStartBinary >= rangeStartBinary && aEndBinary <= rangeEndBinary;
};
