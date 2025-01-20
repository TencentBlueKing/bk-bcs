<template>
  <div class="flex items-center">
    <template v-if="!editable">
      <span class="flex items-center" v-if="isSelectAll">
        <span class="bg-[#1768EF] text-[white] rounded-sm h-[14px] leading-[10px] p-[2px] mr-[5px]">ALL</span>
        <span>{{ $t('cluster.msg.allProject') }}</span>
      </span>
      <span
        v-else
        class="break-all clamp-text"
        v-bk-overflow-tips="{ content: projectNameList.join() }">
        <span v-for="(item, index) in projectNameList" :key="item">
          <span>{{ item }}</span>
          <bcs-tag class="m-0 px-[5px]" theme="info" v-if="isOnlyCurrentPorject">
            {{ $t('cluster.tag.onlyCurrentProject') }}</bcs-tag>
          <bcs-tag
            class="m-0 px-[5px]"
            theme="info"
            v-else-if="curProject?.name === item
              || curProject?.projectID === item">{{ $t('cluster.tag.currentProject') }}</bcs-tag>
          <span v-if="index < projectNameList.length - 1">, </span>
        </span>
        <span v-if="projectNameList.length === 0">--</span>
      </span>
      <span
        class="hover:text-[#3a84ff] cursor-pointer ml-[8px]"
        v-if="!disableEdit"
        @click="handleEdit">
        <i class="bk-icon icon-edit-line"></i>
      </span>
    </template>
    <template v-else>
      <div class="flex-1 flex items-center">
        <bcs-select
          class="flex-1 max-w-[400px]"
          clearable
          searchable
          multiple
          show-on-init
          :value="innerValue"
          :popover-min-width="320"
          enable-scroll-load
          :scroll-loading="{
            isLoading: scrollLoading
          }"
          :loading="loading"
          ref="selectRef"
          :allow-enter="false"
          @selected="handleProjectChange"
          @clear="innerValue = []"
          @scroll-end="handleScrollToBottom">
          <template #search>
            <!-- <div
              :class="[
                'flex items-center justify-between cursor-pointer hover:bg-[#eaf3ff] px-[10px]',
                isSelectAll ? 'bg-[#eaf3ff] text-[#3a84ff]' : ''
              ]"
              @click.stop="handleSelectAll">
              <span class="flex items-center">
                <span class="bg-[#1768EF] text-[white] rounded-sm h-[14px] leading-[10px] p-[2px] mr-[5px]">ALL</span>
                <span>全部项目</span>
              </span>
              <i v-show="isSelectAll" class="bk-option-icon bk-icon icon-check-1 text-[2em] mr-[5px]"></i>
            </div>
            <bcs-divider class="!my-[2px] !border-b-[#c4c6cc]"></bcs-divider> -->
            <bcs-input
              clearable
              behavior="simplicity"
              left-icon="bk-icon icon-search"
              v-model="searchKey">
            </bcs-input>
          </template>
          <bk-option
            v-for="option in projectList"
            :key="option.projectID"
            :id="option.projectID"
            :name="option.name"
            :disabled="!(perms[option.projectID] && perms[option.projectID].project_view)"
            v-authority="{
              clickable: perms[option.projectID]
                && perms[option.projectID].project_view,
              actionId: 'project_view',
              resourceName: option.name,
              disablePerms: true,
              permCtx: {
                project_id: option.projectID
              }
            }">
            <span class="flex items-center justify-between">
              <span class="flex items-center max-w-[90%]">
                <span
                  class="bcs-ellipsis"
                  v-bk-overflow-tips="option.name">{{ option.name }}</span>
                <bcs-tag
                  class="flex-shrink-0 px-[5px]"
                  theme="info"
                  v-if="curProject.projectID === option.projectID">{{ $t('cluster.tag.currentProject') }}</bcs-tag>
              </span>
              <i
                v-show="innerValue.includes(option.projectID)"
                class="bk-option-icon bk-icon icon-check-1 text-[2em]"></i>
            </span>
          </bk-option>
        </bcs-select>
        <span
          class="text-[12px] text-[#3a84ff] ml-[8px] cursor-pointer"
          text
          @click="handleSave">{{ $t('generic.button.save') }}</span>
        <span
          class="text-[12px] text-[#3a84ff] ml-[8px] cursor-pointer"
          text
          @click="handleCancel">{{ $t('generic.button.cancel') }}</span>
      </div>
    </template>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onBeforeMount, PropType, ref, toRefs, watch } from 'vue';

import { IProject } from '@/composables/use-app';
import useDebouncedRef from '@/composables/use-debounce';
import clickoutside from '@/directives/clickoutside';
import $store from '@/store';
import useProjects, { IProjectPerm }  from '@/views/project-manage/project/use-project';

export default defineComponent({
  name: 'EditFormItem',
  directives: {
    clickoutside,
  },
  props: {
    value: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    type: {
      type: String,
      default: 'text',
    },
    placeholder: {
      type: String,
      default: '',
    },
    maxlength: Number,
    editable: {
      type: Boolean,
      default: false,
    },
    disableEdit: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { value } = toRefs(props);
    const { getProjectList } = useProjects();
    const curProject = computed(() => $store.state.curProject);

    const inputRef = ref<any>(null);
    const handleEdit = () => {
      ctx.emit('edit');
      setTimeout(() => {
        inputRef.value?.focus();
      });
    };
    const innerValue = ref<string[]>(value.value);
    const originValue = ref<string[]>(value.value);
    const isOnlyCurrentPorject = computed(() => (!isSelectAll.value && !innerValue.value.length)
      || (innerValue.value.length === 1 && innerValue.value[0] === curProject.value.projectID));
    const handleChange = async (params?) => {
      if (params === true) return;
      ctx.emit('cancel');
      // 值未变更不做保存
      if (innerValue.value === value.value) return;

      ctx.emit('change', innerValue.value);
    };
    function handleSave() {
      if (isSelectAll.value) {}
      ctx.emit('save', {
        isAll: isSelectAll.value,
        isOnlyCurrentPorject: isOnlyCurrentPorject.value,
        value: innerValue.value,
      });
    }
    function handleCancel() {
      ctx.emit('cancel');
      innerValue.value = originValue.value;
    }

    // 初始化数据
    const projectList = ref<IProject[]>([]);
    const projectNameList = computed(() => innerValue.value.map(item => projectIDMap.value[item]?.name || item));
    const loading = ref(false);
    const params = ref({
      offset: 0,
      limit: 20,
    });
    const perms = ref<Record<string, IProjectPerm>>({});
    async function handleInitProjectList() {
      params.value.offset = 0;
      const res = await getProjectList({
        ...params.value,
        searchKey: searchKey.value,
      }).catch(() => ({
        data: {
          results: [],
          total: 0,
        },
        web_annotations: {
          perms: {},
        },
      }));
      projectList.value = res?.data?.results || [];
      perms.value = res?.web_annotations?.perms || {};
      if (!value.value.length) {
        innerValue.value = [curProject.value.projectID];
        originValue.value = [curProject.value.projectID];
      }
    };

    // 远程搜索
    const selectRef = ref();
    const searchKey = useDebouncedRef('', 600);
    watch(searchKey, async () => {
      selectRef.value && (selectRef.value.searchLoading = true);
      await handleInitProjectList();
      selectRef.value && (selectRef.value.searchLoading = false);
    });

    // 滚动加载
    const projectIDMap = computed(() => projectList.value.reduce((pre, item) => {
      pre[item.projectID] = item;
      return pre;
    }, {}));
    const finished = ref(false);
    const scrollLoading = ref(false);
    const handleScrollToBottom = async () => {
      if (finished.value || scrollLoading.value) return;

      scrollLoading.value = true;
      params.value.offset = projectList.value.length;
      const { data, web_annotations } = await getProjectList({
        ...params.value,
        searchKey: searchKey.value,
      });
      // 过滤重复数据
      const filterData = data.results.filter(item => !projectIDMap.value[item.projectID]);
      if (!filterData.length) {
        finished.value = true;
      } else {
        projectList.value.push(...filterData);
        perms.value = Object.assign(perms.value, web_annotations.perms);
      }
      scrollLoading.value = false;
    };

    // 选择项目
    function handleProjectChange(projects) {
      isSelectAll.value = false;
      innerValue.value = [...projects];
      ctx.emit('change', [...projects]);
    }

    // 选择全部
    const isSelectAll = ref(false);
    function handleSelectAll() {
      isSelectAll.value = !isSelectAll.value;
      innerValue.value = [];
      ctx.emit('change', []);
    }

    watch(value, () => {
      if (!value.value.length) {
        innerValue.value = [curProject.value.projectID];
        originValue.value = [curProject.value.projectID];
        return;
      }
      innerValue.value = [...value.value];
      originValue.value = [...value.value];
    });

    onBeforeMount(async () => {
      loading.value = true;
      await handleInitProjectList();
      loading.value = false;
    });

    return {
      innerValue,
      inputRef,
      projectList,
      loading,
      scrollLoading,
      selectRef,
      curProject,
      searchKey,
      projectNameList,
      isSelectAll,
      isOnlyCurrentPorject,
      perms,
      handleSave,
      handleEdit,
      handleChange,
      handleScrollToBottom,
      handleProjectChange,
      handleSelectAll,
      handleCancel,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> textarea::-webkit-scrollbar {
  width: 4px;
}

>>> textarea::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 20px;
}

.clamp-text {
  display: -webkit-box; /* 使用flexbox */
  -webkit-box-orient: vertical; /* 设置flex方向为垂直 */
  -webkit-line-clamp: 4; /* 限制文本行数 */
  overflow: hidden;
  text-overflow: ellipsis; /* 添加省略号 */
}
</style>
