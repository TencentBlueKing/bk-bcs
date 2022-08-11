export interface IStyleStack {
  backgroundColor: string[];
  foregroundColor: string[];
  boldDim: string[];
}

export interface IStyleState {
  italic: boolean;
  underline: boolean;
  inverse: boolean;
  hidden: boolean;
  strikethrough: boolean;
}

export interface IDictionary {
  [prop: string]: string;
}

export interface IAnsiConfig {
  ansiTags: IDictionary; // tag配置
  decorators: IDictionary;
}

export interface IStyle {
  italic?: boolean;
  underline?: boolean;
  inverse?: boolean;
  hidden?: boolean;
  strikethrough?: boolean;
  dim?: boolean;
  bold?: boolean;
  backgroundColor?: string;
  foregroundColor?: string;
}

export interface IChunk {
  type: ChunkType;
  value: string;
  style?: IStyle;
}

export interface IParseResult {
  log: string;
  plainText: string;
  chunks: IChunk[];
}

export type ChunkType = 'text' | 'ansi' | 'newline';

export interface IMeta {
  foregroundColor: string;
  html: string;
  plainText: string;
  firstForegroundColor: string;
}
