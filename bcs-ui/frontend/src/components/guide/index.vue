<template>
    <div class="biz-actions" v-clickoutside="hide">
        <a :href="PROJECT_CONFIG.doc.contact" class="bk-text-button" v-bk-tooltips.top="$t('蓝鲸容器助手')" v-if="$INTERNAL">{{$t('联系我们')}}</a>
        <a :href="PROJECT_CONFIG.doc.help" target="_blank" class="bk-text-button">{{$t('帮助')}}</a>
        <slot></slot>
    </div>
</template>

<script>
    import clickoutside from '@open/directives/clickoutside'

    export default {
        directives: {
            clickoutside
        }
    }
</script>

<style>
    @import './index.css';
</style>
