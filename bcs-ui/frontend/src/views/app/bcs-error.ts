import { VueConstructor } from 'vue';

import $store from '@/store';

export interface IConfig {
  filename?: string // 当前组件文件名
  parentFileName?: string // 父组件文件名
  route?: string // 路由
  info?: string // 基本信息
}

export class BcsError extends Error {
  originStack?: string;
  constructor(message, stack?) {
    super(message);
    this.name = this.constructor.name;
    this.stack = stack || (new Error(message)).stack;
    if (typeof Error.captureStackTrace === 'function') {
      this.originStack = this.stack;
      Error.captureStackTrace(this, this.constructor);
    }
  }
}

export function errorHandler(err, config: IConfig = {}) {
  const { filename, parentFileName, route, info } = config;
  console.group(
    '%csomething is wrong with the bcs page, please contact assistant, sorry!☠️',
    'padding: 2px 5px; background: #ea3636; color: #fff; border-radius: 3px 0 0 3px;',
  );
  console.group('%cinfo', 'padding: 2px 5px; background: #ea3636; color: #fff; border-radius: 3px 0 0 3px;');
  console.error(`
  version: ${localStorage.getItem('__bcs_latest_version__')}
  filename: ${filename}
  parentFileName: ${parentFileName}
  route: ${route}
  info: ${info}
  `);
  console.groupEnd();
  console.group('%cstack', 'padding: 2px 5px; background: #ea3636; color: #fff; border-radius: 3px 0 0 3px;');
  console.error(err);
  console.groupEnd();
  console.groupEnd();

  // 数据上报
  window.BkTrace?.startReported({
    module: 'error',
    operation: 'error',
    desc: '前端异常',
    username: $store.state.user.username,
    projectCode: $store.getters.curProjectCode,
    error: {
      version: localStorage.getItem('__bcs_latest_version__'), // 前端版本
      filename, // 错误组件名称
      parentFileName, // 父组件名称
      route, // 路径
      info, // 抛错位置
      stack: `${err?.name}: ${err?.message}`, // 堆栈信息
    },
  }, 'error');
}

export default class BcsErrorPlugin {
  public static install(Vue: VueConstructor) {
    if (process.env.NODE_ENV === 'development') return;// dev模式直接抛异常
    Vue.config.errorHandler = (err, vm, info) => {
      errorHandler(err, {
        filename: vm?.$options?.name,
        parentFileName: vm?.$parent?.$options?.name,
        route: vm?.$route?.fullPath,
        info,
      });
    };
    window.onerror = (message, source, line, column, error) => {
      errorHandler(error, {
        route: window.location.href,
        filename: source,
        info: `${line}, ${column}`,
      });
    };
    window.addEventListener('unhandledrejection', (event) => {
      errorHandler(event);
    });
  }
}
