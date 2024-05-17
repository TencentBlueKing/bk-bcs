<template>
  <div :class="['package-template-table', { expand }]">
    <div class="header" @click="handleToggleExpand">
      <RightShape class="arrow-icon" />
      <span v-overflow-title class="name">{{ title }}</span>
      <Close v-if="!props.disabled" class="close-icon" @click.stop="handleDeletePkg" />
    </div>
    <table v-if="expand" v-bkloading="{ loading: listLoading }" class="template-table">
      <thead>
        <tr>
          <th>{{ t('配置文件绝对路径') }}</th>
          <th>{{ t('版本号') }}</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="configTemplateList.length > 0">
          <tr v-for="tpl in configTemplateList" :key="tpl.id">
            <td>
              <bk-overflow-title class="cell" type="tips">{{ fileAP(tpl) }}</bk-overflow-title>
            </td>

            <td class="select-version">
              <bk-select
                :clearable="false"
                :popover-options="{ theme: 'light bk-select-popover add-config-selector-popover' }"
                :model-value="getVersionSelectVal(tpl.id)"
                @change="handleSelectVersion(tpl.id, tpl.versions, $event)">
                <bk-option
                  v-for="version in tpl.versions"
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
            </td>
          </tr>
        </template>
        <tr v-else>
          <td colspan="3">
            <bk-exception class="empty-tips" scene="part" type="empty">{{ t('该套餐下暂无模板') }}</bk-exception>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
<script lang="ts" setup>
  import { ref, onMounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { RightShape, Close } from 'bkui-vue/lib/icon';
  import {
    IAllPkgsGroupBySpaceInBiz,
    ITemplateConfigItem,
    ITemplateVersionsName,
  } from '../../../../../../../../../../types/template';
  import { getTemplatesByPackageId, getTemplateVersionsNameByIds } from '../../../../../../../../../api/template';

  interface ITemplateConfigWithVersions extends ITemplateConfigItem {
    versions: { id: number; name: string; memo: string; isLatest: boolean }[];
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
    // 先取套餐下模板列表
    const templateListRes = await getTemplatesByPackageId(props.bkBizId, templateSpaceId.value, props.pkgId, {
      start: 0,
      all: true,
    });
    configTemplateList.value = templateListRes.details.map((item: ITemplateConfigItem) => ({ ...item, versions: [] }));
    const ids = configTemplateList.value.map((item) => item.id);
    if (ids.length > 0) {
      // 再根据模板列表取对应模板的版本列表
      const versionListRes = await getTemplateVersionsNameByIds(props.bkBizId, ids);
      versionListRes.details.forEach((item: ITemplateVersionsName) => {
        const { template_id, latest_template_revision_id, template_revisions } = item;
        const configTemplate = configTemplateList.value.find((tpl) => tpl.id === template_id);
        if (configTemplate) {
          configTemplate.versions = template_revisions.map((version) => {
            const { template_revision_id, template_revision_name, template_revision_memo } = version;
            return {
              id: template_revision_id,
              name: template_revision_name,
              memo: template_revision_memo,
              isLatest: false,
            };
          });
          const version = template_revisions.find((version) => {
            const res = version.template_revision_id === latest_template_revision_id;
            return res;
          });
          if (version) {
            configTemplate.versions.unshift({
              id: version.template_revision_id,
              name: `latest（当前最新为${version.template_revision_name}）`,
              memo: version.template_revision_memo,
              isLatest: true,
            });
          }
        }
      });
    }
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
    &:not(:last-child) {
      margin-bottom: 18px;
    }
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
    .conflict-icon {
      margin-left: 10px;
      font-size: 14px;
      color: #ff9c01;
      cursor: pointer;
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
    border-collapse: collapse;
    table-layout: fixed;
    tr.has-conflict {
      .cell {
        background: #fff3e1;
      }
      .select-version {
        :deep(.bk-input--text) {
          background: #fff3e1;
        }
      }
    }
    th,
    td {
      line-height: 20px;
      font-size: 12px;
      font-weight: 400;
      border: 1px solid #dcdee5;
      text-align: left;
    }
    th {
      padding: 11px 16px;
      color: #313238;
      background: #fafbfd;
    }
    td {
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
      }
    }
    .select-version {
      padding: 0;
      :deep(.bk-input) {
        height: 42px;
        border-color: transparent;
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
