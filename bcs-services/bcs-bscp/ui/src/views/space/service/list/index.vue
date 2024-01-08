<template>
  <LayoutTopBar>
    <template #head>
      <div class="head-body">
        <bk-tab ext-cls="head-tabs" type="unborder-card" :active="activeTabName" @change="handleTabChange">
          <bk-tab-panel v-for="item in panels" :key="item.name" :name="item.name" :label="item.label"> </bk-tab-panel>
        </bk-tab>
      </div>
    </template>
    <ServiceListContent
      :type="activeTabName"
      :space-id="spaceId"
      :perm-check-loading="permCheckLoading"
      :has-create-service-perm="hasCreateServicePerm"
    />
    <AppFooter />
  </LayoutTopBar>
</template>
<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter, useRoute } from 'vue-router';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../../../store/global';
import { permissionCheck } from '../../../../api/index';
import LayoutTopBar from '../../../../components/layout-top-bar.vue';
import ServiceListContent from './components/service-list-content.vue';
import AppFooter from '../../../../components/footer.vue';

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const { spaceId } = storeToRefs(useGlobalStore());

const activeTabName = ref<string>(route.name as string);
const hasCreateServicePerm = ref(false);
const permCheckLoading = ref(false);
const panels = computed(() => [
  { name: 'service-all', label: t('全部服务') },
  { name: 'service-mine', label: t('我的服务') },
]);

watch(
  () => route.name,
  (val) => {
    activeTabName.value = val as string;
  },
);

watch(
  () => spaceId.value,
  () => {
    checkCreateServicePerm();
  },
);

onMounted(() => {
  checkCreateServicePerm();
  // 访问服务管理列表页时，清空上次访问服务记录
  localStorage.removeItem('lastAccessedServiceDetail');
});

const checkCreateServicePerm = async () => {
  permCheckLoading.value = true;
  const res = await permissionCheck({
    resources: [
      {
        biz_id: spaceId.value,
        basic: {
          type: 'app',
          action: 'create',
        },
      },
    ],
  });
  hasCreateServicePerm.value = res.is_allowed;
  permCheckLoading.value = false;
};

const handleTabChange = (name: string) => {
  activeTabName.value = name;
  router.push({ name });
};
</script>

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
