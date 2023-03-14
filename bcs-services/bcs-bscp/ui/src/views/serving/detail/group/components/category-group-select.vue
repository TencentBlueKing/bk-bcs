<script setup lang="ts">
  import { ref, onMounted, watch } from 'vue'
  import { ICategoryItem, ECategoryType, IGroupItem } from '../../../../../../types/group';
  import { getGroupCategories, getCategoryGroupList } from '../../../../../api/group';

  const props = defineProps<{
    appId: number,
    multiple: boolean,
    value: string|number|number[]
  }>()
  const emits = defineEmits(['change'])

  const categoryList = ref<{ config: ICategoryItem, data: IGroupItem[] }[]>([])
  const categoryLoading = ref(false)
  const groups = ref<string|number|number[]>(props.multiple ? [] : '')

  watch(() => props.value, (val: string|number|number[]) => {
    groups.value = props.multiple ? [ ...<number[]>val ] : val
  }, { immediate: true })

  onMounted(() => {
    getCategoryList()
  })

  // 获取全部分组列表
  // @todo 需要一个拉取所有分组下全量分组的接口，这里调试用暂时遍历拉取
  const getCategoryList = async() => {
    categoryLoading.value = true
    const params = {
      mode: ECategoryType.Custom,
      start: 0,
      limit: 200
    }
    const res = await getGroupCategories(props.appId, params)
    const list: { config: ICategoryItem, data: IGroupItem[] }[] = []
    const groupRes = await Promise.all(res.details.map((item: ICategoryItem) => {
      const query = {
        mode: ECategoryType.Custom,
        start: 0,
        limit: 200,
      }
      return getCategoryGroupList(props.appId, item.id, query)
    }))
    groupRes.forEach((item: { count: number, details: IGroupItem[] }, index: number) => {
      list.push({ config: res.details[index], data: item.details })
    })
    categoryList.value = list
    categoryLoading.value = false
  }

</script>
<template>
  <bk-select :value="groups" :loading="categoryLoading" multiple-mode="tag" :multiple="props.multiple" @change="emits('change', $event)">
    <bk-group v-for="category in categoryList" collapsible :key="category.config.id" :label="category.config.spec.name">
      <bk-option v-for="group in category.data" :key="group.id" :value="group.id" :label="group.spec.name"></bk-option>
    </bk-group>
  </bk-select>
</template>