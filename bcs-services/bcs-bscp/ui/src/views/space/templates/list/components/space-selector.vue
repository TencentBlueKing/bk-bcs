<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../store/global'
  import { useTemplateStore } from '../../../../../store/template'
  import { ITemplateSpaceItem } from '../../../../../../types/template'
  import { ICommonQuery } from '../../../../../../types/index'
  import { getTemplateSpaceList } from '../../../../../api/template'

  const { spaceId } = storeToRefs(useGlobalStore())
  const templateStore = useTemplateStore()

  const loading = ref(false)
  const spaceList = ref<ITemplateSpaceItem[]>([])

  onMounted(() => {
    loadList()
  })

  const loadList = async () => {
    loading.value = true
    const params: ICommonQuery = {
      start: 0,
      limit: 100
      // all: true
    }
    const res = await getTemplateSpaceList(spaceId.value, params)
    spaceList.value = res.details
    templateStore.$patch((state) => {
      state.templateSpaceList = res.details
      state.currentTemplateSpace = res.details[0]?.id
    })
    loading.value = false
  }

</script>
<template>
    <bk-select placeholder="请选择空间">
      <bk-option v-for="item in spaceList" :key="item.id" :label="item.spec.name" :value="item.id"></bk-option>
    </bk-select>
</template>
