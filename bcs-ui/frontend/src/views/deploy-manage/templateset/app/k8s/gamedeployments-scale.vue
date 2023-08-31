<template>
  <bk-dialog
    :is-show.sync="isVisible"
    :width="380"
    :title="$t('deploy.templateset.scale')"
    :close-icon="!isUpdating"
    :quick-close="false"
    @cancel="hideScale">
    <template slot="content">
      <div class="gamestatefulset-scale-wrapper" v-bkloading="{ isLoading: isLoading, opacity: 1 }">
        <div class="bk-form-item">
          <label class="bk-label">
            {{$t('dashboard.workload.label.scaleNum')}}
          </label>
          <div class="bk-form-content" style="display: inline-block; margin-left: 10px;">
            <bk-input
              type="number"
              v-model="scaleNum"
              :min="0"
              :max="5000"
              :style="{ 'width': '260px' }"
              :placeholder="$t('generic.placeholder.input')">
            </bk-input>
          </div>
        </div>
      </div>
    </template>
    <div slot="footer">
      <div class="bk-dialog-outer">
        <template v-if="isUpdating">
          <bk-button type="primary" disabled>
            {{$t('generic.status.updating')}}
          </bk-button>
          <bk-button type="button" class="bk-dialog-btn bk-dialog-btn-cancel disabled">
            {{$t('generic.button.cancel')}}
          </bk-button>
        </template>
        <template v-else>
          <bk-button
            type="primary" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
            @click="confirmScale">
            {{$t('generic.button.confirm')}}
          </bk-button>
          <bk-button type="button" @click="hideScale">
            {{$t('generic.button.cancel')}}
          </bk-button>
        </template>
      </div>
    </div>
  </bk-dialog>
</template>

<script>
import { catchErrorHandler } from '@/common/util';

export default {
  name: 'GamestatefulsetScale',
  props: {
    isShow: {
      type: Boolean,
      default: false,
    },
    clusterId: {
      type: String,
    },
    item: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      bkMessageInstance: null,
      width: 740,
      isVisible: false,
      isLoading: false,
      isUpdating: false,
      scaleNum: 0,
    };
  },
  computed: {
    projectId() {
      return this.$route.params.projectId;
    },
    projectCode() {
      return this.$route.params.projectCode;
    },
    isEn() {
      return this.$store.state.isEn;
    },
  },
  watch: {
    isShow: {
      async handler(newVal) {
        this.isVisible = newVal;
        if (!this.isVisible) {
          return;
        }
        if (this.isVisible) {
          this.isLoading = true;
          this.renderItem = Object.assign({}, this.item || {});
          await this.fetchData();
        }
      },
      immediate: true,
    },
  },
  mounted() {
  },
  destroyed() {
    this.bkMessageInstance?.close();
  },
  methods: {
    /**
             * 获取 yaml 数据
             */
    async fetchData() {
      try {
        const res = await this.$store.dispatch('app/getGameStatefulsetInfo', {
          projectId: this.projectId,
          clusterId: this.clusterId,
          gamestatefulsets: this.crd || 'gamedeployments.tkex.tencent.com',
          name: this.renderItem.name,
          data: {
            namespace: this.renderItem.namespace,
          },
        });
        const data = res.data || {};
        this.scaleNum = data.spec.replicas || 0;
      } catch (e) {
        console.error(e);
      } finally {
        this.isLoading = false;
      }
    },

    /**
             * 确定扩缩容
             */
    async confirmScale() {
      if (this.scaleNum === null || this.scaleNum === undefined || this.scaleNum === '') {
        this.bkMessageInstance?.close();
        this.bkMessageInstance = this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.input'),
        });
        return;
      }

      try {
        this.isUpdating = true;
        await this.$store.dispatch('app/scaleGameStatefulsetInfo', {
          projectId: this.projectId,
          clusterId: this.renderItem.cluster_id,
          gamestatefulsets: 'gamedeployments.tkex.tencent.com',
          name: this.renderItem.name,
          data: {
            namespace: this.renderItem.namespace,
            body: {
              spec: {
                replicas: parseInt(this.scaleNum, 10),
              },
            },
          },
        });
        this.bkMessageInstance?.close();
        this.bkMessageInstance = this.$bkMessage({
          theme: 'success',
          message: this.$t('deploy.templateset.success'),
        });
        this.$emit('scale-success');
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        this.isUpdating = false;
      }
    },

    /**
             * 隐藏扩缩容弹框
             */
    hideScale() {
      this.$emit('hide-scale');
    },
  },
};
</script>
<style lang="postcss">
    @import '@/css/variable.css';

    .gamestatefulset-scale {
        .bk-dialog-tool {
            position: absolute;
            top: 0;
            right: 0;
        }
        .bk-dialog-header {
            background: #fafbfd;
            border-radius: 2px;
            color: #737987;
            height: 50px;
            line-height: 50px;
            padding: 0 20px;
            .bk-dialog-title {
                color: #737987;
                text-align: left;
                font-size: 18px;
            }
        }
        button.disabled {
            background-color: #fafafa;
            border-color: $borderLightColor;
            color: #ccc;
            cursor: not-allowed;

            &:hover {
                background-color: #fafafa;
                border-color: $borderLightColor;
            }
        }

        .gamestatefulset-scale-wrapper {
            position: relative;
            .bk-number .bk-number-content.bk-number-larger .bk-number-icon-content {
                margin-top: 0 !important;
            }
        }
    }
</style>
