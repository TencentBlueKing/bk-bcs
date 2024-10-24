<template>
  <bk-select
    v-model="configName"
    ref="selectorRef"
    :class="['config-selector', { 'select-error': isError }]"
    :popover-options="{ theme: 'light bk-select-popover' }"
    :popover-min-width="360"
    :filterable="true"
    :input-search="false"
    :clearable="false"
    :loading="loading"
    :search-placeholder="basicInfo?.serviceType.value === 'file' ? $t('配置文件名') : $t('配置项名称')"
    :no-data-text="$t('暂无可用配置')"
    :no-match-text="$t('搜索结果为空')"
    @change="handleConfigChange">
    <template #trigger>
      <div class="selector-trigger">
        <bk-overflow-title v-if="configName" class="config-name" type="tips">
          {{ configName }}
        </bk-overflow-title>
        <span v-else class="empty">{{ $t('请选择') }}</span>
        <AngleUpFill class="arrow-icon arrow-fill" />
      </div>
    </template>
    <bk-option v-for="item in configList" :key="item.id" :value="item.config" :label="item.config">
      <div class="config-option-item">
        <div class="name-text">{{ item.config }}</div>
      </div>
    </bk-option>
  </bk-select>
</template>

<script lang="ts" setup>
  import { ref, onMounted, inject, Ref } from 'vue';
  import { useRoute } from 'vue-router';
  import { getConfigList, getKvList } from '../../../../../api/config';
  import { AngleUpFill } from 'bkui-vue/lib/icon';
  import { IConfigItem, IConfigKvType } from '../../../../../../types/config';

  const emits = defineEmits(['select-config']);

  const route = useRoute();

  const basicInfo = inject<{ serviceName: Ref<string>; serviceType: Ref<string> }>('basicInfo');

  const loading = ref(false);
  const isError = ref(false);
  const configName = ref('');
  const bizId = ref(String(route.params.spaceId));
  const appId = ref(route.params.appId);
  const configList = ref<{ id: number; config: string }[]>([]);

  onMounted(async () => {
    await loadConfigList();
  });

  // 载入服务列表
  const loadConfigList = async () => {
    loading.value = true;
    try {
      const query = {
        start: 0,
        all: true,
      };
      let res;
      if (basicInfo?.serviceType.value === 'file') {
        // 文件型配置项
        res = await getConfigList(bizId.value, Number(appId.value), query);
      } else {
        // 键值型配置项
        res = await getKvList(bizId.value, Number(appId.value), query);
      }
      configList.value = res.details.map((item: IConfigItem | IConfigKvType) => {
        if ('path' in item.spec) {
          const { path, name } = item.spec;
          return {
            id: item.id,
            config: path.endsWith('/') ? `${path}${name}` : `${path}/${name}`,
          };
        }
        return {
          id: item.id,
          config: item.spec.key,
        };
      });
    } catch (e) {
      console.error(e);
    } finally {
      loading.value = false;
    }
  };

  // 下拉列表操作
  const handleConfigChange = async (name: string) => {
    configName.value = name;
    emits('select-config', name);
    validateConfig();
  };

  // 表单校验失败检查配置项是否为空
  const validateConfig = () => {
    isError.value = !configName.value;
    return !isError.value;
  };

  defineExpose({
    validateConfig,
  });
</script>

<style scoped lang="scss">
  .config-selector {
    &.select-error .selector-trigger {
      border-color: #ea3636;
    }
    &.popover-show .selector-trigger {
      border-color: #3a84ff;
      box-shadow: 0 0 3px #a3c5fd;
      .arrow-icon {
        transform: rotate(-180deg);
      }
    }
    &.is-focus {
      .selector-trigger {
        outline: 0;
      }
    }
    .selector-trigger {
      padding: 0 10px 0;
      height: 32px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-radius: 2px;
      transition: all 0.3s;
      background: #ffffff;
      font-size: 14px;
      border: 1px solid #c4c6cc;
      .config-name {
        max-width: 220px;
        color: #313238;
      }
      .empty {
        font-size: 12px;
        color: #c4c6cc;
      }
      .arrow-icon {
        margin-left: 13.5px;
        color: #979ba5;
        transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      }
    }
  }
</style>
