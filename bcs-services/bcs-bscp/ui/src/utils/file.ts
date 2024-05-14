import { byteUnitConverse } from './index';

// 将文本转换为二进制文件
export const transDataToFile = (data: string, name: string) => {
  const blob = new Blob([data], { type: 'application/octet-stream' });
  const file = new File([blob], name, { type: 'application/octet-stream' });
  return file;
};

// 将File对象转换为json
export const transFileToObject = (file: File) => {
  const { name, size } = file;
  return { name, size: byteUnitConverse(size) };
};

// 文件下载
export const fileDownload = (content = '', name = '', isBlob = true) => {
  let url = '';
  if (isBlob) {
    // 定义MIME类型为二进制流，避免浏览器为文件名称默认添加.txt后缀
    const blob = new Blob([content], { type: 'application/octet-stream' });
    url = window.URL.createObjectURL(blob);
  } else {
    url = content;
  }
  const eleLink = document.createElement('a');
  eleLink.style.display = 'none';
  eleLink.href = url;
  eleLink.setAttribute('download', name);

  // hack HTML5 download attribute
  if (typeof eleLink.download === 'undefined') {
    eleLink.setAttribute('target', '_blank');
  }
  document.body.appendChild(eleLink);
  eleLink.click();
  document.body.removeChild(eleLink);
  if (isBlob) {
    window.URL.revokeObjectURL(url);
  }
};
