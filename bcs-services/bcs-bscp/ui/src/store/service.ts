// 服务实例的pinia数据
import { ref } from 'vue';
import { defineStore, storeToRefs } from 'pinia';
import useGlobalStore from './global';

interface IAppData {
  id: number | string;
  spec: {
    name: string;
  };
}
const { spaceId, permissionQuery, showApplyPermDialog } = storeToRefs(useGlobalStore());

export default defineStore('service', () => {
  // 服务详情数据
  const appData = ref<IAppData>({
    id: '',
    spec: {
      name: '',
    },
  });
  const permCheckLoading = ref(false);
  const hasEditServicePerm = ref(false);

  const checkPermBeforeOperate = (perm: string) => {
    if (perm === 'update' && !hasEditServicePerm.value) {
      permissionQuery.value = {
        resources: [
          {
            biz_id: spaceId.value,
            basic: {
              type: 'app',
              action: 'update',
              resource_id: appData.value.id,
            },
          },
        ],
      };
      showApplyPermDialog.value = true;
      return false;
    }
    return true;
  };

  return { appData, permCheckLoading, hasEditServicePerm, checkPermBeforeOperate };
});
