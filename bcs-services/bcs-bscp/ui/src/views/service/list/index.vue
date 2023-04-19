<script setup lang="ts">
import { reactive, ref, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { storeToRefs } from "pinia";
import LayoutTopBar from "../../../components/layout-top-bar.vue";

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const activeTabName = ref(route.name);
const panels = [
  { name: "service-mine", label: t("我的服务") },
  { name: "service-all", label: t("全部服务") },
]

const handleTabChange = (name: string) => {
  activeTabName.value = name
  router.push({ name, params: { spaceId: route.params.spaceId } })
}

</script>
<template>
  <LayoutTopBar>
    <template #head>
      <div class="head-body">
        <bk-tab
          ext-cls="head-tabs"
          type="unborder-card"
          :active="activeTabName"
          @change="handleTabChange">
          <bk-tab-panel
            v-for="item in panels"
            :key="item.name"
            :name="item.name"
            :label="item.label">
          </bk-tab-panel>
        </bk-tab>
      </div>
    </template>
    <router-view></router-view>
  </LayoutTopBar>
</template>


<style lang="scss" scoped>
.head-body {
  display: flex;
  justify-content: center;
  .head-tabs {
    width: 200px;
    font-size: 14px;

    :deep(.bk-tab-content) {
      display: none;
    }

    :deep(.bk-tab-header) {
      border-bottom: none;
    }
  }
}
</style>