import yaml from 'js-yaml';
import * as monaco from 'monaco-editor';

// 校验xml返回错误行和错误信息
export const validateXML = (xmlString: string) => {
  const parser = new DOMParser();
  const xmlDoc = parser.parseFromString(xmlString, 'text/xml');
  const parserErrors = xmlDoc.getElementsByTagName('parsererror');
  const markers = [];
  if (parserErrors.length > 0) {
    const error = parserErrors[0];
    const errorMsg = error.textContent!.trim();
    const match = errorMsg.match(/error on line (\d+) at column (\d+): (.*)/);
    if (match) {
      const lineNumber = parseInt(match[1], 10) - 1;
      const columnNumber = parseInt(match[2], 10);
      const errorMessage = match[3].trim();
      markers.push({
        severity: monaco.MarkerSeverity.Error,
        message: errorMessage,
        startLineNumber: lineNumber,
        startColumn: columnNumber,
        endLineNumber: lineNumber,
        endColumn: columnNumber,
      });
    }
  }
  return markers;
};

// 校验json 编辑器自带校验错误行和错误信息 只需返回是否正确
export const validateJSON = (jsonString: string) => {
  try {
    // 尝试解析JSON文本
    JSON.parse(jsonString);
    return {
      result: true,
      message: '',
    };
  } catch (e) {
    return {
      result: false,
      message: e,
    };
  }
};

// 校验yaml返回错误行和错误信息
export const validateYAML = (yamlString: string) => {
  const markers = [];
  try {
    yaml.load(yamlString, 'utf8');
  } catch (e: any) {
    markers.push({
      severity: monaco.MarkerSeverity.Error,
      message: e.reason,
      startLineNumber: e.mark.line,
      startColumn: e.mark.column,
      endLineNumber: e.mark.line,
      endColumn: e.mark.column,
    });
  }
  return markers;
};
