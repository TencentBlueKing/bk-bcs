// 客户端管理的pinia数据
import { ref } from 'vue';
import { defineStore } from 'pinia';
import { IClientSearchParams } from '../../types/client';

export default defineStore('client', () => {
  // 选择的查询条件
  const searchQuery = ref<{
    last_heartbeat_time: number;
    search: IClientSearchParams;
  }>({
    last_heartbeat_time: 1,
    search: {},
  });
  // 配置示例-基本信息
  const basicInfo = ref<{
    name: string;
    serviceType: string;
    id: string | number;
  }>({
    name: '',
    serviceType: '',
    id: '',
  });

  return { searchQuery, basicInfo };
});
