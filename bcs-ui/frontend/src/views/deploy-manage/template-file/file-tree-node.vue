<template>
  <div v-show="isVisible">
    <!-- 文件夹节点 -->
    <div
      v-if="node.isFolder"
      :class="[
        'flex items-center cursor-pointer hover:bg-[#F5F7FA]',
        mode === 'batch' ? 'h-[32px] rounded-sm hover:bg-[#F0F1F5]' : 'h-[32px] justify-between',
      ]"
      :style="indentStyle"
      @click="handleToggleExpand"
      @mouseenter="hoverItemName = node.name"
      @mouseleave="hoverItemName = ''">
      <span class="flex items-center min-w-0">
        <!-- 展开箭头 -->
        <i
          :class="[
            'bcs-icon text-[16px] text-[#979BA5] mr-[4px] transition-transform shrink-0',
            isExpanded ? 'bcs-icon-down-shape' : 'bcs-icon-right-shape',
          ]"
          style="transform: scale(0.8);"
          @click.stop="handleToggleExpand()"
        />
        <!-- Checkbox (batch 模式) -->
        <span v-if="mode === 'batch'" class="flex items-center" @click.stop>
          <bk-checkbox
            :indeterminate="indeterminate"
            :checked="isChecked"
            @change="handleCheckChange"
            class="mr-[6px]" />
        </span>
        <!-- 文件夹图标 -->
        <i
          :class="[
            'bcs-icon text-[14px] mr-[4px] flex-shrink-0',
            isExpanded ? 'bcs-icon-folder-open text-[#3A84FF]' : 'bcs-icon-folder',
            mode === 'manage' && !isExpanded ? 'text-[#C8C9CC]' : '',
          ]"
        />
        <!-- 文件夹名称（支持重命名） -->
        <span
          v-if="mode === 'manage' && !isRenaming"
          class="text-[12px] bcs-ellipsis"
          v-bk-overflow-tips="{ interactive: false }"
          v-bk-xss-html="highlightText(node.name)"></span>
        <input
          v-if="mode === 'manage' && isRenaming"
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
        <!-- 文件夹名称（batch 模式） -->
        <span
          v-if="mode === 'batch'"
          class="text-[12px] bcs-ellipsis flex-1"
          v-bk-overflow-tips
          @click.stop="handleToggleExpand()"
          v-bk-xss-html="highlightText(node.name)"></span>
      </span>
      <!-- 文件夹操作 (manage 模式) -->
      <span v-if="mode === 'manage' && (hoverItemName === node.name || curPopover === node.name)" @click.stop>
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
    <div v-if="node.isFolder && isExpanded">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        :mode="mode"
        :checked-paths="checkedPaths"
        :expanded-paths="expandedPaths"
        :cur-file-i-d="curFileID"
        :space-i-d="spaceID"
        :default-expand-all="defaultExpandAll"
        :search-key="searchKey"
        @toggle-check="(...args) => $emit('toggle-check', ...args)"
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
        'flex items-center text-[12px] hover:bg-[#F5F7FA] cursor-pointer group',
        mode === 'batch' ? 'h-[32px] rounded-sm hover:bg-[#F0F1F5]' : 'h-[32px] justify-between',
        { '!text-[#3A84FF] !bg-[#E1ECFF]': mode === 'manage' && curFileID === node.file?.id }
      ]"
      :style="indentStyle"
      @click="handleClickFile"
      @mouseenter="hoverItemName = node.name"
      @mouseleave="hoverItemName = ''">
      <span class="flex items-center min-w-0 pl-[20px]">
        <!-- Checkbox (batch 模式) -->
        <span v-if="mode === 'batch'" @click.stop>
          <bk-checkbox
            :checked="isChecked"
            @change="handleCheckChange"
            class="mr-[6px]" />
        </span>
        <!-- 文件图标 -->
        <i class="bcs-icon bcs-icon-file text-[14px] mr-[4px] flex-shrink-0 text-[#C4C6CC]" />
        <!-- 文件名称 -->
        <span
          :class="['bcs-ellipsis', mode === 'manage' ? '' : 'flex-1']"
          v-bk-overflow-tips="{ interactive: false }"
          v-bk-xss-html="highlightText(node.name)"></span>
      </span>
      <!-- 文件操作 (manage 模式) -->
      <span v-if="mode === 'manage' && (hoverItemName === node.name || curPopover === node.name)" @click.stop>
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
    mode: {
      type: String as PropType<'manage' | 'batch'>,
      default: 'manage',
    },
    // batch 模式 props
    checkedPaths: {
      type: Set as PropType<Set<string>>,
      default: () => new Set(),
    },
    expandedPaths: {
      type: Set as PropType<Set<string>>,
      default: () => new Set(),
    },
    // manage 模式 props
    curFileID: {
      type: String as PropType<string>,
      default: '',
    },
    spaceID: {
      type: String as PropType<string>,
      default: '',
    },
    // 是否默认展开所有文件夹
    defaultExpandAll: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    // 搜索关键字
    searchKey: {
      type: String as PropType<string>,
      default: '',
    },
  },
  emits: [
    'toggle-check',
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
    const hoverItemName = ref('');
    const isRenaming = ref(false);
    const renameValue = ref('');
    const renameInputRef = ref<HTMLInputElement>();
    const curPopover = ref('');
    // defaultExpandAll 模式下，记录手动收起的文件夹路径
    const manuallyCollapsed = ref<Set<string>>(new Set());

    // 缩进样式
    const indentStyle = computed(() => {
      if (props.mode === 'batch') {
        return { paddingLeft: `${props.depth * 16}px` };
      }
      return { paddingLeft: `${props.depth * 16}px`, paddingRight: '8px' };
    });

    // 展开状态（兼容两种模式）
    const isExpanded = computed(() => {
      if (!props.node.isFolder) return false;
      // 搜索时自动展开包含匹配后代的文件夹
      if (props.searchKey && hasMatchingDescendant.value) return true;
      if (props.defaultExpandAll) {
        return !manuallyCollapsed.value.has(props.node.path);
      }
      if (props.mode === 'batch') {
        return props.expandedPaths.has(props.node.path);
      }
      return !!props.node.expanded;
    });

    // 搜索相关
    const lowerSearchKey = computed(() => props.searchKey.toLowerCase());

    // 当前节点名称是否匹配搜索
    const isSelfMatch = computed(() => {
      if (!lowerSearchKey.value) return true;
      return props.node.name.toLowerCase().includes(lowerSearchKey.value)
        || props.node.path.toLowerCase().includes(lowerSearchKey.value);
    });

    // 后代中是否有匹配节点
    const hasMatchingDescendant = computed(() => {
      if (!lowerSearchKey.value) return true;
      function check(node: IFileTreeNode): boolean {
        return node.children.some((child) => {
          const key = lowerSearchKey.value!;
          return child.name.toLowerCase().includes(key)
            || child.path.toLowerCase().includes(key)
            || (child.isFolder && check(child));
        });
      }
      return check(props.node);
    });

    // 当前节点是否可见（搜索过滤）
    const isVisible = computed(() => {
      if (!lowerSearchKey.value) return true;
      return isSelfMatch.value || hasMatchingDescendant.value;
    });

    // 收集节点下所有叶子路径（batch 模式）
    function collectLeafPaths(node: IFileTreeNode): string[] {
      if (!node.isFolder) return [node.path];
      const paths: string[] = [];
      for (const child of node.children) {
        paths.push(...collectLeafPaths(child));
      }
      return paths;
    }

    const leafPaths = computed(() => {
      if (!props.node.isFolder) return [props.node.path];
      return collectLeafPaths(props.node);
    });

    // 选中状态（batch 模式）
    const isChecked = computed(() => {
      if (props.mode !== 'batch') return false;
      if (props.node.isFolder) {
        return leafPaths.value.length > 0 && leafPaths.value.every(p => props.checkedPaths.has(p));
      }
      return props.checkedPaths.has(props.node.path);
    });

    const indeterminate = computed(() => {
      if (props.mode !== 'batch' || !props.node.isFolder) return false;
      const checkedCount = leafPaths.value.filter(p => props.checkedPaths.has(p)).length;
      return checkedCount > 0 && checkedCount < leafPaths.value.length;
    });

    // 当前文件夹是否处于激活状态（manage 模式）
    const isActiveFolder = computed(() => {
      if (props.mode !== 'manage' || !props.curFileID || !props.node.isFolder) return false;
      function hasActiveFile(node: IFileTreeNode): boolean {
        if (!node.isFolder && node.file?.id === props.curFileID) return true;
        return node.children.some(child => hasActiveFile(child));
      }
      return hasActiveFile(props.node);
    });

    // 搜索高亮：返回高亮后的 HTML
    function highlightText(text: string): string {
      if (!lowerSearchKey.value || !text) return text;
      const lowerText = text.toLowerCase();
      const idx = lowerText.indexOf(lowerSearchKey.value);
      if (idx === -1) return text;
      const before = text.slice(0, idx);
      const match = text.slice(idx, idx + lowerSearchKey.value.length);
      const after = text.slice(idx + lowerSearchKey.value.length);
      return `${before}<span class="text-[#3A84FF] font-bold">${match}</span>${after}`;
    }

    function handleToggleExpand() {
      if (props.node.isFolder) {
        if (props.defaultExpandAll) {
          const newSet = new Set(manuallyCollapsed.value);
          if (isExpanded.value) {
            newSet.add(props.node.path);
          } else {
            newSet.delete(props.node.path);
          }
          manuallyCollapsed.value = newSet;
        } else {
          emit('toggle-expand', props.node.path);
        }
      }
    }

    // 文件点击处理
    function handleClickFile() {
      if (props.node.isFolder) return;
      if (props.mode === 'batch') {
        handleCheckChange();
      } else if (props.mode === 'manage' && props.node.file) {
        emit('select-file', { spaceID: props.spaceID, file: props.node.file });
      }
    }

    function handleCheckChange() {
      if (props.mode === 'batch') {
        // indeterminate 状态下点击应取消全选；已全选则取消；未选则全选
        const newChecked = !isChecked.value && !indeterminate.value;
        emit('toggle-check', props.node, newChecked);
      }
    }

    // 操作菜单
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
      emit('clone-version', props.node);
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
      isRenaming,
      renameValue,
      renameInputRef,
      curPopover,
      indentStyle,
      isExpanded,
      leafPaths,
      isChecked,
      indeterminate,
      isActiveFolder,
      isVisible,
      highlightText,
      handleToggleExpand,
      handleClickFile,
      handleCheckChange,
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
