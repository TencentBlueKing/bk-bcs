// 应用全部的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";
import { ISpaceDetail } from '../../types/index';

export const useGlobalStore = defineStore('global', () => {
  const spaceId = ref('') // 空间id
  const spaceList = ref<ISpaceDetail[]>([])
  const showApplyPermDialog = ref(false) // 资源无权限申请弹窗
  const showSpacePermApply = ref(false) // 无业务查看权限时，申请页面
  const applyPermUrl = ref('') // 跳转到权限中心的申请链接
  const permissionQuery = ref({
    biz_id: '',
    basic: {
      type: '',
      action: '',
      resource_id: '',
    },
    gen_apply_url: true,
  })

  return {
    spaceId,
    spaceList,
    showApplyPermDialog,
    showSpacePermApply,
    applyPermUrl,
    permissionQuery
  }
})