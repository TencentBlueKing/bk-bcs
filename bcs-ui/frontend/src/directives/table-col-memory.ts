// table-column-memory.ts
import { Table } from 'bk-magic-vue';

import { TABLE_COLUMNS } from '@/common/constant';

// 512KB 容量
const MAX_SIZE = 512 * 1024;

type StorageSchema = Record<string, number[]>;

let tableInstance: InstanceType<typeof Table> | null = null;

const directive = {
  update(el, binding) {
    if (tableInstance) return;
    const  instance: InstanceType<typeof Table> = binding?.value?.instance;
    // 兼容bcs-table和bk-table，bk-table没有doLayout
    tableInstance = !instance?.doLayout ? instance?.$children?.[0] : instance;
    if (!tableInstance?.doLayout) return;

    // 根据组件路径和简易哈希，生成唯一标识键
    const columns = tableInstance.store?.states?.columns || [];
    const key = generateComponentPath(tableInstance) + hashColumns(columns);
    const tableKey = shortHash(key);

    // 初始化列宽
    const savedWidths = loadWidths(tableKey) || [];

    if (savedWidths.length === columns.length) {
      savedWidths.forEach((w, i) => columns[i].width = w);
      tableInstance.doLayout();
    } else {
      updateStorage(tableKey, columns);
    }

    // 注册拖拽事件监听
    const handleDragend = () => {
      updateStorage(tableKey, columns);
    };

    tableInstance._handleHeaderDragend = handleDragend;
    tableInstance.$on('header-dragend', handleDragend);
  },

  unbind() {
    if (tableInstance?._handleHeaderDragend) {
      tableInstance.$off('header-dragend', tableInstance._handleHeaderDragend);
    }
    tableInstance = null;
  },
};

function updateStorage(tableKey: string, columns: any[]) {
  const rawData = localStorage.getItem(TABLE_COLUMNS) || '{}';
  const configs: StorageSchema = JSON.parse(rawData);

  configs[tableKey] = columns.map(c => c.realWidth);

  const isFull = checkStorageCapacity(JSON.stringify(configs));
  // 容量已满
  if (isFull) {
    return;
  }

  try {
    localStorage.setItem(TABLE_COLUMNS, JSON.stringify(configs));
  } catch (e) {
    console.error('LocalStorage 写入失败:', e);
  }
}

// 加载
function loadWidths(key: string): number[] | undefined {
  const data = localStorage.getItem(TABLE_COLUMNS) || '{}';
  let result: number[] | undefined;
  try {
    const configs: StorageSchema = JSON.parse(data);
    result = configs[key];
  } catch (error) {
    console.warn('JSON.parse error:', error);
  }
  return result;
}

// 生成组件层级路径（编译后）（示例：/user/list>UserTable>DataTable）
function generateComponentPath(instance: any): string {
  const path: string[] = [];
  let current = instance.$parent;

  while (current) {
    if (current.$options?.name) {
      path.unshift(current.$options.name);
    }
    current = current.$parent;
  }

  return path.join('>');
}

// 简易列配置哈希
function hashColumns(columns: any[]): string {
  const columnStr = columns
    .map(c => `${c.label}|${c.width}|${c.minWidth}`)
    .join('_');

  return hashCode(columnStr);
}

// 字符串哈希函数
function hashCode(str: string): string {
  let hash = 5381; // 初始质数

  for (let i = 0; i < str.length; i++) {
    hash = (hash << 5) + hash + str.charCodeAt(i);
    hash = hash & 0xFFFFFFFF; // 转换为32位整数
  }

  return (hash >>> 0).toString(36); // 转36进制缩短长度
}

// 生成8位哈希
function shortHash(str: string): string {
  let hash = 5381;
  for (let i = 0; i < str.length; i++) {
    hash = (hash << 5) + hash + str.charCodeAt(i);
    hash |= 0;
  }
  return (hash >>> 0)
    .toString(36)
    .padStart(8, '0')
    .slice(-8)
    .replace(/^0+/, '')   // 只去开头零
    .padEnd(8, '0');      // 补足到8位
}

// 检查存储容量
function checkStorageCapacity(data: string) {
  const byteSize = new Blob([data]).size;

  if (byteSize > MAX_SIZE) { // 精确到字节
    console.warn(`表格列宽存储数据过大（当前：${(byteSize / 1024).toFixed(2)}KB）`);
    return true;
  }
  return false;
}

// 扩展工具方法
export const tableColumnMemoryUtil = {
  getAllConfigs(): StorageSchema {
    let result: StorageSchema = {};
    try {
      result = JSON.parse(localStorage.getItem(TABLE_COLUMNS) || '{}');
    } catch (error) {
      result = {};
    }
    return  result;
  },

  clearConfig(tableKey: string) {
    const configs = this.getAllConfigs();
    delete configs[tableKey];
    localStorage.setItem(TABLE_COLUMNS, JSON.stringify(configs));
  },

  clearAll() {
    localStorage.removeItem(TABLE_COLUMNS);
  },
};

export default (Vue) => {
  Vue.directive('bk-column-memory', directive);
};
