// ansi编码日志解析

import { IAnsiConfig, IChunk, ChunkType, IParseResult, IStyle, IStyleStack, IStyleState, IMeta } from '../types/types';
import superSplit from './super-split';
import ansiTags from './ansi-tags';
import decorators from './decorator-names';

export default class AnsiParser {
  styleStack: IStyleStack = {
    foregroundColor: [],
    backgroundColor: [],
    boldDim: [],
  };
  styleState: IStyleState = {
    italic: false,
    underline: false,
    inverse: false,
    hidden: false,
    strikethrough: false,
  };
  private config: IAnsiConfig = {
    ansiTags,
    decorators,
  };
  // constructor () {
  //     // this.config = Object.assign(this.config, config)
  // }

  /**
     * 解析ansi编码的日志
     * @param log
     */
  parse(log: string) {
    const plainText = this.getPlainText(log);
    const { words, ansies } = this.atomize(log);
    const result: IParseResult = {
      log,
      plainText,
      chunks: [],
    };

    words.forEach((word) => {
      // 换行符
      if (word === '\n') {
        const chunk = this.bundle('newline', word);
        result.chunks.push(chunk);
        return;
      }

      // 文本
      if (!ansies.includes(word)) {
        const chunk = this.bundle('text', word);
        result.chunks.push(chunk);
        return;
      }

      // ansi字符
      const { ansiTags, decorators } = this.config;
      const ansiTag = ansiTags[word]; // ansi 字符映射表
      const decorator = decorators[ansiTag] || ''; // 开闭区间

      if (decorator.endsWith('Open')) {
        // 开区间
        const index = decorator.lastIndexOf('Open');
        const prop = decorator.substring(0, index);
        if (this.styleStack[prop as keyof IStyleStack]) {
          // 颜色
          this.styleStack[prop as keyof IStyleStack].push(ansiTag);
        } else if (this.styleState[prop as keyof IStyleState]) {
          // 状态
          this.styleState[prop as keyof IStyleState] = true;
        }
      } else if (decorator.endsWith('Close')) {
        // 闭区间
        const index = decorator.lastIndexOf('Close');
        const prop = decorator.substring(0, index);
        if (this.styleStack[prop as keyof IStyleStack]) {
          // 颜色
          this.styleStack[prop as keyof IStyleStack].pop();
        } else if (this.styleState[prop as keyof IStyleState]) {
          // 状态
          this.styleState[prop as keyof IStyleState] = false;
        }
      }

      if (decorator === 'reset') {
        this.styleState.strikethrough = false;
        this.styleState.inverse = false;
        this.styleState.italic = false;
        this.styleStack.boldDim = [];
        this.styleStack.backgroundColor = [];
        this.styleStack.foregroundColor = [];
      }

      const chunk = this.bundle('ansi', word);

      result.chunks.push(chunk);
    });
    return result;
  }

  /**
     * 转换为HTML
     * @param log
     * @returns
     */
  transformToHtml(log: string) {
    const { chunks, plainText } = this.parse(log);
    return chunks.reduce<IMeta>((pre, current) => {
      if (current.type === 'text') {
        pre.html += `<span style="color: ${pre.foregroundColor || '#fff'}">${String(current.value).replace(/\s/g, '&nbsp;')}</span>`;
      } else if (current.type === 'ansi') {
        const { foregroundColor = '' } = current.style || {};
        pre.foregroundColor = foregroundColor;
      }
      // todo：暂时保留第一个颜色，用于换行的颜色的临时处理
      if (!pre.firstForegroundColor) {
        pre.firstForegroundColor = pre.foregroundColor;
      }
      return pre;
    }, { foregroundColor: '', html: '', plainText, firstForegroundColor: '' });
  }

  /**
     * 生成解析后的bundle
     * @param type
     * @param value
     * @returns
     */
  bundle(type: ChunkType, value: string) {
    const chunk: IChunk = {
      type,
      value,
    };
    if (type === 'ansi' || type === 'text') {
      const style: IStyle = {};

      const foregroundColor = this.getForegroundColor();
      const backgroundColor = this.getBackgroundColor();
      const dim = this.getDim();
      const bold = this.getBold();

      if (foregroundColor) {
        style.foregroundColor = foregroundColor;
      }

      if (backgroundColor) {
        style.backgroundColor = backgroundColor;
      }

      if (dim) {
        style.dim = dim;
      }

      if (bold) {
        style.bold = bold;
      }

      if (this.styleState.italic) {
        style.italic = true;
      }

      if (this.styleState.underline) {
        style.underline = true;
      }

      if (this.styleState.inverse) {
        style.inverse = true;
      }

      if (this.styleState.strikethrough) {
        style.strikethrough = true;
      }

      chunk.style = style;
    }
    return chunk;
  }

  /**
     * 获取去除ansi码后的文本
     * @param log
     * @returns
     */
  getPlainText(log: string) {
    if (typeof log !== 'string') {
      throw new TypeError(`Expected a \`string\`, got \`${typeof log}\``);
    }
    return log.replace(this.ansiRegex(), '');
  }

  /**
     * 获取ansi字符匹配的正则
     * @param flag 匹配模式
     * @returns
     */
  ansiRegex(flag = 'g') {
    const pattern = [
      '[\\u001B\\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]*)*)?\\u0007)',
      '(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-ntqry=><~]))',
    ].join('|');

    return new RegExp(pattern, flag);
  }

  /**
     * 转换为数组形势
     * @param log
     * @returns
     */
  atomize(log: string) {
    const ansies = Array.from(new Set(log.match(this.ansiRegex())));
    const words = superSplit(log, ansies.concat(['\n'])) as string[];
    return {
      ansies,
      words,
    };
  }

  /**
     * 获取当前堆栈bundle对应的字体颜色
     * @returns
     */
  getForegroundColor() {
    if (this.styleStack.foregroundColor.length > 0) {
      return this.styleStack.foregroundColor[this.styleStack.foregroundColor.length - 1];
    }
    return false;
  }

  /**
     * 获取当前堆栈bundle对应的背景颜色
     * @returns
     */
  getBackgroundColor() {
    if (this.styleStack.backgroundColor.length > 0) {
      return this.styleStack.backgroundColor[this.styleStack.backgroundColor.length - 1];
    }
    return false;
  }

  getDim() {
    return this.styleStack.boldDim.includes('dim');
  }

  /**
     * 获取当前堆栈bundle对应的字体粗细
     * @returns
     */
  getBold() {
    return this.styleStack.boldDim.includes('bold');
  }
}
