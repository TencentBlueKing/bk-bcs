<script setup lang="ts">
import { computed, onMounted, ref, Ref, watch, defineProps } from "vue";
import { useRouter } from 'vue-router'
import { Plus, Del } from "bkui-vue/lib/icon";
import InfoBox from "bkui-vue/lib/info-box";
import { useI18n } from "vue-i18n";
import { deleteApp, getAppList, getBizList, createApp, updateApp, IAppListQuery } from "../../api";

type IServingItem = {
  id?: number,
  biz_id: number,
  spec: {
    name: string,
    deploy_type: string,
    config_type: string,
    mode: string,
    memo: string,
    reload: {
      file_reload_spec: {
        reload_file_path: string
      },
      reload_type: string
    }
  },
  revision: {
    creator: string,
    reviser: string,
    create_at: string,
    update_at: string,
  }
}

const router = useRouter()
const { t } = useI18n();
const props = defineProps<{
  type: string
}>()

const bizList = ref();
const pagination = ref({
  current: 1,
  limit: 50,
  count: 0,
});
const servingList = ref([]) as Ref<IServingItem[]>

const isEmpty = computed(() => {
  return servingList.value.length === 0;
});

const isCreateAppShow = ref(false);
const isAttrShow = ref(false);
const appName = ref("");
const isLoading = ref(false);
const isBizLoading = ref(false);
const createAppPending = ref(false);
const bkBizId = ref(2); // 目前缺少拉取业务列表的接口，固定写死一个业务ID调试
const formRef = ref();
const formData = ref({
  biz_id: bkBizId.value,
  name: "",
  config_type: "file",
  reload_type: "file",
  reload_file_path: "/bscp_test",
  mode: "normal",
  deploy_type: "common",
  memo: "",
});

const activeAppItem = ref({
  id: 0,
  biz_id: 0,
  spec: {
    name: "",
    deploy_type: "",
    config_type: "",
    mode: "",
    memo: "",
    reload: {
      file_reload_spec: {
        reload_file_path: ""
      },
      reload_type: ""
    }
  },
  revision: {
    creator: "",
    reviser: "",
    create_at: "",
    update_at: "",
  },
});

const isAttrMemoEdit = ref(false);

// 查询条件
const filterKeyword = computed(() => {
  const { current, limit } = pagination.value;

  const rules: IAppListQuery = {
    start: (current - 1) * limit,
    limit: limit,
  };
  if (appName.value) {
    rules.name = appName.value
  }
  if (props.type === 'mine') {
    rules.operator = ''
  }
  return rules
});

const handleCreateAppClick = () => {
  isCreateAppShow.value = true;
};

const handleDeleteItem = (item: any) => {
  InfoBox({
    title: `确认是否删除服务 ${item.spec.name}?`,
    type: "danger",
    headerAlign: "center" as const,
    footerAlign: "center" as const,
    onConfirm: () => {
      const { id, biz_id } = item;
      deleteApp(id, biz_id).then((resp) =>
        resp.validate().then(() => {
          loadServingList();
        })
      );
    },
  } as any);
};

const handleConfigTypeClick = (type: string) => {
  formData.value.deploy_type = "";
  formData.value.config_type = type;
};

const handleCreateAppForm = () => {
  formRef.value.validate().then(() => {
    createAppPending.value = true;
    createApp(formData.value.biz_id, formData.value).then((resp) =>
      resp.validate(true).then(() => {
        isCreateAppShow.value = false;
        createAppPending.value = false;
        InfoBox({
          type: "success",
          title: "服务新建成功",
          subTitle: "接下来你可以在服务下新增并使用配置项",
          headerAlign: "center",
          footerAlign: "center",
          confirmText: "新增配置项",
          cancelText: "稍后再说",
          onConfirm() {
            router.push({
              name: 'serving-config',
              params: {
                id: resp.response.data.id
              }
            })
          },
          onClose() {
            loadServingList()
          }
        } as any);
      })
    );
  });
};

const handlePaginationChange = () => {
  loadServingList();
};

const handleLimitChange = (limit: number) => {
  pagination.value.limit = limit;
  loadServingList();
};

const handleItemAttributeClick = (item: any) => {
  activeAppItem.value = item;
  isAttrShow.value = true;
  isAttrMemoEdit.value = false;
};

const handleEditAttrMemo = () => {
  isAttrMemoEdit.value = true;
};

const handleItemMemoBlur = () => {
  const { id, biz_id, spec } = activeAppItem.value;
  const { name, mode, memo, config_type, reload } = spec;
  const data = {
    id,
    biz_id,
    name,
    mode,
    memo,
    config_type,
    reload_type: reload.reload_type,
    reload_file_path: reload.file_reload_spec.reload_file_path,
    deploy_type: "common",
  }
  updateApp({ id, biz_id, data }).then(resp => resp.validate(true));
}

const handleNameInputChange = (val: string) => {
  if (!val) {
    handleSearch()
  }
}

const handleSearch = () => {
  pagination.value.current = 1
  loadServingList()
}

const loadServingList = () => {
  isLoading.value = true;
  getAppList(Number(bkBizId.value), filterKeyword.value)
    .then((resp) => {
      resp.validate().then((data: any) => {
        servingList.value = data?.details;
        pagination.value.count = data?.count;
      });
    })
    .finally(() => {
      isLoading.value = false;
    });
};

onMounted(() => {
  loadServingList()
  // isBizLoading.value = true;
  // getBizList()
  //   .then((res) => {
  //     res.validate().then((data: any) => {
  //       bizList.value = data?.info || [];
  //       bkBizId.value = bizList.value[0]?.bk_biz_id;
  //     });
  //   })
  //   .finally(() => {
  //     isBizLoading.value = false;
  //   });
});

watch(
  () => bkBizId.value,
  (value) => {
    loadServingList();
  }
);
</script>

<template>
  <bk-loading :loading="isLoading" class="serving-content">
    <div class="head-section">
      <bk-button theme="primary" @click="handleCreateAppClick">
        <Plus />
        {{ t("新建服务") }}
      </bk-button>
      <div class="head-right">
        <bk-select
          v-model="bkBizId"
          class="bk-select"
          :list="bizList"
          :loading="isBizLoading"
          id-key="bk_biz_id"
          display-key="bk_biz_name"
          filterable>
        </bk-select>
        <bk-input
          class="search-app-name"
          type="search"
          v-model="appName"
          :placeholder="t('服务名称')"
          :clearable="true"
          @change="handleNameInputChange"
          @enter="handleSearch"
          @clear="handleSearch">
        </bk-input>
      </div>
    </div>
    <div class="content-body">
      <template v-if="isEmpty">
        <bk-exception
          class="exception-wrap-item"
          type="empty"
          :description="t('你尚未创建或加入任何服务')"
        >
          <div class="exception-actions">
            <bk-button text theme="primary" @click="handleCreateAppClick">{{
              t("立即创建")
            }}</bk-button>
            <span class="divider-middle"></span>
            <bk-button text theme="primary">{{ t("申请权限") }}</bk-button>
          </div>
        </bk-exception>
      </template>
      <template v-else>
        <div class="serving-list">
          <div v-for="item in servingList" :key="item.id" class="serving-item">
            <div class="serving-item-body">
              <div class="item-head">{{ item.spec?.name }}</div>
              <div class="item-tag">
                <bk-tag>MagicBox</bk-tag>
                <Del
                  fill="#979BA5"
                  class="item-tag-del"
                  @click="() => handleDeleteItem(item)"
                />
              </div>
              <div class="item-config">
                <div class="config-info">
                  <span class="bk-bscp-icon icon-configuration-line"></span>
                  xx个配置项
                </div>
                <div class="time-info">
                  <span class="bk-bscp-icon icon-time-2"></span>
                  {{ item.revision?.update_at }}
                </div>
              </div>
              <div class="item-footer">
                <bk-button
                  size="small"
                  @click="() => handleItemAttributeClick(item)"
                  style="width: 50%"
                  text>
                  {{ t("服务属性") }}
                </bk-button>
                <span class="divider-middle"></span>
                <bk-button size="small" style="width: 50%" text @click="router.push({ name: 'serving-config', params: { id: item.id } })">
                  {{t("配置管理")}}
                </bk-button>
              </div>
            </div>
          </div>
        </div>
        <bk-pagination
          v-model="pagination.current"
          :count="pagination.count"
          :limit="pagination.limit"
          @change="handlePaginationChange"
          @limit-change="handleLimitChange"
        />
      </template>
    </div>
    <bk-sideslider
      v-model:isShow="isCreateAppShow"
      :title="t('新建服务')"
      width="640"
    >
      <div class="create-app-form">
        <bk-form form-type="vertical" :model="formData" ref="formRef">
          <bk-form-item :label="t('所属业务')" property="biz_id" required>
            <bk-select
              v-model="formData.biz_id"
              class="bk-select"
              :list="bizList"
              :loading="isBizLoading"
              id-key="bk_biz_id"
              display-key="bk_biz_name"
              filterable
            ></bk-select>
          </bk-form-item>
          <bk-form-item :label="t('服务名称')" property="name" required>
            <bk-input
              placeholder="需以英文、数字和下划线组成，不超过 50 字符"
              v-model="formData.name"
            ></bk-input>
          </bk-form-item>
          <bk-form-item :label="t('服务描述')">
            <bk-input
              placeholder="请输入"
              type="textarea"
              v-model="formData.memo"
            />
          </bk-form-item>
          <!-- <bk-form-item property="config_type" required>
            <div class="config-type">
              你想以哪种方式来为您的服务接入配置管理
            </div>
            <div class="config-type-items">
              <div
                class="config-type-item"
                @click="() => handleConfigTypeClick('runtime')"
                :class="{ active: formData.config_type === 'runtime' }"
              >
                <div class="type-item-title">运行时配置</div>
                <div class="type-item-des">
                  <div>· 需在开发逻辑中集成配置使用</div>
                  <div>· 适用于「自研业务」</div>
                </div>
              </div>
              <div
                class="config-type-item"
                @click="() => handleConfigTypeClick('file')"
                :class="{ active: formData.config_type === 'file' }"
              >
                <div class="type-item-title">配置文件</div>
                <div class="type-item-des">
                  <div>· 以配置文件方式落地到客户端</div>
                  <div>· 适用于「三方交付业务」</div>
                </div>
                <div class="type-item-des">
                  <bk-radio-group v-model="formData.deploy_type">
                    <div class="type-item-radio">
                      <bk-radio value="LF" label="LF" /><span
                        >Unix and macOS (\n)</span
                      >
                    </div>
                    <div class="type-item-radio">
                      <bk-radio value="CRLF" label="CRLF" /><span
                        >Windows (\r\n)</span
                      >
                    </div>
                  </bk-radio-group>
                </div>
              </div>
            </div>
          </bk-form-item> -->
        </bk-form>
      </div>
      <template #footer>
        <div class="create-app-footer">
          <bk-button
            style="width: 88px"
            theme="primary"
            :loading="createAppPending"
            @click="handleCreateAppForm"
            >{{ t("提交") }}</bk-button
          >
          <bk-button style="width: 88px">{{ t("取消") }}</bk-button>
        </div>
      </template>
    </bk-sideslider>
    <bk-sideslider
      v-model:isShow="isAttrShow"
      :title="t('服务属性')"
      width="400"
      quick-close
    >
      <template #header>
        <div class="serving-attribute-head">
          <span class="title">{{ t("服务属性") }}</span>
          <span class="secret-key"
            ><a href="" target="_blank">{{ t("密钥管理") }}</a></span
          >
        </div>
      </template>
      <div class="create-app-form attributes">
        <bk-form :model="activeAppItem" label-width="100">
          <bk-form-item :label="t('服务名称')">{{
            activeAppItem.spec.name
          }}</bk-form-item>
          <bk-form-item :label="t('所属业务')">{{
            activeAppItem.biz_id
          }}</bk-form-item>
          <bk-form-item :label="t('服务描述')">
            <div class="content-edit">
              <template v-if="isAttrMemoEdit">
                <bk-input type="textarea" v-model="activeAppItem.spec.memo" :show-word-limit="true" :maxlength="255" @blur="handleItemMemoBlur"></bk-input>
              </template>
              <template v-else>
                {{ activeAppItem.spec.memo }}
                <span
                @click="handleEditAttrMemo"
                class="bk-bscp-icon icon-edit-small"
              ></span>
              </template>
            </div>
          </bk-form-item>
          <bk-form-item :label="t('接入方式')"
            >{{ activeAppItem.spec.config_type }}-{{
              activeAppItem.spec.deploy_type
            }}</bk-form-item
          >
          <bk-form-item :label="t('创建者')">{{
            activeAppItem.revision.creator
          }}</bk-form-item>
          <bk-form-item :label="t('创建时间')">{{
            activeAppItem.revision.create_at
          }}</bk-form-item>
        </bk-form>
      </div>
    </bk-sideslider>
  </bk-loading>
</template>


<style lang="scss" scoped>
.serving-content {
  overflow-x: hidden;
  .head-section {
    padding: 16px 80px;
    display: flex;
    justify-content: space-between;

    .head-right {
      display: flex;

      .search-app-name {
        margin-left: 16px;
        width: 240px;
      }
    }
  }

  .content-body {
    padding: 0 72px;

    .serving-list {
      display: flex;
      flex-wrap: wrap;
      align-content: flex-start;
      // height: calc(100vh - 210px);
      // overflow: auto;

      :deep(.bk-exception-description) {
        margin-top: 5px;
        font-size: 12px;
        color: #979ba5;
      }

      :deep(.bk-exception-footer) {
        margin-top: 5px;
      }

      .exception-actions {
        display: flex;
        font-size: 12px;
        color: #3a84ff;
        .divider-middle {
          display: inline-block;
          margin: 0 16px;
          width: 1px;
          height: 16px;
          background: #dcdee5;
        }
      }

      .serving-item {
        width: 20%;
        height: 165px;
        padding: 0px 8px 16px 8px;

        .serving-item-body {
          background: #ffffff;
          border: 1px solid #dcdee5;
          border-radius: 2px;
          height: 100%;
          text-align: left;

          &:hover {
            .item-tag {
              .item-tag-del {
                display: block;
              }
            }
          }
          .item-head {
            margin-top: 16px;
            position: relative;
            height: 22px;
            font-weight: Bold;
            font-size: 14px;
            color: #313238;
            line-height: 22px;
            text-align: left;
            padding: 0 16px;
            display: flex;
            align-items: center;
            height: 22px;

            &::before {
              content: "";
              position: absolute;
              left: 0;
              top: 3px;
              width: 4px;
              height: 16px;
              background: #699df4;
              border-radius: 0 2px 2px 0;
            }
          }

          .item-tag {
            width: 100%;
            min-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            padding: 8px 16px;
            height: 32px;
            position: relative;
            .item-tag-del {
              position: absolute;
              right: 16px;
              top: 0;
              display: none;
              cursor: pointer;
            }
          }

          .item-config {
            padding: 0 16px;
            height: 20px;
            font-size: 12px;
            color: #979ba5;
            line-height: 20px;
            margin: 4px 0 12px 0;
            display: flex;
            align-items: center;

            .config-info {
              width: 80px;
            }

            .time-info {
              padding-left: 10px;
            }
          }

          .item-footer {
            height: 40px;
            border-top: solid 1px #f0f1f5;
            display: flex;
            justify-content: center;
            width: 100%;
            font-size: 12px;

            :deep(.bk-button) {
              &.is-text {
                color: #979ba5;
              }

              &:hover {
                color: #3a84ff;
              }
            }

            .divider-middle {
              display: inline-block;
              width: 1px;
              height: 100%;
              background: #f0f1f5;
              margin: 0 16px;
            }
          }
        }
      }
    }
  }
}

.create-app-form {
  padding: 20px 24px;
  height: calc(100vh - 108px);

  .config-type {
    width: 100%;
    font-size: 12px;
    color: #63656e;
    text-align: center;
    padding: 24px 0;
    border-top: solid 1px #dcdee5;
  }
  .config-type-items {
    display: flex;
    justify-content: space-between;
    .config-type-item {
      width: 284px;
      height: 147px;
      background: #f5f7fa;
      border-radius: 2px;
      border: 1px solid transparent;
      padding: 16px;
      cursor: pointer;

      &.active {
        background: #f5f7fa;
        border: 1px solid #699df4;
        border-radius: 2px;
        position: relative;

        &::before {
          content: "";
          color: #fff;
          border-top: 18px solid #3a84ff;
          border-right: 18px solid #3a84ff;
          border-bottom: 18px solid transparent;
          border-left: 18px solid transparent;
          position: absolute;
          top: 0;
          right: 0;
        }

        &::after {
          content: "";
          border-bottom: 1px solid #fff;
          border-left: 1px solid #fff;
          position: absolute;
          transform: rotate(-45deg);
          width: 8px;
          height: 4px;
          top: 8px;
          right: 6px;
        }
      }

      .type-item-title {
        font-weight: Bold;
        font-size: 14px;
        color: #63656e;
      }

      .type-item-des {
        font-size: 12px;
        color: #979ba5;
        margin: 8px 0;
        div {
          line-height: 16px;
          height: 16px;

          .type-item-radio {
            margin-bottom: 8px;
            display: flex;
            span {
              display: inline-block;
              margin-left: 8px;
              font-size: 12px;
              color: #979ba5;
            }
          }
        }
      }
    }
  }

  &.attributes {
    font-size: 12px;
    :deep(.bk-form-item) {
      margin-bottom: 16px;

      .bk-form-label,
      .bk-form-content {
        line-height: 16px;
        font-size: 12px;
      }
    }

    .content-edit {
      position: relative;
      span {
        display: none;
        position: absolute;
        font-size: 16px;
        color: #979ba5;
      }

      &:hover {
        span {
          cursor: pointer;
          display: inline-block;
          right: 0;
          top: 0;
        }
      }
    }
  }
}

.create-app-footer {
  padding: 8px 24px;
  height: 48px;
  width: 100%;
  background: #fafbfd;
  box-shadow: 0 -1px 0 0 #dcdee5;
  button {
    margin-right: 8px;
  }
}

.serving-attribute-head {
  display: flex;
  align-content: center;
  justify-content: space-between;
  padding-right: 24px;

  a {
    font-size: 12px;
  }
}
</style>