// 应用全部的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";
import { ISpaceDetail, IPermissionResource } from '../../types/index';

export const useGlobalStore = defineStore('global', () => {
  const spaceId = ref('') // 空间id
  const spaceList = ref<ISpaceDetail[]>([])
  const showApplyPermDialog = ref(false) // 资源无权限申请弹窗
  const showPermApplyPage = ref(false) // 无业务查看权限时，申请页面
  const applyPermUrl = ref('') // 跳转到权限中心的申请链接
  const applyPermResource = ref<IPermissionResource[]>([]) // 无权限提示页的action
  const permissionQuery = ref({
    resources: []
  })

  return {
    spaceId,
    spaceList,
    showApplyPermDialog,
    showPermApplyPage,
    applyPermUrl,
    applyPermResource,
    permissionQuery
  }
})
