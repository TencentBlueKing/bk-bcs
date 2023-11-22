<template>
  <bk-dialog
    title="导入至套餐"
    ext-cls="create-to-pkg-dialog"
    confirm-text="导入"
    :width="640"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="props.pending"
    @confirm="handleConfirm"
    @closed="close"
  >
    <bk-form ref="formRef" form-type="vertical" :model="{ pkgs: selectedPkgs }">
      <bk-form-item label="模板套餐" property="pkgs" required>
        <bk-select multiple :model-value="selectedPkgs" @change="handleSelectPkg">
          <bk-option v-for="pkg in allOptions" v-show="pkg.id !== 0" :key="pkg.id" :value="pkg.id" :label="pkg.name">
          </bk-option>
          <template #extension>
            <div
              :class="['no-specified-option', { selected: unSpecifiedSelected }]"
              @click="handleSelectUnSpecifiedPkg"
            >
              未指定套餐
              <Done v-if="unSpecifiedSelected" class="selected-icon" />
            </div>
          </template>
        </bk-select>
      </bk-form-item>
    </bk-form>
    <div v-if="citedList.length">
    <p class="tips">{{ tips }}</p>
    <bk-loading style="min-height: 100px" :loading="loading">
      <bk-table v-if="!selectedPkgs.includes(0)" :data="citedList" :max-height="maxTableHeight">
        <bk-table-column label="模板套餐" prop="template_set_name"></bk-table-column>
        <bk-table-column label="使用此套餐的服务">
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
  </bk-dialog>
</template>
<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { Done } from 'bkui-vue/lib/icon';
import useGlobalStore from '../../../../../../../../store/global';
import useTemplateStore from '../../../../../../../../store/template';
import { IPackagesCitedByApps } from '../../../../../../../../../types/template';
import { getUnNamedVersionAppsBoundByPackages } from '../../../../../../../../api/template';
import LinkToApp from '../../../../components/link-to-app.vue';

const { spaceId } = storeToRefs(useGlobalStore());
const { currentTemplateSpace, currentPkg, packageList } = storeToRefs(useTemplateStore());

const props = defineProps<{
  show: boolean;
  pending: boolean;
}>();

const emits = defineEmits(['update:show', 'confirm']);

const router = useRouter();

const selectedPkgs = ref<number[]>([]);
const formRef = ref();
const loading = ref(false);
const citedList = ref<IPackagesCitedByApps[]>([]);

const tips = computed(() => (selectedPkgs.value.includes(0)
  ? '若未指定套餐，此配置文件模板将无法被服务引用。后续请使用「添加至」或「添加已有配置文件」功能添加至指定套餐'
  : '以下服务配置的未命名版本引用目标套餐的内容也将更新'));

const maxTableHeight = computed(() => {
  const windowHeight = window.innerHeight;
  return windowHeight * 0.6 - 200;
});

// 未指定套餐选项是否选中
const unSpecifiedSelected = computed(() => selectedPkgs.value.includes(0));

watch(
  () => props.show,
  (val) => {
    if (val) {
      selectedPkgs.value = typeof currentPkg.value === 'number' ? [currentPkg.value] : [];
      if (selectedPkgs.value.length > 0) {
        getCitedData();
      }
    }
  },
);

const allOptions = computed(() => {
  const pkgs = packageList.value.map((item) => {
    const { id, spec } = item;
    return { id, name: spec.name };
  });
  pkgs.push({ id: 0, name: '未指定套餐' });

  return pkgs;
});

const getCitedData = async () => {
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

const handleSelectPkg = (val: number[]) => {
  if (val.length === 0) {
    selectedPkgs.value = [];
    citedList.value = [];
    return;
  }

  if (unSpecifiedSelected.value) {
    selectedPkgs.value = val.filter(id => id !== 0);
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
};

const handleConfirm = async () => {
  const isValid = await formRef.value.validate();
  if (!isValid) return;
  emits('confirm', selectedPkgs.value);
};

const close = () => {
  emits('update:show', false);
};

const goToConfigPageImport = (id: number) => {
  const { href } = router.resolve({
    name: 'service-config',
    params: { appId: id },
    query: { pkg_id: currentTemplateSpace.value },
  });
  window.open(href, '_blank');
};
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
  margin: 0 0 16px;
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
}
</style>
