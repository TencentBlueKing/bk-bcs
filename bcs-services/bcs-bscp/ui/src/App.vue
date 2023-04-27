<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useGlobalStore } from './store/global'
import { useUserStore } from './store/user'
import isCrossOriginIFrame from './utils/is-cross-origin-iframe'
import { getSpaceList } from './api'
import { ISpaceDetail } from '../types/index'
import Header from "./components/head.vue";
import PermissionDialog from './components/permission/apply-dialog.vue'


const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const globalStore = useGlobalStore()
const { getUserInfo } = userStore
const { spaceId, spaceList, showApplyPermDialog } = storeToRefs(globalStore)

const showPage = ref(true)
const spacesLoading = ref(true)

watch(() => userStore.showLoginModal, (val) => {
  if (val) {
    const topWindow = isCrossOriginIFrame() ? window : window.top
    // @ts-ignore
    topWindow.BLUEKING.corefunc.open_login_dialog(userStore.loginUrl)
  }
})

watch(() => route.params.spaceId, (val) => {
  if (val) {
    reloadPage()
  }
})

onMounted(() => {
  loadSpacesData()
  getUserInfo()
});

// 加载全部空间列表
const loadSpacesData = async () => {
  spacesLoading.value = true;
  const res = await getSpaceList();
  spaceList.value = res.items
  if (route.params.spaceId) {
    spaceId.value = <string>route.params.spaceId;
  } else {
    const hasPermSpace = res.items.find((item: ISpaceDetail) => item.permission)
    if (hasPermSpace) {
      spaceId.value = hasPermSpace.space_id;
    } else {
      spaceId.value = res.items[0]?.space_id;
    }
  }

  if (!route.params.spaceId) {
    const { query, params } = route
    router.push({ name: <string>route.name, query, params: { ...params, spaceId: spaceId.value } })
  }

  spacesLoading.value = false;
}

const reloadPage = () => {
  showPage.value = false
  setTimeout(() => {
    showPage.value = true
  }, 500)
}

</script>

<template>
  <div class="page-content-container" v-if="!spacesLoading">
    <Header></Header>
    <div class="content">
      <router-view v-if="showPage"></router-view>
      <permission-dialog :show="showApplyPermDialog"></permission-dialog>
    </div>
  </div>
</template>

<style scoped>
.page-content-container {
  min-width: 1366px;
  overflow: auto;
}
.content {
  height: calc(100vh - 52px);
}
</style>
