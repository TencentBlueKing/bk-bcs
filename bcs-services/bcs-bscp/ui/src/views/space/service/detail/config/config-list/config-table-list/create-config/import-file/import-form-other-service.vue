<template>
  <div class="other-service-wrap">
    <div class="select-service">
      <div class="label">{{ $t('选择服务') }}</div>
      <bk-select
        v-model="selectAppId"
        :loading="serviceListloading"
        style="width: 362px"
        filterable
        auto-focus
        :clearable="false"
        @select="handleSelectApp">
        <bk-option v-for="item in serviceList" :id="item.id" :key="item.id" :name="item.spec.name" />
      </bk-select>
    </div>
    <div class="select-version">
      <div class="label">{{ $t('选择版本') }}</div>
      <bk-select
        v-model="selectVerisonId"
        :loading="versionListLoading"
        style="width: 362px"
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
  import { ref, onMounted } from 'vue';
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

  onMounted(() => {
    loadServiceList();
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
