<template>
  <bcs-resize-layout :border="false" :min="280" initial-divide="30%">
    <template #aside>
      <div class="h-full overflow-auto">
        <bcs-big-tree
          :data="pathTreeData"
          default-expand-all
          show-link-line
          selectable
          :default-selected-node="curSelectedFilePath"
          ref="treeRef"
          @select-change="handleSelectNodeChange">
          <template #default="{ node, data }">
            <span
              :class="[
                'mr-[5px]',
                {
                  'bk-icon': true,
                  'icon-file': !data.isFolder,
                  'icon-folder-open': node.expanded && data.isFolder,
                  'icon-folder': !node.expanded && data.isFolder,
                }
              ]">
            </span>
            <span>{{data.name}}</span>
          </template>
        </bcs-big-tree>
      </div>
    </template>
    <template #main>
      <CodeEditor
        :value="fileContent"
        :options="{
          roundedSelection: false,
          scrollBeyondLastLine: false,
          renderLineHighlight: 'none',
        }"
        full-screen
        readonly>
      </CodeEditor>
    </template>
  </bcs-resize-layout>
</template>
<script lang="ts">
import { computed, defineComponent, ref, toRefs, watch } from 'vue';

import { path2Tree } from '@/common/util';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';

export default defineComponent({
  name: 'ChartFileTree',
  components: { CodeEditor },
  props: {
    contents: {
      type: Object,
      default: () => ({}),
      required: true,
    },
  },
  setup(props) {
    const { contents } = toRefs(props);

    const treeRef = ref<any>(null);
    const curSelectedFilePath = ref('');
    const pathTreeData = computed(() => path2Tree(Object.keys(contents.value || {})));
    const fileContent = computed(() => contents.value?.[curSelectedFilePath.value]?.content);
    watch(contents, () => {
      curSelectedFilePath.value = Object.keys(contents.value || {})[0] || '';
      setTimeout(() => {
        treeRef.value?.setSelected(curSelectedFilePath.value);
      });
    });
    const handleSelectNodeChange = ({ data }) => {
      if (data.isFolder) return treeRef.value?.setSelected(curSelectedFilePath.value);
      curSelectedFilePath.value = data.id;
    };

    return {
      curSelectedFilePath,
      treeRef,
      pathTreeData,
      fileContent,
      handleSelectNodeChange,
    };
  },
});
</script>
