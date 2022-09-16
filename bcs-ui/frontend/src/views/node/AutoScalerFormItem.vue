<template>
    <bcs-form class="form-content" :label-width="210">
        <bcs-form-item v-for="item in list"
            :key="item.prop"
            :label="item.name"
            class="form-content-item"
            style="width: 50%;"
            desc-type="icon"
            desc-icon="bk-icon icon-info-circle"
            :desc="item.desc">
            <template v-if="item.prop === 'status'">
                <LoadingIcon v-if="autoscalerData[item.prop] === 'UPDATING'">
                    {{scalerStatusMap[autoscalerData[item.prop]]}}
                </LoadingIcon>
                <StatusIcon
                    :status="autoscalerData[item.prop]"
                    :status-color-map="scalerColorMap"
                    v-else>
                    {{scalerStatusMap[autoscalerData[item.prop]] || $t('未知')}}
                </StatusIcon>
            </template>
            <span v-else-if="typeof autoscalerData[item.prop] === 'boolean'">
                {{autoscalerData[item.prop] ? $t('是') : $t('否')}}
            </span>
            <span v-else>
                {{`${autoscalerData[item.prop]} ${item.unit || ''}`}}
                <span v-if="item.suffix" class="ml10">{{item.suffix}}</span>
            </span>
        </bcs-form-item>
    </bcs-form>
</template>
<script lang="ts">
    import { defineComponent } from '@vue/composition-api'
    import $i18n from '@/i18n/i18n-setup'

    export default defineComponent({
        name: 'AutoScalerFormItem',
        props: {
            list: {
                type: Array,
                default: () => []
            },
            autoscalerData: {
                type: Object,
                default: () => ({})
            }
        },
        setup () {
            // 获取自动扩缩容配置
            const scalerStatusMap = { // 自动扩缩容状态
                NORMAL: $i18n.t('正常'),
                UPDATING: $i18n.t('更新中'),
                'UPDATE-FAILURE': $i18n.t('更新失败'),
                STOPPED: $i18n.t('已停用')
            }
            const scalerColorMap = {
                NORMAL: 'green',
                UPDATING: 'green',
                'UPDATE-FAILURE': 'red',
                STOPPED: 'gray'
            }
            return {
                scalerStatusMap,
                scalerColorMap
            }
        }
    })
</script>
<style lang="postcss" scoped>
>>> .form-content {
  display: flex;
  flex-wrap: wrap;
  &-item {
      height: 32px;
      margin-top: 0;
      font-size: 12px;
      width: 100%;
      user-select: none;
  }
  .bk-label {
      font-size: 12px;
      color: #979BA5;
      text-align: left;
  }
  .bk-form-content {
      font-size: 12px;
      color: #313238;
      display: flex;
      align-items: center;
  }
}
</style>
