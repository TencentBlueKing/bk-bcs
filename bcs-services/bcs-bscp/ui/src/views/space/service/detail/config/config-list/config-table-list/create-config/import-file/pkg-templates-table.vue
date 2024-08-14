<template>
  <div :class="['package-template-table', { expand }]">
    <div class="header" @click="handleToggleExpand">
      <RightShape class="arrow-icon" />
      <span v-overflow-title class="name">{{ title }}</span>
      <Close v-if="!props.disabled" class="close-icon" @click.stop="handleDeletePkg" />
    </div>
    <div v-if="expand" v-bkloading="{ loading: listLoading }" class="template-table">
      <div class="table-head">
        <div class="th-cell">{{ t('配置文件绝对路径') }}</div>
        <div class="th-cell">{{ t('版本号') }}</div>
      </div>
      <RecycleScroller
        v-if="configTemplateList.length > 0"
        class="table-body"
        :items="configTemplateList"
        :item-size="42"
        key-field="id">
        <template #default="{ item }">
          <div class="table-row">
            <div class="td-cell">
              <bk-overflow-title class="cell" type="tips">{{ fileAP(item) }}</bk-overflow-title>
            </div>
            <div class="td-cell select-version">
              <bk-select
                :clearable="false"
                :popover-options="{ theme: 'light bk-select-popover add-config-selector-popover' }"
                :model-value="getVersionSelectVal(item.id)"
                @change="handleSelectVersion(item.id, item.versions, $event)">
                <bk-option
                  v-for="version in item.versions"
                  :key="version.isLatest ? 0 : version.id"
                  :id="version.isLatest ? 0 : version.id"
                  :label="version.name">
                  <div
                    v-bk-tooltips="{
                      disabled: !version.memo,
                      content: version.memo,
                    }"
                    class="version-name">
                    {{ version.name }}
                  </div>
                </bk-option>
              </bk-select>
            </div>
          </div>
        </template>
      </RecycleScroller>
      <bk-exception v-else class="empty-tips" scene="part" type="empty">{{ t('该套餐下暂无模板') }}</bk-exception>
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, onMounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { RightShape, Close } from 'bkui-vue/lib/icon';
  import { IAllPkgsGroupBySpaceInBiz, ITemplateConfigItem } from '../../../../../../../../../../types/template';
  import { getTemplateConfigFromPkgId } from '../../../../../../../../../api/template';

  interface ITemplateVersions {
    id: number;
    name: string;
    memo: string;
    isLatest: boolean;
  }

  interface ITemplateConfigWithVersions extends ITemplateConfigItem {
    versions: ITemplateVersions[];
  }
  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    pkgList: IAllPkgsGroupBySpaceInBiz[];
    pkgId: number;
    selectedVersions: { template_id: number; template_revision_id: number; is_latest: boolean }[];
    disabled?: boolean;
    conflictTpls?: number[];
  }>();

  const emits = defineEmits(['delete', 'expand', 'selectVersion', 'updateVersions']);

  const listLoading = ref(false);
  const configTemplateList = ref<ITemplateConfigWithVersions[]>([]);
  const title = ref('');
  const templateSpaceId = ref(0);
  const expand = ref(true);

  onMounted(async () => {
    props.pkgList.some((templateSpace) =>
      templateSpace.template_sets.some((pkg) => {
        if (pkg.template_set_id === props.pkgId) {
          title.value = `${templateSpace.template_space_name} - ${pkg.template_set_name}`;
          templateSpaceId.value = templateSpace.template_space_id;
        }
        return undefined;
      }),
    );
    await getTemplateList();
    setTemplatesDefaultVersion();
  });

  // 配置文件绝对路径
  const fileAP = computed(() => (config: ITemplateConfigWithVersions) => {
    const { path, name } = config.spec;
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  });

  // 获取模板及对应版本列表
  const getTemplateList = async () => {
    listLoading.value = true;
    try {
      const res = await getTemplateConfigFromPkgId(props.bkBizId, props.pkgId);
      configTemplateList.value = res.details.map((item: any) => {
        const { template, template_revision } = item;
        const versions = template_revision.map((version: any) => {
          const {
            id,
            spec: { revision_name, revision_memo },
          } = version;
          return {
            id,
            name: revision_name,
            memo: revision_memo,
            isLatest: false,
          };
        });
        const latestVerison = versions[0];
        if (latestVerison) {
          versions.unshift({
            id: latestVerison.id,
            name: t('latest（当前最新为{n}）', { n: latestVerison?.name }),
            memo: latestVerison.memo,
            isLatest: true,
          });
        }
        return {
          ...template,
          versions,
        };
      });
    } catch (error) {}
    listLoading.value = false;
  };

  // 如果有模板没有选择版本则自动选择latest版本
  const setTemplatesDefaultVersion = () => {
    const selectedTplVersionsData = props.selectedVersions.slice();
    configTemplateList.value.forEach((tpl) => {
      if (!props.selectedVersions.find((item) => item.template_id === tpl.id)) {
        const lasteVersion = tpl.versions.find((v) => v.isLatest);
        if (lasteVersion) {
          selectedTplVersionsData.push({
            template_id: tpl.id,
            template_revision_id: lasteVersion.id,
            is_latest: true,
          });
        }
      }
    });
    if (selectedTplVersionsData.length !== props.selectedVersions.length) {
      emits('updateVersions', selectedTplVersionsData);
    }
  };

  const getVersionSelectVal = (id: number) => {
    const version = props.selectedVersions.find((item) => item.template_id === id);
    if (version) {
      return version.is_latest ? 0 : version.template_revision_id;
    }
    return '';
  };

  const handleToggleExpand = () => {
    expand.value = !expand.value;
    if (expand.value) {
      getTemplateList();
    }
  };

  const handleSelectVersion = (
    tplId: number,
    versions: { id: number; name: string; isLatest: boolean }[],
    val: number,
  ) => {
    const isLatest = val === 0;
    const versionId = isLatest ? versions.find((item) => item.isLatest)?.id : val;
    const versionData = {
      template_id: tplId,
      template_revision_id: versionId,
      is_latest: isLatest,
    };
    emits('selectVersion', versionData);
  };

  const handleDeletePkg = () => {
    emits('delete', props.pkgId);
  };
</script>
<style lang="scss" scoped>
  .package-template-table {
    margin-bottom: 18px;
    &.expand {
      .arrow-icon {
        transform: rotate(90deg);
      }
    }
    .header {
      display: flex;
      align-items: center;
      position: relative;
      padding: 0 9px;
      height: 24px;
      background: #eaebf0;
      font-size: 12px;
      color: #63656e;
      border-radius: 2px 2px 0 0;
      cursor: pointer;
    }
    .arrow-icon {
      font-size: 12px;
      color: #979ba5;
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
    .name {
      margin-left: 5px;
      max-width: calc(100% - 70px);
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .close-icon {
      position: absolute;
      top: 5px;
      right: 5px;
      font-size: 14px;
      color: #979ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
  .template-table {
    width: 100%;
    border: 1px solid #dcdee5;
    .table-head {
      display: flex;
      width: 100%;
    }
    .table-body {
      max-height: calc(42px * 7);
    }
    .table-row {
      display: flex;
      width: 100%;
    }
    .td-cell.has-conflict {
      .cell {
        background: #fff3e1;
      }
      .select-version {
        :deep(.bk-input--text) {
          background: #fff3e1;
        }
      }
    }
    .th-cell,
    .td-cell {
      width: 454px;
      line-height: 20px;
      font-size: 12px;
      font-weight: 400;
      text-align: left;
      &:not(:last-child) {
        border-right: 1px solid #dcdee5;
      }
    }
    .th-cell {
      padding: 11px 16px;
      color: #313238;
      background: #fafbfd;
      border-bottom: none;
    }
    .td-cell {
      padding: 0;
      color: #63656e;
      background: #f5f7fa;
      .cell {
        padding: 11px 16px;
        height: 42px;
        line-height: 20px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        border-top: 1px solid #dcdee5;
        width: 452px;
      }
    }
    .select-version {
      padding: 0;
      :deep(.bk-input) {
        height: 42px;
        border-color: transparent;
        border-top: 1px solid #dcdee5;
      }
    }
    .empty-tips {
      margin: 20px 0;
      font-size: 12px;
      color: #3a84ff;
    }
  }
</style>

<style lang="scss">
  .add-config-selector-popover {
    width: auto !important;
  }
</style>
