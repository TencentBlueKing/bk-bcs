import menusData, { IMenu } from './menus';
import { useCluster, useConfig } from '@/composables/use-app';
import { computed, ref } from 'vue';

export default function useMenu() {
  const menus = ref<IMenu[]>(menusData);
  // 共享集群禁用菜单
  const { _INTERNAL_ } = useConfig();
  const { isSharedCluster } = useCluster();
  const disabledMenuIDs = computed(() => {
    const disabledIDs: string[] = [];
    if (isSharedCluster.value) {
      disabledIDs.push(...[
        'DAEMONSET',
        'PERSISTENTVOLUME',
        'STORAGECLASS',
        'HPA',
        'CRD',
        'CUSTOMOBJECT',
      ]);
    }
    if (_INTERNAL_.value) {
      disabledIDs.push('CLOUDTOKEN');
    }
    return disabledIDs;
  });

  return {
    menus,
    disabledMenuIDs,
  };
}
