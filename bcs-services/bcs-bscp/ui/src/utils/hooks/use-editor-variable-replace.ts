import * as monaco from 'monaco-editor';
import { IVariableEditParams } from '../../../types/variable';

const useEditorVariableReplace = (
  model: monaco.editor.IStandaloneCodeEditor,
  variables: IVariableEditParams[],
): monaco.IDisposable => {
  const ins = new VariableReplace(model, variables);
  ins.replace();
  return VariableReplace.hoverProvider;
};

class VariableReplace {
  public static hoverProvider: monaco.IDisposable;
  public editor: monaco.editor.IStandaloneCodeEditor;
  public model: monaco.editor.ITextModel;
  public variables: IVariableEditParams[];
  public replacedList: { range: monaco.IRange; key: string }[] = [];

  public constructor(editor: monaco.editor.IStandaloneCodeEditor, variables: IVariableEditParams[]) {
    this.editor = editor;
    this.model = editor.getModel() as monaco.editor.ITextModel;
    this.variables = variables;
    this.replacedList = [];
  }

  /**
   * 找到monaco编辑器中文本内容中的变量，替换为变量值，鼠标hover到变量值时，显示变量名
   * 将配置内容按照行分割，遍历变量列表，将每行中的所有变量名替换为变量值，并记录替换后内容的行、列位置
   */
  public replace() {
    const variablesMap: { [key: string]: string } = {};
    this.variables.forEach((v) => {
      variablesMap[v.name] = v.default_val;
    });
    const textList = this.model.getValue().split('\n');
    textList.forEach((text, index) => {
      const lineNumber = index + 1;
      const { replacedText, variablePos } = this.getReplacedData(text, variablesMap);
      textList[index] = replacedText;
      if (variablePos.length > 0) {
        variablePos.forEach((pos) => {
          const { name, start, end } = pos;
          this.replacedList.push({
            range: { startLineNumber: lineNumber, startColumn: start, endLineNumber: lineNumber, endColumn: end },
            key: name,
          });
        });
      }
    });
    this.model.setValue(textList.join('\n'));
    this.highlightVariables();
    this.registerHoverProvider();
  }

  // 递归匹配文本中变量名，逐一替换为变量值，并计算文本替换后的变量值所在新内容的行、列位置以及替换后的文本
  public getReplacedData(text: string, variablesMap: { [key: string]: string }) {
    let replacedText = text;
    const reg = new RegExp('{{\\s*\\.([bB][kK]_[bB][sS][cC][pP]_[A-Za-z0-9_]*)\\s*}}');
    const variablePos: { name: string; start: number; end: number }[] = [];
    let match = text.match(reg);
    while (match && match.length > 0) {
      const name = match[1];
      if (name in variablesMap) {
        const val = variablesMap[name];
        const index = (match.index as number) + 1;
        replacedText = replacedText.replace(reg, val);
        variablePos.push({ name, start: index, end: index + val.length });
        match = replacedText.match(reg);
      }
    }
    return { replacedText, variablePos };
  }

  public highlightVariables() {
    const configs = this.replacedList.map(variable => ({
      range: variable.range,
      options: {
        inlineClassName: 'template-variable-item',
      },
    }));
    this.editor.createDecorationsCollection(configs);
  }

  public registerHoverProvider() {
    const self = this;
    if (VariableReplace.hoverProvider) {
      VariableReplace.hoverProvider.dispose();
    }
    VariableReplace.hoverProvider = monaco.languages.registerHoverProvider('*', {
      provideHover(model, position) {
        const { lineNumber, column } = position;
        const variable = self.replacedList.find(v => v.range.startLineNumber === lineNumber
          && v.range.startColumn <= column && column <= v.range.endColumn);
        if (variable) {
          return {
            range: variable.range,
            contents: [
              { value: '' }, // 去掉标题
              {
                value: variable.key,
              },
            ],
          };
        }
      },
    });
  }
}

export default useEditorVariableReplace;
