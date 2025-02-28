import { ref } from 'vue';

import { getDisktypes } from '@/api/modules/cluster-manager';

export type Disk = {
  id: string;
  name: string;
};

export type DiskItem = {
  diskType: string;
  diskSize: string;
  fileSystem?: string;
  autoFormatAndMount?: boolean;
  mountTarget?: string;
};

// 不同组件共用
export const diskMap = ref<Record<string, string>>({});

export function useDisk() {
  const isLoading = ref(false);
  const systemDisks = ref<Array<Disk>>([]);
  const dataDisks = ref<Array<Disk>>([]);

  async function getDisks(params) {
    isLoading.value = true;
    const res = await getDisktypes(params, { cancelPrevious: true }).catch(() => []);
    const result = res || [];
    systemDisks.value = result.filter(item => item.diskUsage === 'SYSTEM_DISK').map(v => ({
      id: v.diskType,
      name: v.diskTypeName,
    }));
    dataDisks.value = result.filter(item => item.diskUsage === 'DATA_DISK').map(v => ({
      id: v.diskType,
      name: v.diskTypeName,
    }));

    const resultMap = result.reduce((acc, cur) => {
      acc[cur.diskType] = cur.diskTypeName;
      return acc;
    }, {});
    // 汇总，添加多种机型时不会把之前的覆盖掉
    diskMap.value = {
      ...diskMap.value,
      ...resultMap,
    };
    isLoading.value = false;
  }

  return {
    isLoading,
    systemDisks,
    dataDisks,
    diskMap,
    getDisks,
  };
}
