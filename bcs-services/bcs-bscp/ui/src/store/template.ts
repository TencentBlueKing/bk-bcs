// 服务实例的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";
import { ITemplatePackageItem, ITemplateSpaceItem } from '../../types/template'

export const useTemplateStore = defineStore('template', () => {
  // 模板空间列表
  const templateSpaceList = ref<ITemplateSpaceItem[]>([])
  // 当前模板空间
  const currentTemplateSpace = ref()
  // 套餐列表
  const packageList = ref<ITemplatePackageItem[]>([])
  // 当前套餐
  const currentPkg = ref()

  return { templateSpaceList, currentTemplateSpace, packageList, currentPkg }
})
