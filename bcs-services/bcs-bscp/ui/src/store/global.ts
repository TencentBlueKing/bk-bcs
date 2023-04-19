// 应用全部的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";
import { ISpaceDetail } from '../../types/index';

export const useGlobalStore = defineStore('global', () => {
  const spaceId = ref('') // 空间id
  const spaceList = ref<ISpaceDetail[]>([])
  const showApplyPermDialog = ref(false)
  const permissionQuery = ref({
    biz_id: '',
    basic: {
      type: '',
      action: '',
      resource_id: '',
    },
    gen_apply_url: true,
  })

  return { spaceId, spaceList, showApplyPermDialog, permissionQuery }
})