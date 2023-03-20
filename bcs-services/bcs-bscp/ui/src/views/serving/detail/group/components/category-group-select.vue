<script setup lang="ts">
  import { ref, onMounted, watch } from 'vue'
  import { ECategoryType, IGroupItem, IAllCategoryGroupItem } from '../../../../../../types/group';
  import { getAllGroupList } from '../../../../../api/group';

  const props = defineProps<{
    appId: number,
    multiple: boolean,
    value: string|number|number[]
  }>()
  const emits = defineEmits(['change'])

  const groupList = ref<IAllCategoryGroupItem[]>([])
  const groupListLoading = ref(false)
  const groups = ref<string|number|number[]>(props.multiple ? [] : '')

  watch(() => props.value, (val: string|number|number[]) => {
    groups.value = props.multiple ? [ ...<number[]>val ] : val
  }, { immediate: true })

  onMounted(() => {
    getGroupList()
  })

  // 获取全部分组列表
  // @todo 需要一个拉取所有分组下全量分组的接口，这里调试用暂时遍历拉取
  const getGroupList = async() => {
    groupListLoading.value = true
    const query = {
      mode: ECategoryType.Custom,
      start: 0,
      limit: 200,
    }
    const res = await getAllGroupList(props.appId, query)
    groupList.value = res.details
    groupListLoading.value = false
  }

</script>
<template>
  <bk-select :value="groups" :loading="groupListLoading" multiple-mode="tag" :multiple="props.multiple" @change="emits('change', $event)">
    <bk-group v-for="category in groupList" collapsible :key="category.group_category_id" :label="category.group_category_name">
      <bk-option v-for="group in category.groups" :key="group.id" :value="group.id" :label="group.spec.name"></bk-option>
    </bk-group>
  </bk-select>
</template>