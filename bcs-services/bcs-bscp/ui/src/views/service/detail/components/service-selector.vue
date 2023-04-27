<script setup lang="ts">
  import { ref, watch ,onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { useUserStore } from '../../../../store/user'
  import { IAppItem } from '../../../../../types/app'
  import { getAppList } from "../../../../api";

  interface IServiceGroupItem {
    space_id: number;
    space_name: string;
    space_type_id: string;
    space_type_name: string;
    children: Array<IAppItem>;
  }

  const route = useRoute()
  const router = useRouter()

  const { userInfo } = storeToRefs(useUserStore())

  const bizId = <string>route.params.spaceId

  const props = defineProps<{
    value: number
  }>()

  defineEmits(['change'])

  const serviceList = ref<IAppItem[]>([])
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
      const query = {
        start: 0,
        limit: 100,
        operator: userInfo.value.username
      }
      const resp = await getAppList(bizId, query);
      serviceList.value = resp.details
    } catch (e) {
      console.error(e);
    } finally {
        loading.value = false;
    }
  };

  const handleAppChange = (id: number) => {
    const service = serviceList.value.find(service => service.id === id)
    if (service) {
      router.push({ name: <string>route.name, params: { spaceId: service.space_id, appId: id } })
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
    <bk-option
      v-for="item in serviceList"
      :key="item.id"
      :value="item.id"
      :label="item.spec.name">
    </bk-option>
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
