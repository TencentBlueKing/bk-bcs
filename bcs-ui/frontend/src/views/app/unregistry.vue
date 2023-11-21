<template>
  <!-- 容器服务未开通时引导界面 -->
  <section class="bcs-unregistry">
    <header class="header">{{ $t('bcs.registry.text') }}</header>
    <main class="main">
      <div class="form-item">
        <div class="form-item-label">{{ $t('bcs.registry.label.projectKind.text') }}</div>
        <div class="form-item-content kind">
          <div
            v-for="item in kindList"
            :class="['kind-panel', { active: kind === item.id, disabled: item.disabled }]"
            :key="item.id"
            v-bk-tooltips="{
              disabled: !item.disabled && !item.tips,
              content: item.tips
            }"
            @click="handleKindChange(item)">
            <div class="kind-panel-title">{{ item.name }}</div>
            <div class="kind-panel-desc mt5">{{ item.desc }}</div>
          </div>
        </div>
      </div>
      <div class="form-item mt30">
        <div class="form-item-label">{{ $t('bcs.registry.label.business.text') }}</div>
        <div class="form-item-content cc-list">
          <bcs-select
            class="cc-selector"
            :placeholder="$t('bcs.registry.label.business.placeholder')"
            v-model="ccKey"
            searchable>
            <bcs-option
              v-for="item in ccList"
              :key="item.businessID"
              :id="String(item.businessID)"
              :name="item.name">
            </bcs-option>
          </bcs-select>
        </div>
        <div class="form-item-tips" v-if="!ccList.length && $INTERNAL">
          <i18n path="bcs.registry.label.business.emptyMsg">
            <bk-link
              theme="primary"
              :href="PROJECT_CONFIG.teaApply"
              target="_blank">
              {{ $t('bcs.registry.button.viewAndApplyOperation') }}
            </bk-link>
          </i18n>
        </div>
      </div>
      <div class="form-item enable-bcs">
        <bk-button
          theme="primary"
          :disabled="!enableBtn"
          :loading="isLoading"
          @click="updateProject">{{ $t('bcs.registry.text') }}</bk-button>
      </div>
      <div class="form-item guide" v-if="$INTERNAL">
        <div
          v-for="(item, index) in guideList"
          :class="['guide-link', { mr18: index < (guideList.length - 1) }]"
          :key="item.id">
          <i :class="`bcs-icon bcs-icon-${item.id}`" :style="{ color: item.iconColor }"></i>
          <span class="desc mt20">{{ item.desc }}</span>
          <bk-link class="mt8" theme="primary" :href="item.link" target="_blank">{{ item.linkText }}</bk-link>
        </div>
      </div>
    </main>
  </section>
</template>
<script>
import { isEmpty } from '@/common/util';
import useProject from '@/views/project-manage/project/use-project';

export default {
  name: 'BcsUnregistry',
  props: {
    curProject: {
      type: Object,
      default: () => ({}),
      require: true,
    },
  },
  data() {
    return {
      kindList: [],
      guideList: [],
      kind: 'k8s',
      ccList: [],
      ccKey: '',
      isLoading: false,
    };
  },
  computed: {
    enableBtn() {
      return !isEmpty(this.ccKey);
    },
    iamDocLink() {
      return `${window.BK_IAM_HOST}/apply-custom-perm?system_id=bk_cmdb`;
    },
    ccDocLink() {
      return `${window.BK_CC_HOST}/#/resource/business`;
    },
  },
  created() {
    this.kindList = [
      {
        id: 'k8s',
        name: 'K8S',
        desc: this.$t('bcs.registry.label.projectKind.k8s'),
      },
    ];
    this.guideList = [
      {
        id: 'binding',
        iconColor: '#4540DC',
        desc: this.$t('bcs.registry.button.bindBusiness.desc'),
        link: this.ccDocLink,
        linkText: this.$t('bcs.registry.button.bindBusiness.text'),
      },
      {
        id: 'auth',
        iconColor: '#FFB200',
        desc: this.$t('bcs.registry.button.applyPerm.desc'),
        link: this.iamDocLink,
        linkText: this.$t('iam.button.apply2'),
      },
      {
        id: 'wiki',
        iconColor: '#66EFE3',
        desc: this.$t('bcs.registry.button.docs.desc'),
        link: this.PROJECT_CONFIG.quickStart,
        linkText: this.$t('blueking.docs'),
      },
    ];
  },
  mounted() {
    this.fetchCCList();
  },
  methods: {
    handleKindChange(item) {
      if (item.disabled) return;

      this.kind = item.id;
    },
    /**
     * 启用容器服务 更新项目
     */
    async updateProject() {
      try {
        const { updateProject } = useProject();

        this.isLoading = true;
        const result =  await updateProject(Object.assign({}, this.curProject, {
          // deployType 值固定，就是原来页面上的：部署类型：容器部署
          deployType: 2,
          // kind 业务编排类型
          kind: this.kind,
          // useBKRes 值固定，就是原来页面上的：使用蓝鲸部署服务
          useBKRes: true,
          businessID: String(this.ccKey),
        }));

        this.isLoading = false;

        this.$nextTick(() => {
          result && window.location.reload();
        });
      } catch (e) {
        console.error(e);
        this.isLoading = false;
      }
    },
    /**
             * 获取关联 CC 的数据
             */
    async fetchCCList() {
      if (!this.curProject.project_id) return;

      const { getBusinessList } = useProject();
      this.ccList = await getBusinessList();
    },
  },
};
</script>
<style scoped lang="postcss">
@define-mixin row-center {
    display: flex;
    justify-content: center;
}
@define-mixin column-center {
    display: flex;
    flex-direction: column;
    align-items: center;
}

.mt8 {
    margin-top: 8px;
}

.bcs-unregistry {
    overflow: hidden;
    font-size: 12px;
    .header {
        @mixin row-center;
        font-size: 20px;
        line-height: 20px;
        color: #222;
        margin-top: 86px;
        margin-bottom: 32px;
    }
    .main {
        @mixin column-center;
        .form-item {
            width: 720px;
            &-label {
                font-size: 14px;
                line-height: 14px;
                color: #000;
                margin-bottom: 8px;
            }
            &-content {
                @mixin row-center;
                .kind-panel {
                    min-width: 350px;
                    height: 64px;
                    border: 1px solid #c3cdd7;
                    border-radius: 3px;
                    background: #fff;
                    cursor: pointer;
                    padding: 12px 16px;
                    &:nth-child(1) {
                        margin-right: 24px;
                    }
                    &.active {
                        background: #E1ECFF;
                        border-color: #3A84FF;
                    }
                    &.disabled {
                        border-color: #dcdee5;
                        color: #c4c6cc;
                        cursor: not-allowed;
                    }
                    &-title {
                        font-size: 14px;
                        font-weight: 700;
                    }
                }
                .cc-selector {
                    width: 100%;
                }
                &.cc-list {
                    background: #fff;
                }
                &.kind {
                    justify-content: flex-start;
                }
            }
            &-tips {
                color: #63656e;
                margin-top: 8px;
                display: flex;
                align-items: center;
                >>> .bk-link-text {
                    font-size: 12px;
                    margin-left: 2px;
                }
            }
            &.enable-bcs {
                margin-top: 36px;
            }
            &.guide {
                display: flex;
                margin-top: 82px;
                .guide-link {
                    display: flex;
                    flex-direction: column;
                    align-items: flex-start;
                    width: 228px;
                    min-height: 160px;
                    background: #fff;
                    border-radius: 2px;
                    box-shadow: 0px 1px 1px 0px rgba(0,0,0,.09);
                    padding: 24px;
                    &.mr18 {
                        margin-right: 18px;
                    }
                    i {
                        font-size: 24px;
                    }
                }
            }
        }
    }
}
</style>
