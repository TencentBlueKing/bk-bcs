<template>
  <div class="other-service-wrap">
    <div class="select-service">
      <div class="label">{{ $t('选择服务') }}</div>
      <bk-select
        v-model="selectAppId"
        :loading="serviceListloading"
        style="width: 342px"
        filterable
        auto-focus
        :clearable="false"
        @select="handleSelectApp">
        <bk-option
          v-for="item in serviceList"
          :id="item.id"
          :key="item.id"
          :name="item.spec.name"
          :disabled="appDiabled(item)">
          <span
            v-bk-tooltips="{
              content: $t('当前服务仅允许导入数据类型为 {n} 的服务配置项', { n: kvAppType }),
              disabled: !appDiabled(item),
              extCls: 'disabled-service-tips',
            }">
            {{ item.spec.name }}
          </span>
        </bk-option>
      </bk-select>
    </div>
    <div class="select-version">
      <div class="label">{{ $t('选择版本') }}</div>
      <bk-select
        v-model="selectVerisonId"
        :loading="versionListLoading"
        style="width: 342px"
        filterable
        auto-focus
        :clearable="false"
        @select="emits('selectVersion', selectAppId, selectVerisonId)">
        <bk-option v-for="item in versionList" :id="item.id" :key="item.id" :name="item.spec.name" />
      </bk-select>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed } from 'vue';
  import { storeToRefs } from 'pinia';
  import { getAppList } from '../../../../../../../../../api';
  import { getConfigVersionList } from '../../../../../../../../../api/config';
  import { IAppItem } from '../../../../../../../../../../types/app';
  import { IConfigVersion } from '../../../../../../../../../../types/config';
  import useServiceStore from '../../../../../../../../../store/service';

  const serviceStore = useServiceStore();
  const { isFileType } = storeToRefs(serviceStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();
  const emits = defineEmits(['selectVersion', 'clear']);

  const serviceListloading = ref(false);
  const selectAppId = ref();
  const selectVerisonId = ref();
  const serviceList = ref<IAppItem[]>([]);
  const versionList = ref<IConfigVersion[]>([]);
  const versionListLoading = ref(false);
  const kvAppType = ref('any');

  onMounted(() => {
    loadServiceList();
  });

  const appDiabled = computed(() => (app: IAppItem) => {
    return !isFileType.value && kvAppType.value !== 'any' && app.spec.data_type !== kvAppType.value;
  });

  const loadServiceList = async () => {
    serviceListloading.value = true;
    try {
      const query = {
        start: 0,
        all: true,
      };
      const resp = await getAppList(props.bkBizId, query);
      serviceList.value = resp.details.filter((app: IAppItem) => {
        if (isFileType.value) {
          return app.spec.config_type === 'file' && app.id !== props.appId;
        }
        return app.spec.config_type === 'kv' && app.id !== props.appId;
      });
      if (!isFileType.value) {
        kvAppType.value = resp.details.find((app: IAppItem) => app.id === props.appId)?.spec.data_type || 'any';
        if (kvAppType.value !== 'any') {
          serviceList.value.sort((a, b) => {
            // 判断 a 和 b 是否匹配 kvAppType.value
            const aMatches = a.spec.data_type === kvAppType.value;
            const bMatches = b.spec.data_type === kvAppType.value;

            // 如果 a 匹配，b 不匹配，a 应该排在前面
            if (aMatches && !bMatches) return -1;
            // 如果 b 匹配，a 不匹配，b 应该排在前面
            if (bMatches && !aMatches) return 1;

            // 如果 a 和 b 都匹配或都不匹配，保持原有顺序
            return 0;
          });
        }
      }
    } catch (e) {
      console.error(e);
    } finally {
      serviceListloading.value = false;
    }
  };

  const getVersionList = async () => {
    try {
      versionListLoading.value = true;
      const params = {
        start: 0,
        all: true,
      };
      const res = await getConfigVersionList(props.bkBizId, selectAppId.value, params);
      versionList.value = res.data.details;
    } catch (e) {
      console.error(e);
    } finally {
      versionListLoading.value = false;
    }
  };

  const handleSelectApp = () => {
    getVersionList();
    selectVerisonId.value = undefined;
    emits('clear');
  };
</script>

<style scoped lang="scss">
  .select-service {
    display: flex;
  }
  .select-version {
    display: flex;
  }
</style>

<style>
  .disabled-service-tips {
    z-index: 9999 !important;
  }
</style>
