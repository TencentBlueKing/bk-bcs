<template>
  <section class="scripts-manage-page">
    <div class="side-menu">
      <div class="group-wrapper">
        <li :class="['group-item', { actived: showAllTag }]" @click="handleSelectTag('', true)">
          <i class="bk-bscp-icon icon-block-shape group-icon"></i>全部脚本
        </li>
        <li :class="['group-item', { actived: !showAllTag && selectedTag === '' }]" @click="handleSelectTag('')">
          <i class="bk-bscp-icon icon-tags group-icon"></i>未分类
        </li>
      </div>
      <div class="custom-tag-list">
        <li
          v-for="(item, index) in tagsData"
          :key="index"
          :class="['group-item', { actived: selectedTag === item.tag }]"
          @click="handleSelectTag(item.tag)"
        >
          <i class="bk-bscp-icon icon-tags group-icon"></i>
          <span class="name">{{ item.tag }}</span>
          <span class="num">{{ item.counts }}</span>
        </li>
      </div>
    </div>
    <div class="script-list-wrapper">
      <div class="operate-area">
        <bk-button theme="primary" @click="showCreateScript = true"><Plus class="button-icon" />新建脚本</bk-button>
        <bk-input
          v-model="searchStr"
          class="search-script-input"
          placeholder="脚本名称"
          :clearable="true"
          @clear="refreshList"
          @input="handleNameInputChange"
        >
          <template #suffix>
            <Search class="search-input-icon" />
          </template>
        </bk-input>
      </div>
      <bk-loading style="min-height: 300px" :loading="scriptsLoading">
        <bk-table
          :border="['outer']"
          :data="scriptsData"
          :remote-pagination="true"
          :pagination="pagination"
          @page-limit-change="handlePageLimitChange"
          @page-value-change="handlePageCurrentChange"
        >
          <bk-table-column label="脚本名称">
            <template #default="{ row }">
              <div
                v-if="row.hook"
                class="hook-name"
                @click="handleViewVersionClick(row.hook.id)"
              >
                {{ row.hook.spec.name }}
              </div>
            </template>
          </bk-table-column>
          <bk-table-column label="脚本语言" prop="hook.spec.type" width="120"></bk-table-column>
          <bk-table-column label="分类标签">
            <template #default="{ row }">
              <span v-if="row.hook">{{ row.hook.spec.tag || '--' }}</span>
            </template>
          </bk-table-column>
          <bk-table-column label="被引用" width="100">
            <template #default="{ row }">
              <bk-button v-if="row.bound_num > 0" text theme="primary" @click="handleOpenCitedSlider(row.hook.id)">{{
                row.bound_num
              }}</bk-button>
              <span v-else>0</span>
            </template>
          </bk-table-column>
          <bk-table-column label="更新人" prop="hook.revision.reviser"></bk-table-column>
          <bk-table-column label="更新时间" width="220">
            <template #default="{ row }">
              <span v-if="row.hook">{{ datetimeFormat(row.hook.revision.update_at) }}</span>
            </template>
          </bk-table-column>
          <bk-table-column label="操作">
            <template #default="{ row }" width="180">
              <div class="action-btns">
                <bk-button text theme="primary" @click="handleEditClick(row)">编辑</bk-button>
                <bk-button
                  text
                  theme="primary"
                  @click="router.push({ name: 'script-version-manage', params: { spaceId, scriptId: row.hook.id } })"
                  >版本管理</bk-button
                >
                <bk-button text theme="primary" @click="handleDeleteScript(row)">删除</bk-button>
              </div>
            </template>
          </bk-table-column>
          <template #empty>
            <TableEmpty :is-search-empty="isSearchEmpty" @clear="clearSearchStr"></TableEmpty>
          </template>
        </bk-table>
      </bk-loading>
    </div>
    <CreateScript v-if="showCreateScript" v-model:show="showCreateScript" @created="handleCreatedScript" />
    <ScriptCited v-model:show="showCiteSlider" :id="currentId" />
  </section>
  <DeleteConfirmDialog
    v-model:isShow="isDeleteScriptDialogShow"
    title="确认删除该脚本？"
    @confirm="handleDeleteScriptConfirm"
  >
    <div style="margin-bottom: 8px">
      脚本: <span style="color: #313238; font-weight: 600">{{ deleteScriptItem?.hook.spec.name }}</span>
    </div>
    <div style="margin-bottom: 8px">一旦删除，该操作将无法撤销，以下服务配置的未命名版本中引用该脚本也将清除</div>
    <div class="service-table">
      <bk-loading style="min-height: 200px" :loading="appsLoading">
        <bk-table :data="appList" :max-height="maxTableHeight" empty-text="暂无未命名版本引用此脚本">
          <bk-table-column label="引用此脚本的服务">
            <template #default="{ row }">
              <div class="app-info" @click="goToConfigPageImport(row.app_id)">
                <div v-overflow-title class="name-text">{{ row.app_name }}</div>
                <LinkToApp class="link-icon" :id="row.app_id" :auto-jump="true" />
              </div>
            </template>
          </bk-table-column>
        </bk-table>
      </bk-loading>
    </div>
  </DeleteConfirmDialog>
</template>
<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { Message } from 'bkui-vue';
import { Plus, Search } from 'bkui-vue/lib/icon';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../../../store/global';
import useScriptStore from '../../../../store/script';
import { getScriptList, getScriptTagList, deleteScript, getScriptCiteList } from '../../../../api/script';
import { IScriptItem, IScriptTagItem, IScriptListQuery } from '../../../../../types/script';
import { datetimeFormat } from '../../../../utils/index';
import CreateScript from './create-script.vue';
import ScriptCited from './script-cited.vue';
import TableEmpty from '../../../../components/table/table-empty.vue';
import DeleteConfirmDialog from '../../../../components/delete-confirm-dialog.vue';
import LinkToApp from '../../templates/list/components/link-to-app.vue';
import { debounce } from 'lodash';

const { spaceId } = storeToRefs(useGlobalStore());
const { versionListPageShouldOpenEdit, versionListPageShouldOpenView } = storeToRefs(useScriptStore());
const router = useRouter();

interface IAppItem {
  app_id: number;
  app_name: string;
}
const showCreateScript = ref(false);
const showCiteSlider = ref(false);
const scriptsData = ref<IScriptItem[]>([]);
const scriptsLoading = ref(false);
const tagsData = ref<IScriptTagItem[]>([]);
const tagsLoading = ref(false);
const showAllTag = ref(true); // 全部脚本
const selectedTag = ref(''); // 未分类或具体tag下脚本
const currentId = ref(0);
const searchStr = ref('');
const isDeleteScriptDialogShow = ref(false);
const deleteScriptItem = ref<IScriptItem>();
const appsLoading = ref(false);
const appList = ref<IAppItem[]>([]);

const maxTableHeight = computed(() => {
  const windowHeight = window.innerHeight;
  return windowHeight * 0.6 - 200;
});
const pagination = ref({
  current: 1,
  count: 0,
  limit: 10,
});
const isSearchEmpty = ref(false);
watch(
  () => spaceId.value,
  () => {
    refreshList();
    getTags();
  },
);

onMounted(() => {
  getScripts();
  getTags();
});

// 获取脚本列表
const getScripts = async () => {
  scriptsLoading.value = true;
  const params: IScriptListQuery = {
    start: (pagination.value.current - 1) * pagination.value.limit,
    limit: pagination.value.limit,
  };
  if (selectedTag.value === '' && !showAllTag.value) {
    params.not_tag = true;
  } else {
    params.tag = selectedTag.value;
  }
  if (searchStr.value) {
    params.name = searchStr.value;
  }
  const res = await getScriptList(spaceId.value, params);
  scriptsData.value = res.details;
  scriptsData.value.forEach((item) => {
    item.hook.spec.type = item.hook.spec.type.charAt(0).toUpperCase() + item.hook.spec.type.slice(1);
  });
  pagination.value.count = res.count;
  scriptsLoading.value = false;
};

// 获取标签列表
const getTags = async () => {
  tagsLoading.value = true;
  const res = await getScriptTagList(spaceId.value);
  tagsData.value = res.details;
  tagsLoading.value = false;
};

const handleSelectTag = (tag: string, all = false) => {
  searchStr.value = '';
  selectedTag.value = tag;
  showAllTag.value = all;
  refreshList();
};

const handleOpenCitedSlider = (id: number) => {
  currentId.value = id;
  showCiteSlider.value = true;
};

const handleEditClick = (script: IScriptItem) => {
  router.push({ name: 'script-version-manage', params: { spaceId: spaceId.value, scriptId: script.hook.id } });
  versionListPageShouldOpenEdit.value = true;
};

const handleViewVersionClick = (id: number) => {
  router.push({ name: 'script-version-manage', params: { spaceId: spaceId.value, scriptId: id } });
  versionListPageShouldOpenView.value = true;
};

// 删除分组
const handleDeleteScript = async (script: IScriptItem) => {
  const params = {
    start: (pagination.value.current - 1) * pagination.value.limit,
    limit: pagination.value.limit,
  };
  const res = await getScriptCiteList(spaceId.value, script.hook.id, params);
  const allAppInfo = res.details.map((item: any) => {
    const { app_id, app_name } = item;
    return {
      app_id,
      app_name,
    };
  });
  appList.value = Array.from(allAppInfo.reduce((map: Map<number, IAppItem>, obj: IAppItem) => map.set(obj.app_id, obj), new Map()).values());
  deleteScriptItem.value = script;
  isDeleteScriptDialogShow.value = true;
};

const handleDeleteScriptConfirm = async () => {
  await deleteScript(spaceId.value, deleteScriptItem.value!.hook.id);
  if (scriptsData.value.length === 1 && pagination.value.current > 1) {
    pagination.value.current = pagination.value.current - 1;
  }
  Message({
    theme: 'success',
    message: '删除版本成功',
  });
  isDeleteScriptDialogShow.value = false;
  getScripts();
};

// const getRelatedApps = async () => {
//   appsLoading.value = true;
//   const params = {
//     start: 0,
//     all: true,
//   };
//   const res = await getUnNamedVersionAppsBoundByPackage(spaceId.value, props.templateSpaceId, props.pkg.id, params);
//   appList.value = res.details;
//   appsLoading.value = false;
// };

const goToConfigPageImport = (id: number) => {
  const { href } = router.resolve({
    name: 'service-config',
    params: { appId: id },
  });
  window.open(href, '_blank');
};

const handleNameInputChange = debounce(() => refreshList(), 300);

const handleCreatedScript = () => {
  refreshList();
  getTags();
};

const refreshList = () => {
  searchStr.value ? (isSearchEmpty.value = true) : (isSearchEmpty.value = false);
  pagination.value.current = 1;
  getScripts();
};

const handlePageLimitChange = (val: number) => {
  pagination.value.limit = val;
  refreshList();
};

const handlePageCurrentChange = (val: number) => {
  pagination.value.current = val;
  getScripts();
};

const clearSearchStr = () => {
  searchStr.value = '';
  refreshList();
};
</script>
<style lang="scss" scoped>
.scripts-manage-page {
  display: flex;
  align-items: center;
  height: 100%;
  background: #ffffff;
}
.side-menu {
  padding: 16px 0;
  width: 280px;
  height: 100%;
  background: #f5f7fa;
  box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.15);
  z-index: 1;
  .group-wrapper {
    padding-bottom: 16px;
    border-bottom: 1px solid #dcdee5;
  }
  .group-item {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 16px 8px 44px;
    color: #313238;
    font-size: 12px;
    cursor: pointer;
    &:hover {
      color: #348aff;
      .group-icon {
        color: #3a84ff;
      }
    }
    &.actived {
      background: #e1ecff;
      color: #3a84ff;
      .group-icon {
        color: #3a84ff;
      }
      .num {
        color: #ffffff;
        background: #a3c5fd;
      }
    }
    .group-icon {
      position: absolute;
      top: 9px;
      left: 22px;
      margin-right: 8px;
      font-size: 14px;
      color: #979ba5;
    }
    .name {
      flex: 0 1 auto;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .num {
      flex: 0 0 auto;
      padding: 0 8px;
      height: 16px;
      line-height: 16px;
      color: #979ba5;
      background: #f0f1f5;
      border-radius: 2px;
    }
  }
  .custom-tag-list {
    padding: 16px 0;
    height: calc(100% - 82px);
    overflow: auto;
  }
}
.script-list-wrapper {
  padding: 24px;
  width: calc(100% - 280px);
  height: 100%;
  background: #ffffff;
  overflow: auto;
}
.operate-area {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  .button-icon {
    font-size: 18px;
  }
}
.search-script-input {
  width: 320px;
}
.search-input-icon {
  padding-right: 10px;
  color: #979ba5;
  background: #ffffff;
}
.action-btns {
  .bk-button {
    margin-right: 8px;
  }
}
.hook-name {
  color: #348aff;
  cursor: pointer;
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

<style lang="scss">
.service-table {
  thead th[colspan] {
    background-color: #f0f1f5 !important;
  }
}
</style>
