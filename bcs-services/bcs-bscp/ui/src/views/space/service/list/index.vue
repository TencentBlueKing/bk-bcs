<script setup lang="ts">
  import { ref, watch } from "vue";
  import { useI18n } from "vue-i18n";
  import { useRouter, useRoute } from "vue-router";
  import { storeToRefs } from 'pinia';
  import { useGlobalStore } from '../../../../store/global';

  import LayoutTopBar from "../../../../components/layout-top-bar.vue";
  import ServiceListContent from "./components/service-list-content.vue";
  
  const { t } = useI18n();
  const router = useRouter();
  const route = useRoute();
  const { spaceId } = storeToRefs(useGlobalStore())

  const activeTabName = ref<string>(route.name as string);
  const panels = [
    { name: "service-mine", label: t("我的服务") },
    { name: "service-all", label: t("全部服务") },
  ]

  watch(() => route.name, (val) => {
    activeTabName.value = <string>val
  })

  const handleTabChange = (name: string) => {
    activeTabName.value = name
    router.push({ name })
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
    <ServiceListContent :type="activeTabName" :space-id="spaceId" />
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