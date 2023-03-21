<script setup lang="ts">
  import { storeToRefs } from 'pinia'
  import { useUserStore } from '../store/user'

  const store = useUserStore()
  const { userInfo } = storeToRefs(store)

</script>

<template>
  <div class="header">
    <div class="head-left">
      <span class="logo">
        <img src="../assets/logo.svg" alt="" />
      </span>
      <span class="head-title"> 服务配置中心 </span>
      <div class="head-routes">
        <router-link to="/serving/mine/">服务管理</router-link>
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

      a {
        color: #fff;
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