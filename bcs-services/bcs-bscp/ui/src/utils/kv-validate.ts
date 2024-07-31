import yaml from 'js-yaml';
import * as monaco from 'monaco-editor';

interface IMonacoEditorErrorMarkerItem {
  severity: monaco.MarkerSeverity.Error;
  message: string;
  startLineNumber: number;
  startColumn: number;
  endLineNumber: number;
  endColumn: number;
}

// 校验xml返回错误行和错误信息
export const validateXML = (xmlString: string): IMonacoEditorErrorMarkerItem[] => {
  const parser = new DOMParser();
  const xmlDoc = parser.parseFromString(xmlString, 'text/xml');
  const parserErrors = xmlDoc.getElementsByTagName('parsererror');
  const markers: IMonacoEditorErrorMarkerItem[] = [];
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

// 校验yaml返回错误行和错误信息
export const validateYAML = (yamlString: string): IMonacoEditorErrorMarkerItem[] => {
  const markers: IMonacoEditorErrorMarkerItem[] = [];
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

// 校验json返回错误行和错误信息
export const validateJSON = (jsonString: string): IMonacoEditorErrorMarkerItem[] => {
  const markers: IMonacoEditorErrorMarkerItem[] = [];
  try {
    JSON.parse(jsonString);
  } catch (error) {
    if (error instanceof SyntaxError) {
      const position = getErrorPosition(error, jsonString);
      markers.push({
        severity: monaco.MarkerSeverity.Error,
        message: error.message,
        startLineNumber: position.line,
        startColumn: position.column,
        endLineNumber: position.line,
        endColumn: position.column,
      });
    }
  }
  return markers;
};

function getErrorPosition(error: SyntaxError, jsonString: string) {
  const errorMessage = error.message;
  const match = /at position (\d+)/.exec(errorMessage);
  if (match) {
    const position = parseInt(match[1], 10);
    return calculateLineAndColumn(jsonString, position);
  }
  return { line: 0, column: 0 };
}

function calculateLineAndColumn(jsonString: string, position: number) {
  let line = 1;
  let column = 1;
  for (let i = 0; i < position; i++) {
    if (jsonString[i] === '\n') {
      line += 1;
      column = 1;
    } else {
      column += 1;
    }
  }
  return { line, column };
}
