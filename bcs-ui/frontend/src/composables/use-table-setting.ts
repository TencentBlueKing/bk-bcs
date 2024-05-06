import { ref } from 'vue';
import VueI18n from 'vue-i18n';

import { CLUSTER_NODE_TABLE_COL } from '@/common/constant';

export interface IField {
  id: string;
  label: string | VueI18n.TranslateResult;
  disabled?: boolean;
  defaultChecked?: boolean;
}
// 表格设置功能
export default function useTableSetting(fieldsData: Array<IField>) {
  const storageFields = JSON.parse(localStorage.getItem(CLUSTER_NODE_TABLE_COL) || '{}');
  const tableSetting = ref({
    size: 'medium',
    fields: fieldsData,
    selectedFields: fieldsData.filter((item) => {
      if (item.disabled) return true;
      if (item.id in storageFields) return storageFields[item.id];
      return !!item.defaultChecked;
    }),
  });
  const isColumnRender = id => tableSetting.value.selectedFields.some(item => item.id === id);
  const handleSettingChange = ({ size, fields }) => {
    tableSetting.value.size = size;
    tableSetting.value.selectedFields = fields;
    const storageData = fieldsData.reduce((pre, cur) => {
      pre[cur.id] = fields.some(item => item.id === cur.id);
      return pre;
    }, {});
    localStorage.setItem(CLUSTER_NODE_TABLE_COL, JSON.stringify(storageData));
  };
  return {
    tableSetting,
    isColumnRender,
    handleSettingChange,
  };
}
