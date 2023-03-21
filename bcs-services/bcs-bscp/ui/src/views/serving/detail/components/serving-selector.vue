<script setup lang="ts">
  import { ref, Ref, watch ,onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { IServingItem } from '../../../../types'
  import { getAllApp } from "../../../../api";

  interface IServingGroupItem {
    space_id: number;
    space_name: string;
    space_type_id: string;
    space_type_name: string;
    children: Array<IServingItem>;
  }

  const route = useRoute()
  const router = useRouter()

  const props = defineProps<{
    value: number
  }>()

  defineEmits(['change'])

  const servingList = ref([]) as Ref<IServingGroupItem[]>
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
      const resp = await getAllApp();
      const groupedList: Array<IServingGroupItem> = []
      // @ts-ignore
      resp.details.forEach(service => {
        const group = groupedList.find(item => item.space_id === service.space_id)
        if (group) {
          group.children.push(service)
        } else {
          const {space_id, space_name, space_type_id, space_type_name  } = service
          groupedList.push({
            space_id,
            space_name,
            space_type_id,
            space_type_name,
            children: [{ ...service }]
          })
        }
      })
      servingList.value = groupedList
    } catch (e) {
      console.error(e);
    } finally {
        loading.value = false;
    }
  };

  const handleAppChange = (id: number) => {
    const group = servingList.value.find(group => {
      return group.children.find(item => item.id === id)
    })
    if (group) {
      router.push({ name: <string>route.name, params: { spaceId: group.space_id, appId: id } })
    }
  }
</script>
<template>
<div>
  <bk-select v-model="localVal" :filterable="true" :clearable="false" @change="handleAppChange">
    <bk-group
      v-for="group in servingList"
      collapsible
      :key="group.space_id"
      :label="group.space_name">
      <bk-option
        v-for="item in group.children"
        :key="item.id"
        :value="item.id"
        :label="item.spec.name">
      </bk-option>
    </bk-group>
    <div class="selector-extensition" slot="extension">
      <div class="content" @click="router.push({ name: 'serving-mine' })">
        <i class="bk-bscp-icon icon-app-store app-icon"></i>
        服务管理
      </div>
    </div>
  </bk-select>
</div>
</template>
<style lang="scss" scoped>
  .selector-extensition {
    .content {
      height: 40px;
      line-height: 40px;
      text-align: center;
      border-top: 1px solid #dcdee5;
      background: #fafbfd;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
    .app-icon {
      font-size: 14px;
    }
  }
</style>
