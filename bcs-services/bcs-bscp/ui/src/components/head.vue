<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../store/global'
  import { useUserStore } from '../store/user'
  import { ISpaceDetail } from '../../types/index'

  const route = useRoute()
  const router = useRouter()
  const { spaceId, spaceList, showApplyPermDialog, permissionQuery } = storeToRefs(useGlobalStore())
  const { userInfo } = storeToRefs(useUserStore())

  const navList = [
    { id: 'service', name: '服务管理'},
    { id: 'groups-management', name: '分组管理'},
    { id: 'scripts-management', name: '脚本管理'},
    { id: 'credentials-management', name: '服务密钥'}
  ]

  const activedRootRoute = computed(() => {
    if (['service-config', 'service-mine', 'service-all'].includes(<string>route.name)) {
      return 'service'
    }
    return route.matched[0]?.name
  })

  const handleSelectSpace = (id: string) => {
    const space = spaceList.value.find((item: ISpaceDetail) => item.space_id === id)
    if (space) {
      if (!space.permission) {
        permissionQuery.value = {
          biz_id: id,
          basic: {
            type: "biz",
            action: "find_business_resource",
            resource_id: id
          },
          gen_apply_url: true
        }
        
        showApplyPermDialog.value = true
        return
      }
      spaceId.value = id
      const { query, params } = route
      router.push({ name: <string>route.name, query, params: { ...params, spaceId: id } })
    }
  }

</script>

<template>
  <div class="header">
    <div class="head-left">
      <span class="logo">
        <img src="../assets/logo.svg" alt="" />
      </span>
      <span class="head-title"> 服务配置平台 </span>
      <div class="head-routes">
        <router-link
          v-for="nav in navList"
          :class="['nav-item', { actived: activedRootRoute === nav.id }]"
          :key="nav.id"
          :to="{ name: nav.id, params: { spaceId } }">
          {{ nav.name }}
        </router-link>
      </div>
    </div>
    <div class="head-right">
      <bk-select
        class="space-selector"
        id-key="space_id"
        display-key="space_name"
        :model-value="spaceId"
        :popover-options="{ theme: 'light bk-select-popover space-selector-popover' }"
        :filterable="true"
        :clearable="false"
        @change="handleSelectSpace">
        <bk-option v-for="item in spaceList" :key="item.space_id" :value="item.space_id" :label="item.space_name">
          <div v-cursor="{ active: !item.permission }" :class="['biz-option-item', { 'no-perm': !item.permission }]">
            <div class="name">{{ item.space_name }}</div>
            <span class="tag">{{ item.space_type_name }}</span>
          </div>
        </bk-option>
      </bk-select>
      <span>{{ userInfo.username }}</span>
    </div>
  </div>
</template>


<style lang="scss" scoped>
.header {
  height: 52px;
  background: #182132;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;

  .head-left {
    display: flex;
    align-items: center;
    .logo {
      display: inline-flex;
      width: 30px;
      height: 30px;
    }
    .head-routes {
      padding-left: 112px;
      font-size: 14px;
      .nav-item {
        margin-left: 32px;
        color: #96a2b9;
        font-size: 14px;
        &.actived {
          color: #ffffff;
        }
      }
    }

    .head-title {
      display: inline-flex;
      padding-left: 20px;
      font-weight: Bold;
      font-size: 18px;
      color: #96a2b9;
    }
  }

  .head-right {
    display: flex;
    align-items: center;
    justify-self: flex-end;
    font-size: 14px;
    color: #979ba5;
  }
}
.space-selector {
  margin-right: 24px;
  width: 240px;
  :deep(.bk-select-trigger) {
    .bk-input--default {
      border: none;
    }
    .bk-input--text {
      background: #303d55;
      color: #d3d9e4;
    }
  }
}
.biz-option-item {
  position: relative;
  padding: 0 12px;
  &.no-perm {
    background-color: #fafafa !important;
    color: #cccccc !important;
    .tag {
      border-color: #e6e6e6 !important;
      color: #cccccc !important;
    }
  }
  .name {
    padding-right: 30px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .tag {
    position: absolute;
    top: 10px;
    right: 8px;
    padding: 2px;
    font-size: 12px;
    line-height: 1;
    color: #3a84ff;
    border: 1px solid #3a84ff;
    border-radius: 2px;
  }
}
</style>
<style lang="scss">
  .space-selector-popover .bk-select-option {
    padding: 0 !important;
  }
</style>
