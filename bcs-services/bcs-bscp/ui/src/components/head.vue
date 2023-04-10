<script setup lang="ts">
  import { computed } from 'vue'
  import { useRoute } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { useUserStore } from '../store/user'

  const route = useRoute()
  const store = useUserStore()
  const { userInfo } = storeToRefs(store)

  const navList = [
    { id: 'serving', name: '服务管理'},
    { id: 'groups-management', name: '分组管理'},
    { id: 'scripts-management', name: '脚本管理'},
    { id: 'keys-management', name: '密钥管理'}
  ]

  const activedRootRoute = computed(() => {
    if (route.name === 'serving-config') {
      return 'serving'
    }
    return route.matched[0]?.name
  })

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
          :to="{ name: nav.id }">
          {{ nav.name }}
        </router-link>
      </div>
    </div>
    <div class="head-right">
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
    justify-self: flex-end;
    font-size: 14px;
    color: #979ba5;
  }
}
</style>
