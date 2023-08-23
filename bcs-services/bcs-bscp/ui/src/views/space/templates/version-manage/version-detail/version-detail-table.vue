<script lang="ts" setup>
  import { ref, computed, watch } from 'vue'
  import { AngleDoubleRightLine } from 'bkui-vue/lib/icon'
  import { IPagination } from '../../../../../../types/index';
  import { ITemplateVersionItem } from '../../../../../../types/template'
  import List from './list.vue'
  import VersionEditor from './version-editor.vue'

  const props = defineProps<{
    spaceId: string;
    templateSpaceId: number;
    templateId: number;
    list: ITemplateVersionItem[];
    pagination: IPagination;
    type: string;
    versionId: number;
  }>()

  const emits = defineEmits(['close', 'refresh', 'select'])

  const versionName = ref('')
  const templateName = ref('')

  const versionEditingData = computed(() => {
    let data = {
      revision_memo: '',
      file_type: '',
      file_mode: "unix",
      user: '',
      user_group: '',
      privilege: '',
      sign: '',
      byte_size: 0
    }
    if (props.versionId) {
      const version = props.list.find(item => item.id === props.versionId)
      if (version) {
        const { revision_memo, file_type, file_mode, content_spec, permission } = version.spec
        const { signature: sign, byte_size } = content_spec
        const { user, user_group, privilege } = permission
        data = { revision_memo, file_type, file_mode, user, user_group, privilege, sign, byte_size }
      }
    }
    return data
  })

  watch(() => props.versionId, val => {
    if (val) {
      const version = props.list.find(item => item.id === val)
      if (version) {
        versionName.value = version.spec.revision_name
        templateName.value = version.spec.name
      }
    }
  }, { immediate: true })

  const handleCreated = (id: number) => {
    emits('refresh')
    emits('select', id)
  }

</script>
<template>
  <div class="version-detail-table">
    <div class="version-list-container">
      <bk-button
        class="close-btn"
        text
        theme="primary"
        @click="emits('close')">
        展开列表
        <AngleDoubleRightLine class="arrow-icon" />
      </bk-button>
      <List
        :list="props.list"
        :pagination="props.pagination"
        :id="props.type === 'create' ? 0 : props.versionId"
        @select="emits('select', $event)" />
    </div>
    <div class="version-detail-content">
      <VersionEditor
        :spaceId="props.spaceId"
        :template-space-id="props.templateSpaceId"
        :templateId="props.templateId"
        :version-id="props.versionId"
        :version-name="versionName"
        :template-name="templateName"
        :data="versionEditingData"
        :type="props.type"
        @created="handleCreated"
        @close="emits('close')" />
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .version-detail-table {
    display: flex;
    align-items: flex-start;
    height: 100%;
    background: #ffffff;
    overflow: hidden;
  }
  .version-list-container {
    position: relative;
    width: 216px;
    height: 100%;
    border: 1px solid #dcdee5;
    border-radius: 2px 0 0 2px;
  }
  .close-btn {
    position: absolute;
    top: 18px;
    right: 10px;
    font-size: 12px;
    z-index: 1;
    .arrow-icon {
      margin-left: 4px;
    }
  }
  .version-detail-content {
    width: calc(100% - 216px);
    height: 100%;
  }
</style>
