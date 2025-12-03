import { cloneDeep } from 'lodash';
import { computed, ref, set, shallowRef } from 'vue';

export type Effect = 'PreferNoSchedule' | 'NoExecute' | 'NoSchedule';
export interface IData {
  nodeName: string;
  labels: Record<string, string>,
  taints: Record<string, string>,
  taintsEffect: Record<string, Effect>
  annotations: Record<string, string>
}
export type SettingType = 'labels' | 'taints' | 'annotations';
export type Status = 'remove' | 'add' | 'modify';// remove: 移除, add: 添加, modify: 修改

export default function proxyOperate(prop: SettingType) {
  const originData = shallowRef<IData[]>([]);

  const data = shallowRef<IData[]>([]);
  const tableCol = ref<string[]>([]);
  const colStatusMap = ref<Record<string, Status>>({});// 每一列状态（key的状态）
  const cellStatusMap = ref<Record<string, Status>>({});// 每一个单元格状态(值的状态)
  const effectStatusMap = ref<Record<string, Status>>({});// effect状态

  // 数据map
  const originDataMap = computed<Record<string, IData>>(() => cloneDeep(originData.value).reduce((pre, row) => {
    pre[row.nodeName] = row;
    return pre;
  }, {}));

  // 初始化数据
  function initData(list: any[]) {
    // 把 taints 转换成对象格式（方便数据双向绑定）
    data.value = list.map((item) => {
      const taints = {};
      const taintsEffect = {};
      item.taints.forEach((taintItem) => {
        taints[taintItem.key] = taintItem.value;
        taintsEffect[taintItem.key] = taintItem.effect;
      });
      return {
        nodeName: item.nodeName,
        labels: item.labels,
        annotations: item.annotations,
        taints, // 污点map
        taintsEffect, // 污点effect map
      };
    });
    // 缓存原始数据
    originData.value = cloneDeep(data.value);
    // 初始化表格列
    const cols = new Set<string>();
    data.value.forEach((row) => {
      const item = row[prop] || {}; // 标签或者污点
      Object.keys(item).forEach((key) => {
        if (!cols.has(key)) {
          cols.add(key);
        }
      });
    });
    tableCol.value = [...cols];
  }

  /**
   * 添加标签或污点
   * @param key 键
   * @param value 值
   * @param effect
   */
  function add(labelOrTaintData: Record<string, string> | Array<{key: string, value: string, effect: Effect}>) {
    const parseData = Array.isArray(labelOrTaintData)
      ? labelOrTaintData
      : Object.keys(labelOrTaintData).map(key => ({ key, value: labelOrTaintData[key], effect: undefined }));

    const newKeyValues = {};
    const newKeyEffect = {};
    parseData.forEach((d) => {
      newKeyValues[d.key] = d.value;
      newKeyEffect[d.key] = d.effect;
    });

    data.value.forEach((row) => {
      row[prop] = {
        ...row[prop],
        ...newKeyValues,
      };
      // hack 设置污点effect
      if (prop === 'taints' && Object.keys(newKeyEffect).length) {
        row.taintsEffect = {
          ...row.taintsEffect,
          ...newKeyEffect,
        };
      }
    });

    parseData.forEach((d) => {
      if (d.key) {
        set(colStatusMap.value, d.key, 'add');
        !tableCol.value.includes(d.key) && tableCol.value.unshift(d.key);
      }
    });
  }

  /**
   * 标记删除标签或污点
   * @param key 键
   */
  function remove(key: string) {
    if (!key) return;

    if (colStatusMap.value[key] === 'add') {
      // 新增的列直接删除
      set(colStatusMap.value, key, '');
      data.value.forEach((row) => {
        delete row[prop][key];
      });
      const index = tableCol.value.findIndex(k => k === key);
      if (index > -1) {
        tableCol.value.splice(index, 1);
      }
    } else {
      // 原来有的列，标记删除
      set(colStatusMap.value, key, 'remove');
    }
  }

  /**
   * 恢复删除项
   * @param key
   */
  function undo(key: string) {
    if (!key) return;

    set(colStatusMap.value, key, '');
  }

  function undoValue(nodeName: string, key: string) {
    const row = data.value.find(item => item.nodeName === nodeName);
    if (!key || !row) return;

    if (row?.[prop]?.[key] === originDataMap.value[row.nodeName]?.[prop]?.[key]) {
      set(cellStatusMap.value, `${nodeName}_${key}`, '');
    } else {
      set(cellStatusMap.value, `${nodeName}_${key}`, 'modify');
    }
  }

  /**
   * 标记删除单元格
   * @param nodeName
   * @param key
   */
  function removeValue(nodeName: string, key: string) {
    const row = data.value.find(item => item.nodeName === nodeName);
    if (!row || !key) return;

    set(cellStatusMap.value, `${nodeName}_${key}`, 'remove');
  }

  /**
   * 批量设置某一列
   * @param key
   * @param value
   */
  function setKey(key: string, value: string) {
    if (!key) return;

    data.value.forEach((row) => {
      row[prop][key] = value;

      // 如果当前列不是新增，则判断跟以前值是否一致，不同就标识为修改
      if (colStatusMap.value[key] !== 'add' && value !== originDataMap.value[row.nodeName]?.[prop]?.[key]) {
        set(cellStatusMap.value, `${row.nodeName}_${key}`, 'modify');
      } else {
        set(cellStatusMap.value, `${row.nodeName}_${key}`, '');
      }
    });
  }

  /**
   * 修改单元格的值
   * @param nodeName
   * @param key
   * @param value
   */
  function setValue(nodeName: string, key: string, value: string) {
    const row = data.value.find(item => item.nodeName === nodeName);
    if (!row || !key) return;

    row[prop][key] = value;

    // 如果当前列不是新增，则判断跟以前值是否一致，不同就标识为修改
    if (colStatusMap.value[key] !== 'add' && value !== originDataMap.value[row.nodeName]?.[prop]?.[key]) {
      set(cellStatusMap.value, `${nodeName}_${key}`, 'modify');
    } else {
      set(cellStatusMap.value, `${nodeName}_${key}`, '');
    }
  }

  /**
   * 设置污点effect
   * @param nodeName
   * @param key
   * @param effect
   * @returns
   */
  function setTaintEffect(nodeName: string, key: string, effect: Effect) {
    const row = data.value.find(item => item.nodeName === nodeName);
    if (!row || !key || !effect) return;

    row.taintsEffect[key] = effect;

    // 如果当前列不是新增，则判断跟以前值是否一致，不同就标识为修改
    if (colStatusMap.value[key] !== 'add' && effect !== originDataMap.value[row.nodeName]?.taintsEffect?.[key]) {
      set(effectStatusMap.value, `${nodeName}_${key}`, 'modify');
    } else {
      set(effectStatusMap.value, `${nodeName}_${key}`, '');
    }
  }

  return {
    tableCol,
    originDataMap,
    data,
    colStatusMap,
    cellStatusMap,
    effectStatusMap,
    initData,
    add,
    remove,
    undo,
    undoValue,
    removeValue,
    setKey,
    setValue,
    setTaintEffect,
  };
}
