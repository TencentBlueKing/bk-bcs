// 配置管理的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";
import { IConfigVersion } from '../../types/config'
import { GET_UNNAMED_VERSION_DATE } from '../constants/config'

export const useConfigStore = defineStore('config', () => {

  // 当前选中版本, 用id为0表示未命名版本
  const versionData = ref<IConfigVersion>(GET_UNNAMED_VERSION_DATE())

  // 是否为版本详情视图，切换服务后保持当前视图
  const versionDetailView = ref(false)

  // 是否需要刷新版本列表标识，配置生成版本、发布版本、调整分组上线之后需要更新版本列表
  const refreshVersionListFlag = ref(false)

  return { versionData, versionDetailView, refreshVersionListFlag }
})