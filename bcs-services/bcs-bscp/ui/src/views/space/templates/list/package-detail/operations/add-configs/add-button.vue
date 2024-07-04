<template>
  <bk-popover
    ref="buttonRef"
    theme="light add-configs-button-popover"
    placement="bottom-end"
    trigger="click"
    :arrow="false">
    <bk-button theme="primary" class="create-config-btn">
      <Plus class="button-icon" />
      {{ t('添加配置文件') }}
    </bk-button>
    <template #content>
      <div class="add-config-operations">
        <div class="operation-item" @click="handleOpenSlider('isCreateOpen')">{{ t('新建配置文件') }}</div>
        <div
          v-if="props.showAddExistingConfigOption && packageGroups.length > 0"
          class="operation-item"
          @click="handleOpenSlider('isAddOpen')">
          {{ t('添加已有配置文件') }}
        </div>
        <div class="operation-item" @click="handleOpenSlider('isImportOpen')">{{ t('批量上传配置文件') }}</div>
      </div>
    </template>
  </bk-popover>
  <AddFromExistingConfigs v-model:show="silders.isAddOpen" :groups="packageGroups" @added="emits('refresh')" />
  <CreateNewConfig v-model:show="silders.isCreateOpen" @added="emits('refresh', true)" />
  <ImportConfigs v-model:show="silders.isImportOpen" @added="emits('refresh', true)" />
</template>
<script lang="ts" setup>
  import { onMounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../store/template';
  import { Plus } from 'bkui-vue/lib/icon';
  import AddFromExistingConfigs from './add-from-existing-configs/index.vue';
  import CreateNewConfig from './create-new-config/index.vue';
  import ImportConfigs from './import-configs/index.vue';
  import { getTemplatesBySpaceId, getTemplatePackageList } from '../../../../../../../api/template';
  import { ITemplatePackageItem, ITemplateConfigItem } from '../../../../../../../../types/template';

  interface IPackageTableGroup {
    id: number | string;
    name: string;
    configs: ITemplateConfigItem[];
  }
  const { t } = useI18n();

  const props = defineProps<{
    showAddExistingConfigOption?: boolean;
  }>();

  const emits = defineEmits(['refresh']);

  const { spaceId } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace, currentPkg } = storeToRefs(useTemplateStore());
  const packageGroups = ref<IPackageTableGroup[]>([]); // 所有套餐配置文件数据
  const buttonRef = ref();
  const silders = ref<Record<string, boolean>>({
    isAddOpen: false,
    isCreateOpen: false,
    isImportOpen: false,
  });

  onMounted(() => {
    getGroupConfigs();
  });

  const handleOpenSlider = (slider: string) => {
    silders.value[slider] = true;
    buttonRef.value.hide();
  };

  // 加载全部配置文件
  const getGroupConfigs = async () => {
    const params = {
      start: 0,
      all: true,
    };
    const [packagesRes, configsRes] = await Promise.all([
      getTemplatePackageList(spaceId.value, currentTemplateSpace.value, params),
      getTemplatesBySpaceId(spaceId.value, currentTemplateSpace.value, params),
    ]);
    // 第一个分组默认为“全部配置文件”
    const packages: IPackageTableGroup[] = [
      {
        id: 0,
        name: t('全部配置文件'),
        configs: configsRes.details,
      },
    ];
    packagesRes.details
      .filter((pkg: ITemplatePackageItem) => pkg.id !== currentPkg.value)
      .forEach((pkg: ITemplatePackageItem) => {
        const { name, template_ids } = pkg.spec;
        const pkgGroup: IPackageTableGroup = {
          id: pkg.id,
          name,
          configs: [],
        };
        template_ids.forEach((id) => {
          const config = configsRes.details.find((item: ITemplateConfigItem) => item.id === id);
          if (config) {
            pkgGroup.configs.push(config);
          }
        });
        packages.push(pkgGroup);
      });
    packageGroups.value = packages.slice();
  };
</script>
<style lang="scss" scoped>
  .create-config-btn {
    min-width: 122px;
  }
  .button-icon {
    font-size: 18px;
  }
</style>
<style lang="scss">
  .add-configs-button-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    .add-config-operations {
      .operation-item {
        padding: 0 12px;
        min-width: 58px;
        height: 32px;
        line-height: 32px;
        color: #63656e;
        font-size: 12px;
        cursor: pointer;
        &:hover {
          background: #f5f7fa;
        }
      }
    }
  }
</style>
