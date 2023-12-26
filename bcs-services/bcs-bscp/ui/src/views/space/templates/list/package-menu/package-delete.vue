<template>
  <DeleteConfirmDialog
    :is-show="isShow"
    title="确认删除该配置模板套餐？"
    @confirm="handleDelete"
    @close="emits('update:show',false)"
  >
    <div style="margin-bottom: 8px;">
      配置模板套餐: <span style="color: #313238;font-weight: 600;">{{ props.pkg.spec.name }}</span>
    </div>
    <div style="margin-bottom: 8px;">一旦删除，该操作将无法撤销，以下服务配置的未命名版本中引用该套餐的内容也将清除</div>
    <div class="service-table">
      <bk-loading style="min-height: 200px" :loading="appsLoading">
        <bk-table :data="appList" :max-height="maxTableHeight" empty-text="暂无未命名版本引用此套餐">
          <bk-table-column label="引用此套餐的服务">
            <template #default="{ row }">
              <div class="app-info" @click="goToConfigPageImport(row.app_id)">
                <div v-overflow-title class="name-text" >{{ row.app_name }}</div>
                <LinkToApp class="link-icon" :id="row.app_id" :auto-jump="true" />
              </div>
            </template>
          </bk-table-column>
        </bk-table>
      </bk-loading>
    </div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import Message from 'bkui-vue/lib/message';
import useGlobalStore from '../../../../../store/global';
import { getUnNamedVersionAppsBoundByPackage, deleteTemplatePackage } from '../../../../../api/template';
import { ITemplatePackageItem, IPackageCitedByApps } from '../../../../../../types/template';
import useTemplateStore from '../../../../../store/template';
import LinkToApp from '../components/link-to-app.vue';
import DeleteConfirmDialog from '../../../../../components/delete-confirm-dialog.vue';

const { spaceId } = storeToRefs(useGlobalStore());
const { currentTemplateSpace } = storeToRefs(useTemplateStore());

const props = defineProps<{
  show: boolean;
  templateSpaceId: number;
  pkg: ITemplatePackageItem;
}>();

const emits = defineEmits(['update:show', 'deleted']);

const router = useRouter();

const isShow = ref(false);
const appsLoading = ref(false);
const appList = ref<IPackageCitedByApps[]>([]);
const pending = ref(false);

const maxTableHeight = computed(() => {
  const windowHeight = window.innerHeight;
  return windowHeight * 0.6 - 200;
});

watch(
  () => props.show,
  (val) => {
    isShow.value = val;
    if (val) {
      getRelatedApps();
    }
  },
);

const goToConfigPageImport = (id: number) => {
  const { href } = router.resolve({
    name: 'service-config',
    params: { appId: id },
    query: { pkg_id: currentTemplateSpace.value },
  });
  window.open(href, '_blank');
};

const getRelatedApps = async () => {
  appsLoading.value = true;
  const params = {
    start: 0,
    all: true,
  };
  const res = await getUnNamedVersionAppsBoundByPackage(spaceId.value, props.templateSpaceId, props.pkg.id, params);
  appList.value = res.details;
  appsLoading.value = false;
};

const handleDelete = async () => {
  pending.value = true;
  await deleteTemplatePackage(spaceId.value, props.templateSpaceId, props.pkg.id);
  close();
  emits('deleted', props.pkg.id);
  Message({
    theme: 'success',
    message: '删除配置模板套餐成功',
  });
  pending.value = false;
};

const close = () => {
  emits('update:show', false);
};
</script>
<style lang="postcss" scoped>
.tips {
  margin: 0 0 16px;
  font-size: 14px;
  line-height: 22px;
  color: #63656e;
  text-align: center;
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
.action-btns {
  padding-top: 32px;
  text-align: center;
  /* border-top: 1px solid #dcdee5; */
  .delete-btn {
    margin-right: 8px;
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
