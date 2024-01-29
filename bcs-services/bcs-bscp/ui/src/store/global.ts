// 应用全部的pinia数据
import { ref } from 'vue';
import { defineStore } from 'pinia';
import { ISpaceDetail, IPermissionResource, IPermissionQueryResourceItem } from '../../types/index';

export default defineStore('global', () => {
  const bscpVersion = ref(''); // 产品版本号
  const spaceId = ref(''); // 空间id
  const spaceFeatureFlags = ref<{ [key: string]: boolean }>({}); // 空间的配置开关
  const spaceList = ref<ISpaceDetail[]>([]);
  // @ts-ignore
  const showNotice = ref(window.ENABLE_BK_NOTICE === 'true'); // 是否展示消息通知
  const showApplyPermDialog = ref(false); // 资源无权限申请弹窗
  const showPermApplyPage = ref(false); // 无业务查看权限时，申请页面
  const applyPermUrl = ref(''); // 跳转到权限中心的申请链接
  const applyPermResource = ref<IPermissionResource[]>([]); // 无权限提示页的action
  const permissionQuery = ref<{ resources: IPermissionQueryResourceItem[] }>({
    resources: [],
  });

  return {
    bscpVersion,
    spaceId,
    spaceFeatureFlags,
    spaceList,
    showNotice,
    showApplyPermDialog,
    showPermApplyPage,
    applyPermUrl,
    applyPermResource,
    permissionQuery,
  };
});
