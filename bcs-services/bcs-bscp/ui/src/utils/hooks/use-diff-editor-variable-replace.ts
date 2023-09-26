import * as monaco from 'monaco-editor'
import { IVariableEditParams } from '../../../types/variable';

interface IReplacedVariableItem {
  range: monaco.IRange;
  key: string;
}

const useDiffEditorVariableReplace = (
  model: monaco.editor.IStandaloneDiffEditor,
  currentVariables: IVariableEditParams[] = [],
  baseVariables: IVariableEditParams[] = []
): monaco.IDisposable => {
  const ins = new VariableReplace(model, currentVariables, baseVariables);
  ins.init()
  return VariableReplace.hoverProvider
}

class VariableReplace {
  editor: monaco.editor.IStandaloneDiffEditor;
  model: monaco.editor.IDiffEditorModel;
  currentVariables: IVariableEditParams[];
  baseVariables: IVariableEditParams[];
  currentReplacedList: IReplacedVariableItem[] = []
  baseReplacedList: IReplacedVariableItem[] = []

  static hoverProvider: monaco.IDisposable

  constructor(editor: monaco.editor.IStandaloneDiffEditor, currentVariables: IVariableEditParams[], baseVariables: IVariableEditParams[]) {
    this.editor = editor
    this.model = <monaco.editor.IDiffEditorModel>editor.getModel()
    this.currentVariables = currentVariables
    this.baseVariables = baseVariables
    this.currentReplacedList = []
    this.baseReplacedList = []
  }

  init () {
    const { modified, original } = this.model
    if (this.currentVariables.length > 0) {
      this.replace(this.editor.getModifiedEditor(), modified, this.currentVariables, this.currentReplacedList)
    }
    if (this.baseVariables.length > 0) {
      this.replace(this.editor.getOriginalEditor(), original, this.baseVariables, this.baseReplacedList)
    }

    this.registerHoverProvider()
  }

  /**
   * 找到monaco编辑器中文本内容中的变量，替换为变量值，鼠标hover到变量值时，显示变量名
   * 将配置内容按照行分割，遍历变量列表，将每行中的所有变量名替换为变量值，并记录替换后内容的行、列位置
   */
  replace(editor: monaco.editor.ICodeEditor, model: monaco.editor.ITextModel, variables: IVariableEditParams[], replacedList: IReplacedVariableItem[]) {
    const variablesMap: { [key: string]: string } = {}
    variables.forEach(v => {
      variablesMap[v.name] = v.default_val
    })
    const textList = model.getValue().split('\n')
    textList.forEach((text, index) => {
      const lineNumber = index + 1
      const { replacedText, variablePos } = this.getReplacedData(text, variablesMap)
      textList[index] = replacedText
      if (variablePos.length > 0) {
        variablePos.forEach(pos => {
          const { name, start, end } = pos
          replacedList.push({
            range: { startLineNumber: lineNumber, startColumn: start, endLineNumber: lineNumber, endColumn: end },
            key: name
          })
        })
      }
    })
    model.setValue(textList.join('\n'))
    this.highlightVariables(editor, replacedList)
  }

  // 递归匹配文本中变量名，逐一替换为变量值，并计算文本替换后的变量值所在新内容的行、列位置以及替换后的文本
  getReplacedData (text: string, variablesMap: { [key: string]: string }) {
    let replacedText = text
    const reg = new RegExp('{{\\s*\\.([bB][kK]_[bB][sS][cC][pP]_[A-Za-z0-9_]*)\\s*}}')
    const variablePos: { name: string; start: number; end: number }[] = []
    let match = text.match(reg)
    while (match && match.length > 0) {
      const name = match[1]
      if (name in variablesMap) {
        const val = variablesMap[name]
        const index = <number>match.index + 1
        replacedText = replacedText.replace(reg, val)
        variablePos.push({ name, start: index, end: index + val.length })
        match = replacedText.match(reg)
      }
    }
    return { replacedText, variablePos }
  }

  highlightVariables (editor: monaco.editor.ICodeEditor, replacedList: IReplacedVariableItem[]) {
    const configs = replacedList.map(variable => {
      return {
        range: variable.range,
        options: {
          inlineClassName: "template-variable-item",
        }
      }
    })
    editor.createDecorationsCollection(configs);
  }

  registerHoverProvider () {
    const self = this
    if (VariableReplace.hoverProvider) {
      VariableReplace.hoverProvider.dispose()
    }
    VariableReplace.hoverProvider = monaco.languages.registerHoverProvider('plaintext', {
      provideHover(model, position) {
        const { modified, original } = self.model
        const { lineNumber, column } = position

        if (model.uri.toString() === modified.uri.toString()) {
          return self.getProviderConfig(self.currentReplacedList, lineNumber, column)
        } else if (model.uri.toString() === original.uri.toString()) {
          return self.getProviderConfig(self.baseReplacedList, lineNumber, column)
        }
      }
    })
  }

  getProviderConfig (replacedList: IReplacedVariableItem[] = [], lineNumber: number, column: number) {
    const variable = replacedList.find(v => v.range.startLineNumber === lineNumber && v.range.startColumn <= column && column <= v.range.endColumn)
    if (variable) {
      return {
        range: variable.range,
        contents: [
          {value: ''}, // 去掉标题
          {
            value: variable.key,
          }
        ]
      }
    }
  }
}

export default useDiffEditorVariableReplace;
