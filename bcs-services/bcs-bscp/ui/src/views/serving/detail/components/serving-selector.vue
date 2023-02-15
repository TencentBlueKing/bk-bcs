<script setup lang="ts">
  import { defineProps, defineEmits, ref, Ref, watch ,onMounted } from 'vue'
  import { IServingItem } from '../../../../types'
  import { getAppList } from "../../../../api";

  const props = defineProps<{
    value: number,
    bkBizId: number
  }>()

  defineEmits(['change'])

  const servingList = ref([]) as Ref<IServingItem[]>
  const loading = ref(false)
  const localVal = ref(props.value)

  watch(() => props.value, (val) => {
    localVal.value = val
  })

  onMounted(() => {
    loadServingList()
  })

  const loadServingList = async() => {
    loading.value = true;
    try {
      const resp = await getAppList(props.bkBizId);
      // @ts-ignore
      servingList.value = resp.details
    } catch (e) {
      console.error(e);
    } finally {
        loading.value = false;
    }
      
  };
</script>
<template>
  <bk-select v-model="localVal" filterable :clearable="false" @change="$emit('change', $event)">
    <bk-option
      v-for="item in servingList"
      :key="item.id"
      :value="item.id"
      :label="item.spec.name">
    </bk-option>
  </bk-select>
</template>
