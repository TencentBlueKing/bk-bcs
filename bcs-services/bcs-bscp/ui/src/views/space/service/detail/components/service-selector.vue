<script setup lang="ts">
  import { ref, watch ,onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { AngleDown } from 'bkui-vue/lib/icon'
  import { useUserStore } from '../../../../../store/user'
  import { useServiceStore } from '../../../../../store/service'
  import { IAppItem } from '../../../../../../types/app'
  import { getAppList } from "../../../../../api";

  const route = useRoute()
  const router = useRouter()
  
  const { appData } = storeToRefs(useServiceStore())
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
    class="app-selector"
    :filterable="true"
    :input-search="false"
    :clearable="false"
    :loading="loading"
    @change="handleAppChange">
    <template #trigger>
      <div class="selector-trigger">
        <input readonly :value="appData.spec.name">
        <AngleDown class="arrow-icon" />
      </div>
    </template>
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
  .app-selector {
    &.popover-show {
      .selector-trigger .arrow-icon {
        transform: rotate(-180deg);
      }
    }
    &.is-focus {
      .selector-trigger {
        border-color: #3a84ff;
        box-shadow: 0 0 3px #a3c5fd;
        outline: 0;
      }
    }
  }
  .selector-trigger {
    display: inline-flex;
    align-items: stretch;
    width: 100%;
    height: 32px;
    font-size: 12px;
    border: 1px solid #c4c6cc;
    border-radius: 2px;
    transition: all .3s;
    & > input {
      flex: 1;
      width: 100%;
      padding: 0 24px 0 10px;
      line-height: 1;
      color: #63656e;
      background-color: #fff;
      border-radius: 2px;
      border: none;
      outline: none;
      transition: all .3s;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      cursor: pointer;
    }
    .arrow-icon {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      position: absolute;
      right: 4px;
      top: 0;
      width: 20px;
      height: 100%;
      transition: transform .3s cubic-bezier(.4,0,.2,1);
      font-size: 20px;
      color: #979ba5;
    }
  }
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
