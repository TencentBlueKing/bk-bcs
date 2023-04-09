// 应用全部的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";

export const useGlobalStore = defineStore('global', () => {
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

  return { showApplyPermDialog, permissionQuery }
})