import dayjs from 'dayjs';
import Cookies from 'js-cookie';

// 字节数转换为对应的显示单位
export const byteUnitConverse = (size: number): string => {
  if (0 <= size && size < 1024) {
    return `${size}B`;
  }
  if (1024 <= size && size < 1024 * 1024) {
    return `${Math.ceil(size / 1024)}KB`;
  }
  if (1024 * 1024 <= size && size < 1024 * 1024 * 1024) {
    return `${(size / (1024 * 1024)).toFixed(1)}MB`;
  }
  if (1024 * 1024 * 1024 <= size) {
    return `${(size / (1024 * 1024 * 1024)).toFixed(1)}GB`;
  }
  return '';
};

// 字符串内容的字节大小
// @notice：edge 79版本才开始支持，发布时间2020-01-15 https://developer.mozilla.org/zh-CN/docs/Web/API/TextEncode
export const stringLengthInBytes = (content: string) => new TextEncoder().encode(content).length;

export const copyToClipBoard = (content: string) => {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(content);
  } else {
    const $textarea = document.createElement('textarea');
    document.body.appendChild($textarea);
    $textarea.style.position = 'fixed';
    $textarea.style.clip = 'rect(0 0 0 0)';
    $textarea.style.top = '10px';
    $textarea.value = content;
    $textarea.select();
    document.execCommand('copy', true);
    document.body.removeChild($textarea);
  }
};

// 时间格式化
export const datetimeFormat = (str: string): string => dayjs(str).format('YYYY-MM-DD HH:mm:ss');

// 获取diff类型
export const getDiffType = (base: string, current: string) => {
  if (base === '' && current !== '') {
    return 'add';
  }
  if (base !== '' && current === '') {
    return 'delete';
  }
  if (base !== '' && current !== '' && base !== current) {
    return 'modify';
  }
  return '';
};

export function getCookie(key: string) {
  return Cookies.get(key);
}

export function setCookie(key: string, val: string, domain: string) {
  Cookies.set(key, val, { domain, expires: 1, path: '/' });
}

export const getTimeRange = (n: number) => {
  const end = new Date();
  const start = new Date();
  start.setTime(start.getTime() - 3600 * 1000 * 24 * n);
  start.setHours(0);
  start.setMinutes(0);
  start.setSeconds(0);
  end.setHours(23);
  end.setMinutes(59);
  end.setSeconds(59);
  return [dayjs(start).format('YYYY-MM-DD HH:mm:ss'), dayjs(end).format('YYYY-MM-DD HH:mm:ss')];
};

export const sortObjectKeysByAscii = (obj: any) => {
  // 获取对象的所有键，并按ASCII码排序
  const sortedKeys = Object.keys(obj).sort((a, b) => a.localeCompare(b, 'en'));
  const sortedObj: any = {};
  sortedKeys.forEach((key) => {
    sortedObj[key] = obj[key];
  });

  return sortedObj;
};

