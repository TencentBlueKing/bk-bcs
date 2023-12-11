import { Message } from 'bkui-vue';
import yaml from 'js-yaml';

export const validateXML = (xmlString: string) => {
  // 创建一个新的DOMParser实例
  const parser = new DOMParser();
  // 解析XML字符串
  const xmlDoc = parser.parseFromString(xmlString, 'text/xml');
  // 检查解析是否出错
  const parseError = xmlDoc.getElementsByTagName('parsererror');
  // 如果有错误，返回错误信息
  if (parseError.length > 0) {
    Message({
      message: 'xml格式错误',
      theme: 'error',
    });
    return false; // XML不合法
  }
  return true; // XML合法
};

export const validateJSON = (jsonString: string) => {
  try {
    // 尝试解析JSON文本
    JSON.parse(jsonString);
    return true;
  } catch (e) {
    Message({
      message: 'json格式错误',
      theme: 'error',
    });
    return false;
  }
};

export const validateYAML = (yamlString: string) => {
  try {
    yaml.load(yamlString, 'utf8');
    return true; // YAML合法
  } catch (error) {
    Message({
      message: 'yaml格式错误',
      theme: 'error',
    });
    return false; // YAML不合法
  }
};
