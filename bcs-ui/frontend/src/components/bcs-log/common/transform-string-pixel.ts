import CharMap from './chart-pixel-map';

interface IConfig {
  size: number; // 字体大小
  font: string; // 字体 目前仅支持courier new字体
  bold: boolean; // 是否粗体
  italic: boolean; // 是否斜体
}

type KeyType<T extends (string | number)> = T;

// 字符串转像素
export default class TransformStringPixel {
  config: IConfig = {
    size: 12,
    font: 'courier new',
    bold: false,
    italic: false,
  };

  constructor(config?: IConfig) {
    this.config = Object.assign(this.config, config || {});
  }

  getStringPixel(str: string) {
    const { bold, italic, font, size } = this.config;
    // 像素索引，如："#": [60,60,60,60] 表示 [默认，粗体，斜体，其他] 对应的像素大小
    const variant = (bold ? 1 : 0) + (italic ? 2 : 0);
    // 字符像素映射map
    const map = CharMap[font as KeyType<'courier new'>] || CharMap['courier new'];
    let total = 0;
    for (const char of str) {
      let width = 0;
      // eslint-disable-next-line no-control-regex
      if (/[\x00-\x1F]/.test(char)) {
        // 特殊字符
        width = 60;
      } else if (map[char]) {
        // 可输入字符
        width = map[char][variant];
      } else {
        // 其他字符
        width = 60;
      }
      total += width;
    }
    return total * (size / 100);
  }
}
