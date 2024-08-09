<template>
  <bk-dialog
    ext-cls="add-configs-to-pkg-dialog"
    :confirm-text="t('添加')"
    :cancel-text="t('取消')"
    :width="640"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close">
    <template #header>
      <div class="header-wrapper">
        <div class="title">{{ isMultiple || isAcrossChecked ? t('批量添加至') : t('添加至套餐') }}</div>
        <div v-if="props.value.length === 1" class="config-name">{{ fileAP(props.value[0]) }}</div>
      </div>
    </template>
    <div v-if="isMultiple || isAcrossChecked" class="selected-mark">
      {{ t('已选') }}
      <span class="num">{{ isAcrossChecked ? dataCount - props.value.length : props.value.length }}</span>
      {{ t('个配置文件') }}
    </div>
    <bk-form ref="formRef" form-type="vertical" :model="{ pkgs: selectedPkgs }">
      <bk-form-item
        :label="isMultiple || isAcrossChecked ? t('添加至模板套餐') : t('模板套餐')"
        property="pkgs"
        required>
        <bk-select v-model="selectedPkgs" multiple @change="handPkgsChange" @clear="handleClearPkgs">
          <bk-option
            v-for="pkg in allPackages"
            :key="pkg.id"
            :value="pkg.id"
            :label="pkg.spec.name"
            :disabled="citeByPkgIds && citeByPkgIds.includes(pkg.id)">
          </bk-option>
        </bk-select>
      </bk-form-item>
    </bk-form>
    <div v-if="citedList.length">
      <p class="tips">
        {{ t('以下服务配置的未命名版本中将添加已选配置文件的') }} <span class="notice">latest {{ t('版本') }}</span>
      </p>
      <div class="service-table">
        <bk-loading style="min-height: 100px" :loading="loading">
          <bk-table :data="citedList" :max-height="maxTableHeight">
            <bk-table-column :label="t('目标模板套餐')" prop="template_set_name"></bk-table-column>
            <bk-table-column :label="t('使用此套餐的服务')">
              <template #default="{ row }">
                <div v-if="row.app_id" class="app-info" @click="goToConfigPageImport(row.app_id)">
                  <div v-overflow-title class="name-text">{{ row.app_name }}</div>
                  <LinkToApp class="link-icon" :id="row.app_id" />
                </div>
              </template>
            </bk-table-column>
          </bk-table>
        </bk-loading>
      </div>
    </div>
  </bk-dialog>
</template>
<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import { ITemplateConfigItem, IPackagesCitedByApps } from '../../../../../../../../types/template';
  import useGlobalStore from '../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../store/template';
  import { addTemplateToPackage, getUnNamedVersionAppsBoundByPackages } from '../../../../../../../api/template';
  import LinkToApp from '../../../components/link-to-app.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { packageList, currentTemplateSpace, currentPkg, isAcrossChecked, dataCount } = storeToRefs(useTemplateStore());
  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
    value: ITemplateConfigItem[];
    citeByPkgIds?: number[];
  }>();

  const emits = defineEmits(['update:show', 'added']);

  const router = useRouter();

  const formRef = ref();
  const selectedPkgs = ref<number[]>([]);
  const loading = ref(false);
  const citedList = ref<IPackagesCitedByApps[]>([]);
  const pending = ref(false);

  const allPackages = computed(() => packageList.value.filter((pkg) => pkg.id !== currentPkg.value));

  const isMultiple = computed(() => props.value.length > 1);

  const maxTableHeight = computed(() => {
    const windowHeight = window.innerHeight;
    return windowHeight * 0.6 - 200;
  });

  watch(
    () => props.show,
    (val) => {
      if (val) {
        citedList.value = [];
        pending.value = false;
        if (props.citeByPkgIds && props.citeByPkgIds.length > 0) {
          selectedPkgs.value = allPackages.value
            .filter((pkg) => props.citeByPkgIds!.includes(pkg.id))
            .map((pkg) => pkg.id);
          if (selectedPkgs.value.length > 0) {
            getCitedData();
          }
        } else {
          selectedPkgs.value = [];
        }
      }
    },
  );

  // 配置文件绝对路径
  const fileAP = computed(() => (config: ITemplateConfigItem) => {
    const { path, name } = config.spec;
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  });

  const goToConfigPageImport = (id: number) => {
    const { href } = router.resolve({
      name: 'service-config',
      params: { appId: id },
      query: { pkg_id: currentTemplateSpace.value },
    });
    window.open(href, '_blank');
  };

  const getCitedData = async () => {
    console.log('获取表格');
    loading.value = true;
    const params = {
      start: 0,
      all: true,
    };
    const res = await getUnNamedVersionAppsBoundByPackages(
      spaceId.value,
      currentTemplateSpace.value,
      selectedPkgs.value,
      params,
    );
    citedList.value = res.details;
    loading.value = false;
  };

  const handPkgsChange = () => {
    if (selectedPkgs.value.length > 0) {
      getCitedData();
    } else {
      citedList.value = [];
    }
  };

  const handleClearPkgs = () => {
    selectedPkgs.value = allPackages.value.filter((pkg) => props.citeByPkgIds!.includes(pkg.id)).map((pkg) => pkg.id);
    getCitedData();
  };

  const handleConfirm = async () => {
    const isValid = await formRef.value.validate();
    if (!isValid) return;
    try {
      pending.value = true;
      const templateIds = props.value.map((item) => item.id);
      await addTemplateToPackage(
        spaceId.value,
        currentTemplateSpace.value,
        templateIds,
        selectedPkgs.value,
        isAcrossChecked.value,
        typeof currentPkg.value === 'string' ? 0 : currentPkg.value,
        currentPkg.value === 'no_specified',
      );
      emits('added');
      close();
      Message({
        theme: 'success',
        message: t('添加配置文件成功'),
      });
    } catch (e) {
      console.log(e);
    } finally {
      pending.value = false;
    }
  };

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .header-wrapper {
    display: flex;
    align-items: center;
    .title {
      line-height: 24px;
    }
    .config-name {
      flex: 1;
      margin-left: 16px;
      padding-left: 16px;
      line-height: 24px;
      color: #979ba5;
      border-left: 1px solid #dcdee5;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
  }
  .selected-mark {
    display: inline-block;
    margin-bottom: 16px;
    padding: 0 12px;
    height: 32px;
    line-height: 32px;
    border-radius: 16px;
    font-size: 12px;
    color: #63656e;
    background: #f0f1f5;
    .num {
      color: #3a84ff;
    }
  }
  .tips {
    margin: 0 0 16px;
    font-size: 12px;
    color: #63656e;
    .notice {
      color: #ff9c01;
    }
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
  }
</style>
