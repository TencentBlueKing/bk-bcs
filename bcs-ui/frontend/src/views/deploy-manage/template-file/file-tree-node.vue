<template>
  <div>
    <!-- 文件夹节点 -->
    <div
      v-if="node.isFolder"
      :class="[
        'flex items-center justify-between h-[32px] text-[12px] cursor-pointer hover:bg-[#F5F7FA]',
      ]"
      :style="{ paddingLeft: `${depth * 16 + 20}px`, paddingRight: '8px' }"
      @click="handleToggleExpand"
      @mouseenter="hoverItemName = node.name"
      @mouseleave="hoverItemName = ''">
      <span class="flex items-center min-w-0">
        <!-- 展开箭头 -->
        <i
          :class="[
            'bcs-icon text-[16px] text-[#979BA5] mr-[4px] transition-transform shrink-0',
            node.expanded ? 'bcs-icon-down-shape' : 'bcs-icon-right-shape',
          ]"
          style="transform: scale(0.8);"
        />
        <!-- 文件夹图标 -->
        <i
          :class="[
            'bcs-icon text-[14px] mr-[4px] flex-shrink-0',
            node.expanded ? 'bcs-icon-folder-open text-[#3A84FF]' : 'bcs-icon-folder text-[#C8C9CC]',
          ]"
        />
        <!-- 文件夹名称（支持重命名） -->
        <span
          v-if="!isRenaming"
          class="bcs-ellipsis"
          v-bk-overflow-tips="{ interactive: false }">{{ node.name }}</span>
        <input
          v-else
          ref="renameInputRef"
          v-model.trim="renameValue"
          :class="[
            'h-[22px] text-[12px] border-[1px] border-[#3A84FF]',
            'outline-none px-[4px] rounded-sm w-full min-w-[60px]'
          ]"
          @click.stop
          @keyup.enter="confirmRename"
          @keyup.escape="cancelRename"
          @blur="confirmRename" />
      </span>
      <!-- 文件夹操作 -->
      <span v-if="hoverItemName === node.name || curPopover === node.name" @click.stop>
        <PopoverSelector offset="0, 6" :on-hide="hidePopover" :on-show="() => showPopover(node.name)">
          <span class="bcs-icon-more-btn w-[16px] h-[16px]">
            <i class="bcs-icon bcs-icon-more"></i>
          </span>
          <template #content>
            <ul class="bg-[#fff]">
              <li class="bcs-dropdown-item" @click="handleAddSubFolder">
                {{ $t('templateFile.button.addSubFolder') }}
              </li>
              <li class="bcs-dropdown-item" @click="handleAddTemplateFile">
                {{ $t('templateFile.button.createFile') }}
              </li>
              <li v-if="node.isTemp" class="bcs-dropdown-item" @click="startRename">
                {{ $t('templateFile.button.rename') }}
              </li>
              <li v-if="node.isTemp" class="bcs-dropdown-item" @click="handleDeleteFolder">
                {{ $t('templateFile.button.delete') }}
              </li>
            </ul>
          </template>
        </PopoverSelector>
      </span>
    </div>
    <!-- 文件夹展开的子节点 -->
    <div v-if="node.isFolder && node.expanded">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        :cur-file-i-d="curFileID"
        :space-i-d="spaceID"
        @toggle-expand="$emit('toggle-expand', $event)"
        @add-sub-folder="$emit('add-sub-folder', $event)"
        @rename-folder="$emit('rename-folder', $event)"
        @delete-folder="$emit('delete-folder', $event)"
        @select-file="$emit('select-file', $event)"
        @edit-file="$emit('edit-file', $event)"
        @deploy-file="$emit('deploy-file', $event)"
        @clone-version="$emit('clone-version', $event)"
        @manage-version="$emit('manage-version', $event)"
        @delete-file="$emit('delete-file', $event)"
      />
    </div>
    <!-- 文件节点 -->
    <div
      v-if="!node.isFolder"
      :class="[
        'flex items-center justify-between h-[32px] text-[12px] hover:bg-[#F5F7FA] cursor-pointer group',
        { '!text-[#3A84FF] !bg-[#E1ECFF]': curFileID === node.file?.id }
      ]"
      :style="{ paddingLeft: `${depth * 16 + 20}px`, paddingRight: '8px' }"
      @click="handleSelectFile"
      @mouseenter="hoverItemName = node.name"
      @mouseleave="hoverItemName = ''">
      <span class="flex items-center min-w-0">
        <i class="bcs-icon bcs-icon-file text-[14px] mr-[4px] flex-shrink-0 text-[#C8C9CC]" />
        <span class="bcs-ellipsis" v-bk-overflow-tips="{ interactive: false }">{{ node.name }}</span>
      </span>
      <!-- 文件操作 -->
      <span v-if="hoverItemName === node.name || curPopover === node.name" @click.stop>
        <PopoverSelector offset="0, 6" :on-hide="hidePopover" :on-show="() => showPopover(node.name)">
          <span class="bcs-icon-more-btn w-[16px] h-[16px] opacity-0 group-hover:opacity-100">
            <i class="bcs-icon bcs-icon-more"></i>
          </span>
          <template #content>
            <ul class="bg-[#fff]">
              <li class="bcs-dropdown-item" @click="handleEditFile">
                {{ $t('generic.button.edit') }}
              </li>
              <li
                class="bcs-dropdown-item"
                :class="{ 'disabled': !node.file?.version }"
                @click="handleDeployFile">
                {{ $t('templateSet.button.deploy') }}
              </li>
              <li class="bcs-dropdown-item" @click="handleCloneVersion">
                {{ $t('generic.button.clone') }}
              </li>
              <li class="bcs-dropdown-item" @click="handleManageVersion">
                {{ $t('templateFile.button.versionManage') }}
              </li>
              <li class="bcs-dropdown-item" @click="handleDeleteFile">
                {{ $t('generic.button.delete') }}
              </li>
            </ul>
          </template>
        </PopoverSelector>
      </span>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, nextTick, PropType, ref } from 'vue';

import { IFileTreeNode } from './use-file-tree';

import PopoverSelector from '@/components/popover-selector.vue';
import $router from '@/router';

export default defineComponent({
  name: 'FileTreeNode',
  components: { PopoverSelector },
  props: {
    node: {
      type: Object as PropType<IFileTreeNode>,
      required: true,
    },
    depth: {
      type: Number as PropType<number>,
      default: 0,
    },
    curFileID: {
      type: String as PropType<string>,
      default: '',
    },
    spaceID: {
      type: String as PropType<string>,
      default: '',
    },
  },
  emits: [
    'toggle-expand',
    'add-sub-folder',
    'rename-folder',
    'delete-folder',
    'select-file',
    'edit-file',
    'deploy-file',
    'clone-version',
    'manage-version',
    'delete-file',
  ],
  setup(props, { emit }) {
    const hoverNode = ref(false);
    const isRenaming = ref(false);
    const renameValue = ref('');
    const renameInputRef = ref<HTMLInputElement>();
    const hoverItemName = ref('');

    // 当前文件夹是否处于激活状态（子节点中有选中的文件）
    const isActiveFolder = computed(() => {
      if (!props.curFileID || !props.node.isFolder) return false;
      function hasActiveFile(node: IFileTreeNode): boolean {
        if (!node.isFolder && node.file?.id === props.curFileID) return true;
        return node.children.some(child => hasActiveFile(child));
      }
      return hasActiveFile(props.node);
    });

    function handleToggleExpand() {
      if (props.node.isFolder) {
        emit('toggle-expand', props.node.path);
      }
    }

    function handleSelectFile() {
      if (!props.node.isFolder && props.node.file) {
        emit('select-file', { spaceID: props.spaceID, file: props.node.file });
      }
    }

    // 操作菜单
    const curPopover = ref('');
    function showPopover(name: string) {
      curPopover.value = name;
    }
    function hidePopover() {
      curPopover.value = '';
    }

    function handleAddSubFolder() {
      hidePopover();
      emit('add-sub-folder', props.node.path);
    }

    function startRename() {
      hidePopover();
      isRenaming.value = true;
      renameValue.value = props.node.name;
      nextTick(() => {
        renameInputRef.value?.focus();
        renameInputRef.value?.select();
      });
    }

    function confirmRename() {
      if (!isRenaming.value) return;
      if (renameValue.value && renameValue.value !== props.node.name) {
        emit('rename-folder', { path: props.node.path, newName: renameValue.value });
      }
      isRenaming.value = false;
    }

    function cancelRename() {
      isRenaming.value = false;
    }

    function handleDeleteFolder() {
      hidePopover();
      emit('delete-folder', props.node.path);
    }

    // 新增模板文件
    function handleAddTemplateFile() {
      hidePopover();

      if (!props.spaceID) return;
      $router.push({
        name: 'addTemplateFile',
        params: {
          templateSpace: props.spaceID,
        },
        query: {
          folderPath: props.node.path,
        },
      });
    }

    // 文件操作
    function handleEditFile() {
      hidePopover();
      emit('edit-file', props.node.file);
    }

    function handleDeployFile() {
      if (!props.node.file?.version) return;
      hidePopover();
      emit('deploy-file', props.node.file);
    }

    function handleCloneVersion() {
      hidePopover();
      emit('clone-version', props.node.file);
    }

    function handleManageVersion() {
      hidePopover();
      emit('manage-version', props.node.file);
    }

    function handleDeleteFile() {
      hidePopover();
      emit('delete-file', props.node.file);
    }

    return {
      hoverItemName,
      hoverNode,
      isRenaming,
      renameValue,
      renameInputRef,
      isActiveFolder,
      curPopover,
      handleToggleExpand,
      handleSelectFile,
      showPopover,
      hidePopover,
      handleAddSubFolder,
      handleAddTemplateFile,
      startRename,
      confirmRename,
      cancelRename,
      handleDeleteFolder,
      handleEditFile,
      handleDeployFile,
      handleCloneVersion,
      handleManageVersion,
      handleDeleteFile,
    };
  },
});
</script>
