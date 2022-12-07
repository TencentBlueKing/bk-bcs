<script setup lang="ts">
import { reactive, ref, watch } from "vue";
import LayoutTopBar from "../../components/layout-top-bar.vue";
import { useRouter, useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
const { t } = useI18n();

const router = useRouter();
const route = useRoute();
const activeTabName = ref("serving-mine");
const panels = reactive([
  { name: "serving-mine", label: t("我的服务"), count: 10 },
  { name: "serving-all", label: t("全部服务"), count: 20 },
]);

watch(() => activeTabName.value,(value: string) => {
    router.push({ name: value });
  },
{ immediate: true });

watch(() => route.name, (name: any) => {
    activeTabName.value = name;
  },
{ immediate: true });
</script>

<template>
  <LayoutTopBar>
    <template #head>
      <div class="head-body">
        <bk-tab
          ext-cls="head-tabs"
          v-model:active="activeTabName"
          type="unborder-card">
          <bk-tab-panel
            v-for="item in panels"
            :key="item.name"
            :name="item.name"
            :label="item.label"
          >
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