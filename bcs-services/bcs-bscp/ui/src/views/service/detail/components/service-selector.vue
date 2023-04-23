<script setup lang="ts">
  import { ref, Ref, watch ,onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { IAppItem } from '../../../../../types/app'
  import { getAllApp } from "../../../../api";

  interface IServiceGroupItem {
    space_id: number;
    space_name: string;
    space_type_id: string;
    space_type_name: string;
    children: Array<IAppItem>;
  }

  const route = useRoute()
  const router = useRouter()

  const props = defineProps<{
    value: number
  }>()

  defineEmits(['change'])

  const serviceList = ref([]) as Ref<IServiceGroupItem[]>
  const loading = ref(false)
  const localVal = ref(props.value)

  watch(() => props.value, (val) => {
    localVal.value = val
  })

  onMounted(() => {
    loadServiceList()
  })

  const loadServiceList = async() => {
    loading.value = true;
    try {
      const resp = await getAllApp();
      const groupedList: Array<IServiceGroupItem> = []
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
      serviceList.value = groupedList
    } catch (e) {
      console.error(e);
    } finally {
        loading.value = false;
    }
  };

  const handleAppChange = (id: number) => {
    const group = serviceList.value.find(group => {
      return group.children.find(item => item.id === id)
    })
    if (group) {
      router.push({ name: <string>route.name, params: { spaceId: group.space_id, appId: id } })
    }
  }
</script>
<template>
<div>
  <bk-select
    v-model="localVal"
    :filterable="true"
    :clearable="false"
    :loading="loading"
    @change="handleAppChange">
    <bk-option-group
      v-for="group in serviceList"
      collapsible
      :key="group.space_id"
      :label="group.space_name">
      <bk-option
        v-for="item in group.children"
        :key="item.id"
        :value="item.id"
        :label="item.spec.name">
      </bk-option>
    </bk-option-group>
    <div class="selector-extensition" slot="extension">
      <div class="content" @click="router.push({ name: 'service-mine' })">
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
