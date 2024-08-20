<template>
  <bk-form ref="formRef" form-type="vertical" :model="{ pkgs: selectedPkgs }">
    <bk-form-item :label="t('上传至模板套餐')" property="pkgs" required>
      <bk-select multiple :model-value="selectedPkgs" @change="handleSelectPkg">
        <bk-option v-for="pkg in allOptions" v-show="pkg.id !== 0" :key="pkg.id" :value="pkg.id" :label="pkg.name">
        </bk-option>
        <template #extension>
          <div :class="['no-specified-option', { selected: unSpecifiedSelected }]" @click="handleSelectUnSpecifiedPkg">
            {{ t('未指定套餐') }}
            <Done v-if="unSpecifiedSelected" class="selected-icon" />
          </div>
        </template>
      </bk-select>
    </bk-form-item>
  </bk-form>
  <div v-if="citedList.length">
    <p class="tips">{{ tips }}</p>
    <bk-alert
      v-if="isExceedMaxFileCount"
      style="margin-bottom: 8px"
      theme="error"
      :title="
        $t('上传后，部分套餐/服务的配置文件数量将超过最大限制 ({n} 个文件)', {
          n: spaceFeatureFlags.RESOURCE_LIMIT.TmplSetTmplCnt,
        })
      " />
    <bk-loading style="min-height: 100px" :loading="loading">
      <bk-table
        v-if="!selectedPkgs.includes(0)"
        class="cited-app-table"
        :row-class="getRowCls"
        :data="citedList"
        :max-height="maxTableHeight">
        <bk-table-column :label="t('模板套餐')">
          <template #default="{ row }">
            <div v-if="row.template_set_exceeds_limit" class="app-info">
              <span class="exceeds-limit">{{ row.template_set_name }}</span>
              <InfoLine
                class="warn-icon"
                v-bk-tooltips="{
                  content: $t('上传后，该套餐的配置文件数量将达到 {n} 个，超过了最大限制', {
                    n: row.template_set_exceeds_quantity,
                  }),
                }" />
            </div>
            <span v-else>{{ row.template_set_name }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="t('使用此套餐的服务')">
          <template #default="{ row }">
            <div v-if="row.app_id">
              <div v-if="row.app_exceeds_limit" class="app-info" @click="goToConfigPage(row.app_id)">
                <div v-overflow-title class="name-text">{{ row.app_name }}</div>
                <InfoLine
                  class="warn-icon"
                  v-bk-tooltips="{
                    content: $t('上传后，该服务的配置文件数量将达到 {n} 个，超过了最大限制', {
                      n: row.app_exceeds_quantity,
                    }),
                  }" />
                <LinkToApp class="link-icon" :id="row.app_id" />
              </div>
              <div v-else-if="row.app_id" class="app-info" @click="goToConfigPage(row.app_id)">
                <div v-overflow-title class="name-text">{{ row.app_name }}</div>
                <LinkToApp class="link-icon" :id="row.app_id" />
              </div>
            </div>
            <span v-else>--</span>
          </template>
        </bk-table-column>
      </bk-table>
    </bk-loading>
  </div>
</template>
<script lang="ts" setup>
  import { computed, ref, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { Done, InfoLine } from 'bkui-vue/lib/icon';
  import useGlobalStore from '../../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../../store/template';
  import { IPackagesCitedByApps } from '../../../../../../../../../types/template';
  import { getCheckTemplateSetReferencesApps } from '../../../../../../../../api/template';
  import LinkToApp from '../../../../components/link-to-app.vue';

  const { spaceId, spaceFeatureFlags } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace, currentPkg, packageList } = storeToRefs(useTemplateStore());
  const { t } = useI18n();

  const props = defineProps<{
    configIdList: {
      name: string;
      id: number;
    }[];
  }>();

  const emits = defineEmits(['toggleBtnDisabled']);

  const router = useRouter();

  const selectedPkgs = ref<number[]>([]);
  const formRef = ref();
  const loading = ref(false);
  const citedList = ref<IPackagesCitedByApps[]>([]);

  const tips = computed(() => {
    return selectedPkgs.value.includes(0)
      ? t('若未指定套餐，此配置文件模板将无法被服务引用。后续请使用「添加至」或「添加已有配置文件」功能添加至指定套餐')
      : t('以下服务配置的未命名版本引用目标套餐的内容也将更新');
  });

  const maxTableHeight = computed(() => {
    const windowHeight = window.innerHeight;
    return windowHeight * 0.6 - 150;
  });

  // 未指定套餐选项是否选中
  const unSpecifiedSelected = computed(() => selectedPkgs.value.includes(0));

  // 套餐或服务是否有超出限制
  const isExceedMaxFileCount = computed(
    () =>
      citedList.value.some((item) => item.app_exceeds_limit || item.template_set_exceeds_limit) &&
      !unSpecifiedSelected.value,
  );

  onMounted(() => {
    selectedPkgs.value = typeof currentPkg.value === 'number' ? [currentPkg.value] : [];
    if (selectedPkgs.value.length > 0) {
      getCitedData();
    }
  });

  const allOptions = computed(() => {
    const pkgs = packageList.value.map((item) => {
      const { id, spec } = item;
      return { id, name: spec.name };
    });
    pkgs.push({ id: 0, name: t('未指定套餐') });

    return pkgs;
  });

  const getCitedData = async () => {
    loading.value = true;
    const res = await getCheckTemplateSetReferencesApps(
      spaceId.value,
      currentTemplateSpace.value,
      selectedPkgs.value,
      props.configIdList,
    );
    citedList.value = res.items;
    emits('toggleBtnDisabled', isExceedMaxFileCount.value || selectedPkgs.value.length === 0);
    loading.value = false;
  };

  const handleSelectPkg = (val: number[]) => {
    if (val.length === 0) {
      selectedPkgs.value = [];
      citedList.value = [];
      emits('toggleBtnDisabled', true);
      return;
    }

    if (unSpecifiedSelected.value) {
      selectedPkgs.value = val.filter((id) => id !== 0);
    } else {
      selectedPkgs.value = val.slice();
    }

    getCitedData();
  };

  const handleSelectUnSpecifiedPkg = () => {
    if (!unSpecifiedSelected.value) {
      selectedPkgs.value = [0];
    } else {
      selectedPkgs.value = [];
    }
    emits('toggleBtnDisabled', false);
  };

  const goToConfigPage = (id: number) => {
    const { href } = router.resolve({
      name: 'service-config',
      params: { appId: id },
    });
    window.open(href, '_blank');
  };

  const getRowCls = (data: IPackagesCitedByApps) => {
    if (data.app_exceeds_limit || data.template_set_exceeds_limit) {
      return 'error-row';
    }
    return '';
  };

  defineExpose({
    selectedPkgs,
  });
</script>
<style lang="scss" scoped>
  .header-wrapper {
    display: flex;
    align-items: center;
    .title {
      margin-right: 16px;
      padding-right: 16px;
      line-height: 24px;
      border-right: 1px solid #dcdee5;
    }
    .config-name {
      flex: 1;
      line-height: 24px;
      color: #979ba5;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
  }
  .angle-icon {
    position: absolute;
    top: 0;
    right: 4px;
    height: 100%;
    font-size: 20px;
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .no-specified-option {
    display: flex;
    align-items: center;
    position: relative;
    padding: 0 32px 0 12px;
    width: 100%;
    height: 100%;
    color: #63656e;
    cursor: pointer;
    &.selected {
      color: #3a84ff;
    }
    .selected-icon {
      position: absolute;
      top: 8px;
      right: 10px;
      font-size: 22px;
    }
  }
  .tips {
    margin-bottom: 8px;
    font-size: 12px;
    color: #63656e;
  }
  .app-info {
    display: flex;
    align-items: center;
    overflow: hidden;
    cursor: pointer;
    .name-text {
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
    .link-icon {
      flex-shrink: 0;
      margin-left: 10px;
    }
    .warn-icon {
      margin-left: 10px;
      color: #ea3636;
      font-size: 14px;
    }
  }

  .cited-app-table {
    :deep(.bk-table-body) {
      tr.error-row td {
        background: #ffeeee !important;
      }
    }
  }
</style>
