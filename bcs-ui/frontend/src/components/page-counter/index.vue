<template>
    <div class="biz-page-count">
        <div class="total-page">
            {{$t('共计{total}条', { total: total || totalPage })}}
        </div>
        <div class="page-count-selector">
            <template v-if="isEn">
                <bk-selector
                    style="width: 100px; display: inline-block;"
                    :placeholder="placeholder"
                    :selected.sync="selected"
                    :list="pageCounterListEN"
                    @item-selected="pageCounterSelect">
                </bk-selector>
            </template>
            <template v-else>
                {{$t('每页')}}
                <bk-selector
                    style="width: 70px; display: inline-block;"
                    :placeholder="placeholder"
                    :selected.sync="selected"
                    :list="pageCounterList"
                    @item-selected="pageCounterSelect">
                </bk-selector>
                {{$t('条')}}
            </template>
        </div>
    </div>
</template>

<script>
    export default {
        props: {
            totalPage: {
                type: Number,
                default: 0
            },
            total: {
                type: Number,
                default: 0
            },
            pageSize: {
                type: Number,
                default: 10
            },
            placeholder: {
                type: String,
                default: window.i18n.t('请选择要实例化的模板')
            },
            isEn: {
                type: Boolean,
                default: false
            }
        },
        data () {
            return {
                selected: this.pageSize,
                pageCounterList: [
                    { id: 5, name: '5' },
                    { id: 10, name: '10' },
                    { id: 20, name: '20' },
                    { id: 50, name: '50' },
                    { id: 100, name: '100' }
                ],
                pageCounterListEN: [
                    { id: 5, name: '5/page' },
                    { id: 10, name: '10/page' },
                    { id: 20, name: '20/page' },
                    { id: 50, name: '50/page' },
                    { id: 100, name: '100/page' }
                ]
            }
        },
        methods: {
            pageCounterSelect (data) {
                this.$emit('change', data)
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
