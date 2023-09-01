<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue'
  import { RightShape, Close } from 'bkui-vue/lib/icon'
  import { IAllPkgsGroupBySpaceInBiz, ITemplateConfigItem, ITemplateVersionsName } from '../../../../../../../../../types/template';
  import { getTemplatesByPackageId, getTemplateVersionsNameByIds } from '../../../../../../../../api/template';


  interface ITemplateConfigWithVersions extends ITemplateConfigItem {
    versions: { id: number; name: string; isLatest: boolean; }[]
  }

  const props = defineProps<{
    bkBizId: string;
    pkgList: IAllPkgsGroupBySpaceInBiz[];
    pkgId: number;
    expanded: boolean;
    selectedVersions: { template_id: number; template_revision_id: number; is_latest: boolean; }[];
    disabled?: boolean;
  }>()

  const emits = defineEmits(['delete', 'expand', 'selectVersion'])

  const listLoading = ref(false)
  const configTemplateList = ref<ITemplateConfigWithVersions[]>([])
  const title = ref('')
  const templateSpaceId = ref(0)

  watch(() => props.expanded, val => {
    if (val) {
      getTemplateList()
    }
  })

  onMounted(() => {
    props.pkgList.some(templateSpace => {
      return templateSpace.template_sets.some(pkg => {
        if (pkg.template_set_id === props.pkgId) {
          title.value = `${templateSpace.template_space_name} - ${pkg.template_set_name}`
          templateSpaceId.value = templateSpace.template_space_id
        }
      })
    })
    if (props.expanded) {
      getTemplateList()
    }
  })

  const getTemplateList = async () => {
    listLoading.value = true
    const templateListRes = await getTemplatesByPackageId(props.bkBizId, templateSpaceId.value, props.pkgId, { start: 0, all: true })
    configTemplateList.value = templateListRes.details.map((item: ITemplateConfigItem) => {
      return { ...item, versions: [] }
    })
    const ids = configTemplateList.value.map(item => item.id)
    if (ids.length > 0) {
      const versionListRes = await getTemplateVersionsNameByIds(props.bkBizId, ids)
      versionListRes.details.forEach((item: ITemplateVersionsName) => {
        const { template_id, latest_template_revision_id, template_revisions } = item
        const configTemplate = configTemplateList.value.find(tpl => tpl.id === template_id)
        if (configTemplate) {
          configTemplate.versions = template_revisions.map(version => {
            const { template_revision_id, template_revision_name } = version
            return { id: template_revision_id, name: template_revision_name, isLatest: false }
          })
          const version = template_revisions.find(version => version.template_revision_id === latest_template_revision_id)
          if (version) {
            configTemplate.versions.unshift({
              id: version.template_revision_id,
              name: `latest（当前最新为${version.template_revision_name}）`,
              isLatest: true
            })
          }
        }
      })
    }
    listLoading.value = false
  }

  const getVersionSelectVal = (id: number) => {
    const version = props.selectedVersions.find(item => item.template_id === id)
    if (version) {
      return version.is_latest ? 0 : version.template_revision_id
    }
    return ''
  }

  const handleSelectVersion = (tplId: number,versions: { id: number; name: string; isLatest: boolean; }[], val: number) => {
    const isLatest = val === 0
    const versionId = isLatest ? versions.find(item => item.isLatest)?.id : val
    const versionData = {
      template_id: tplId,
      template_revision_id: versionId,
      is_latest: isLatest
    }
    emits('selectVersion', versionData)
  }

  const handleDeletePkg = () => {
    emits('delete', props.pkgId)
  }

</script>
<template>
  <div :class="['package-template-table', { expand: props.expanded }]">
    <div class="header" @click="emits('expand', props.pkgId)">
      <RightShape class="arrow-icon" />
      <span v-overflow-title class="name">{{ title }}</span>
      <Close v-if="!props.disabled" class="close-icon" @click.stop="handleDeletePkg"/>
    </div>
    <table v-if="props.expanded" v-bkloading="{ loading: listLoading }" class="template-table">
      <thead>
        <tr>
          <th>模板名称</th>
          <th>模板路径</th>
          <th>版本号</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="configTemplateList.length > 0">
          <tr v-for="tpl in configTemplateList" :key="tpl.id">
            <td><div class="cell">{{ tpl.spec.name }}</div></td>
            <td><div class="cell">{{ tpl.spec.path }}</div></td>
            <td class="select-version">
              <bk-select
                :clearable="false"
                :model-value="getVersionSelectVal(tpl.id)"
                :disabled="props.disabled"
                @change="handleSelectVersion(tpl.id, tpl.versions, $event)">
                <bk-option
                  v-for="version in tpl.versions"
                  :key="version.isLatest ? 0 : version.id"
                  :id="version.isLatest ? 0 : version.id"
                  :label="version.name">
                </bk-option>
              </bk-select>
            </td>
          </tr>
        </template>
        <tr v-else>
          <td colspan="3">
            <bk-exception class="empty-tips" scene="part" type="empty">该套餐下暂无模板</bk-exception>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
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
      border-radius: 2px 2px 0 0 ;
      cursor: pointer;
    }
    .arrow-icon {
      font-size: 12px;
      color: #979ba5;
      transition: transform .3s cubic-bezier(.4,0,.2,1);
    }
    .name {
      margin-left: 5px;
      max-width: calc(100% - 40px);
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
    border-collapse: collapse;
    table-layout: fixed;
    th,td {
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
    }
  }
</style>
