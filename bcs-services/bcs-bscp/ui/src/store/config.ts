// 配置管理的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";
import { IConfigVersion } from '../../types/config'

export const useConfigStore = defineStore('config', () => {
  const versionData = ref<IConfigVersion>({
    id: 0,
    attachment: {
      app_id: 0,
      biz_id: 0
    },
    revision: {
      create_at: '',
      creator: ''
    },
    spec: {
      name: '',
      memo: ''
    }
  })

  return { versionData }
})