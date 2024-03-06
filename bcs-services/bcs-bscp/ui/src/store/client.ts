// 客户端管理的pinia数据
import { ref } from 'vue';
import { defineStore } from 'pinia';

interface IAppData {
  id: number | string;
  spec: {
    name: string;
    config_type: string;
    data_type?: string;
  };
}

export default defineStore('config', () => {
  // 服务详情数据
  const appData = ref<IAppData>({
    id: '',
    spec: {
      name: '',
      config_type: 'file',
    },
  });

  return { appData };
});
