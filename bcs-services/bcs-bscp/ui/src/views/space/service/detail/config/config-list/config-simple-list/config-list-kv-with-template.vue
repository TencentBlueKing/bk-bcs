<template>
  <section class="config-list-with-templates">
    <SearchInput
      v-model="searchStr"
      class="config-search-input"
      placeholder="配置项名/创建人/修改人"
      @search="getListData"
    />
    <bk-loading class="loading-wrapper" :loading="loading">
      <div v-for="config in configList" :class="['config-item']" :key="config.id" @click="handleConfigClick(config)">
        <div class="config-name">{{ config.spec.key }}</div>
        <div class="config-type">{{ config.spec.kv_type }}</div>
      </div>
      <bk-exception v-if="configList.length === 0" scene="part" type="empty" description="暂无配置文件"></bk-exception>
    </bk-loading>
    <EditConfig
      v-model:show="isShowEdit"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :config="(selectedConfig as IConfigKvItem)"
      :editable="isEditConfig"
      :view="!isEditConfig"
      @confirm="getListData"
    />
  </section>
</template>
<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue';
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../../../store/config';
import { IConfigKvType, IConfigKvItem } from '../../../../../../../../types/config';
import { ICommonQuery } from '../../../../../../../../types/index';
import { getKvList, getReleaseKvList } from '../../../../../../../api/config';
import SearchInput from '../../../../../../../components/search-input.vue';
import EditConfig from '../config-table-list/edit-config-kv.vue';

const store = useConfigStore();
const { versionData } = storeToRefs(store);

const props = defineProps<{
  bkBizId: string;
  appId: number;
}>();

const loading = ref(false);
const configList = ref<IConfigKvType[]>([]);
const searchStr = ref('');
const selectedConfig = ref<IConfigKvItem>();
const isShowEdit = ref(false);
const isEditConfig = ref(true);

watch(
  () => versionData.value.id,
  () => {
    getListData();
  },
);

// 是否为未命名版本
const isUnNamedVersion = computed(() => versionData.value.id === 0);

onMounted(() => {
  getListData();
});

// 获取配置文件列表
const getListData = async () => {
  // 拉取到版本列表之前不加在列表数据
  if (typeof versionData.value.id !== 'number') {
    return;
  }
  loading.value = true;
  try {
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };
    if (searchStr.value) {
      params.search_fields = 'key,kv_type,creator';
      params.search_key = searchStr.value;
    }
    let res;
    if (isUnNamedVersion.value) {
      res = await getKvList(props.bkBizId, props.appId, params);
    } else {
      res = await getReleaseKvList(props.bkBizId, props.appId, versionData.value.id, params);
    }
    configList.value = res.details;
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
};

const handleConfigClick = (config: IConfigKvType) => {
  selectedConfig.value = config.spec;
  isEditConfig.value = isUnNamedVersion.value;
  isShowEdit.value = true;
};
</script>
<style lang="scss" scoped>
.config-list-with-templates {
  padding: 24px;
  height: 100%;
  background: #fafbfd;
  overflow: auto;
}
.config-search-input {
  margin-bottom: 16px;
}
.loading-wrapper {
  height: calc(100% - 48px);
  overflow: auto;
}
.group-title {
  display: flex;
  align-items: center;
  margin: 8px 0;
  line-height: 20px;
  font-size: 12px;
  color: #63656e;
  cursor: pointer;
  .fold-icon {
    margin-right: 8px;
    font-size: 14px;
    color: #3a84ff;
    transition: transform 0.2s ease-in-out;
    &.fold {
      color: #c4c6cc;
      transform: rotate(-90deg);
    }
  }
}
.config-list-wrapper {
  max-height: 472px; // 每个分组最多显示10条，超出后滚动显示
  overflow: auto;
}
.config-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
  background: #ffffff;
  box-shadow: 0 1px 1px 0 rgba(0, 0, 0, 0.06);
  border-radius: 2px;
  cursor: pointer;
  &.disabled {
    cursor: not-allowed;
    .config-type,
    .config-name {
      color: #dcdee5;
    }
  }
  &:not(.disabled):hover {
    background: #e1ecff;
  }
  .config-name {
    padding: 0 16px;
    width: 242px;
    height: 40px;
    line-height: 40px;
    color: #313238;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
  }
  .config-type {
    color: #979ba5;
  }
}
</style>
