// 配置管理的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";
import { IConfigVersion } from '../../types/config'

export const useConfigStore = defineStore('config', () => {
  // 当前选中版本, 用id为0表示未命名版本
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

  // 是否为版本详情视图
  const versionDetailView = ref(false)

  return { versionData, versionDetailView }
})