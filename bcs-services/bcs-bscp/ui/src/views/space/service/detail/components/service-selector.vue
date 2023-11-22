<template>
  <div>
    <bk-select
      v-model="localVal"
      ref="selectorRef"
      class="service-selector"
      :popover-options="{ theme: 'light bk-select-popover service-selector-popover' }"
      :popover-min-width="360"
      :filterable="true"
      :input-search="false"
      :clearable="false"
      :loading="loading"
      @change="handleAppChange">
      <template #trigger>
        <div class="selector-trigger">
          <input readonly :value="appData.spec.name" />
          <AngleDown class="arrow-icon" />
        </div>
      </template>
      <bk-option
        v-for="item in serviceList"
        :key="item.id"
        :value="item.id"
        :label="item.spec.name">
        <div
            v-cursor="{
              active: !item.permissions.view
            }"
            :class="['service-option-item', { 'no-perm': !item.permissions.view }]"
            @click="handleOptionClick(item, $event)">
            <div class="name-text">{{ item.spec.name }}</div>
            <div class="type-tag">@todo</div>
          </div>
      </bk-option>
      <template #extension>
        <div class="selector-extensition">
          <div class="content" @click="router.push({ name: 'service-all' })">
            <i class="bk-bscp-icon icon-app-store app-icon"></i>
            服务管理
          </div>
        </div>
      </template>
    </bk-select>
  </div>
</template>
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { AngleDown } from 'bkui-vue/lib/icon';
import useGlobalStore from '../../../../../store/global';
import useServiceStore from '../../../../../store/service';
import { IAppItem } from '../../../../../../types/app';
import { getAppList } from '../../../../../api';

const route = useRoute();
const router = useRouter();

const { appData } = storeToRefs(useServiceStore());
const { showApplyPermDialog, permissionQuery } = storeToRefs(useGlobalStore());

const bizId = route.params.spaceId as string;

const props = defineProps<{
  value: number;
}>();

defineEmits(['change']);

const serviceList = ref<IAppItem[]>([]);
const loading = ref(false);
const localVal = ref(props.value);
const selectorRef = ref();

watch(
  () => props.value,
  (val) => {
    localVal.value = val;
  },
);

onMounted(() => {
  loadServiceList();
});

const loadServiceList = async () => {
  loading.value = true;
  try {
    const query = {
      start: 0,
      all: true,
    };
    const resp = await getAppList(bizId, query);
    serviceList.value = resp.details;
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
};

// 点击无查看权限的选项，弹出申请权限弹窗
const handleOptionClick = (service: IAppItem, event: Event) => {
  if (!service.permissions.view) {
    selectorRef.value.hidePopover();
    event.stopPropagation();
    permissionQuery.value = {
      resources: [
        {
          biz_id: service.biz_id,
          basic: {
            type: 'app',
            action: 'view',
            resource_id: service.id,
          },
        },
      ],
    };

    showApplyPermDialog.value = true;
  }
};

const handleAppChange = (id: number) => {
  const service = serviceList.value.find(service => service.id === id);
  if (service) {
    router.push({ name: route.name as string, params: { spaceId: service.space_id, appId: id } });
  }
};
</script>
<style lang="scss" scoped>
.service-selector {
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
  transition: all 0.3s;
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
    transition: all 0.3s;
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
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    font-size: 20px;
    color: #979ba5;
  }
}

.service-option-item {
  position: relative;
  flex: 1;
  padding: 0 80px 0 12px;
  &.no-perm {
    background-color: #fafafa !important;
    color: #cccccc !important;
  }
  .name-text {
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
  }
  .type-tag {
    position: absolute;
    top: 5px;
    right: 16px;
    width: 52px;
    height: 22px;
    line-height: 22px;
    color: #63656e;
    font-size: 12px;
    text-align: center;
    background: #f0f1f5;
    border-radius: 2px;
  }
}
.selector-extensition {
  flex: 1;
  .content {
    height: 40px;
    line-height: 40px;
    text-align: center;
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
<style lang="scss">
  .service-selector-popover {
    .bk-select-option {
      padding: 0 !important;
    }
  }
</style>
