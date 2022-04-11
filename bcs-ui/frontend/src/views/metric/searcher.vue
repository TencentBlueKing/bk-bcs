<template>
    <div class="metric-searcher">
        <template v-if="localScopeList.length && !clusterFixed">
            <bk-dropdown-menu ref="dropdown" trigger="click" :align="'left'">
                <bk-button class="bk-button trigger-btn" slot="dropdown-trigger">
                    <span class="btn-text">{{curScope.name}}</span><i class="bcs-icon bcs-icon-angle-down"></i>
                </bk-button>
                <ul class="bk-dropdown-list" slot="dropdown-content">
                    <li class="dropdown-item">
                        <a href="javascript:;"
                            v-for="scopeItem of localScopeList"
                            :title="scopeItem.name"
                            :key="scopeItem.id"
                            :class="{ active: scopeItem.id === curScope.id }"
                            @click="handleSechScope(scopeItem)"
                        >{{scopeItem.name}}</a>
                    </li>
                </ul>
            </bk-dropdown-menu>
        </template>
        <div class="biz-search-input" style="width: 300px;">
            <bkbcs-input right-icon="bk-icon icon-search"
                clearable
                :placeholder="placeholderRender"
                v-model="localKey"
                @enter="handleSearch"
                @clear="clearSearch" />
        </div>
        <div class="biz-refresh-wrapper" v-if="widthRefresh">
            <bcs-popover class="refresh" :content="$t('刷新')" :delay="500" placement="top">
                <bk-button :class="['bk-button bk-default is-outline is-icon']" @click="handleRefresh">
                    <i class="bcs-icon bcs-icon-refresh"></i>
                </bk-button>
            </bcs-popover>
        </div>
    </div>
</template>

<script>
    export default {
        props: {
            placeholder: {
                type: String,
                default: ''
            },
            searchKey: {
                type: String,
                default: ''
            },
            searchScope: {
                type: String,
                default: ''
            },
            widthRefresh: {
                type: Boolean,
                default: true
            },
            scopeList: {
                type: Array,
                default () {
                    return []
                }
            },
            clusterFixed: {
                type: Boolean,
                default: false
            }
        },
        data () {
            return {
                isTriggerSearch: false,
                isRefresh: false,
                localKey: this.searchKey,
                localScopeList: [],
                curScope: {
                    id: '',
                    name: this.$t('全部集群')
                },
                placeholderRender: ''
            }
        },
        watch: {
            searchKey (val) {
                this.localKey = val
            },
            scopeList () {
                this.initLocalScopeList()
            },
            localKey (newVal, oldVal) {
                // 如果删除，为空时触发搜索
                if (oldVal && !newVal && !this.isRefresh) {
                    this.clearSearch()
                }
            },
            searchScope (v) {
                const curScope = this.localScopeList.find(item => item.id === v)
                this.curScope = Object.assign({}, curScope)
            }
        },
        created () {
            this.initLocalScopeList()
            this.placeholderRender = this.placeholder || this.$t('输入关键字，按Enter搜索')
        },
        methods: {
            handleSechScope (data) {
                this.curScope = data
                // this.$refs.dropdown.hide()
                sessionStorage['bcs-cluster'] = this.curScope.id
                this.$emit('update:searchScope', this.curScope.id)
            },
            initLocalScopeList () {
                this.localScopeList = JSON.parse(JSON.stringify(this.scopeList))
                if (this.localScopeList.length) {
                    if (this.searchScope) {
                        this.curScope = this.localScopeList.find(item => item.id === this.searchScope)
                    } else {
                        this.curScope = this.localScopeList[0]
                    }
                }
            },
            handleSearch () {
                this.isTriggerSearch = true
                this.$emit('update:searchKey', this.localKey)
                this.isRefresh = false
            },
            handleRefresh () {
                this.isRefresh = true
                this.$emit('refresh')
            },
            clearSearch () {
                this.localKey = ''
                if (this.isTriggerSearch) {
                    this.handleSearch()
                    this.isTriggerSearch = false
                }
            }
        }
    }
</script>

<style lang="postcss">
    @import '@/css/mixins/clearfix.css';
    @import '@/css/mixins/ellipsis.css';
    .metric-searcher {
        @mixin clearfix;

        .biz-search-input {
            .bk-form-input {
                border-radius: 0 2px 2px 0;
            }
        }

        .bk-dropdown-menu {
            .dropdown-item {
                > a {
                    width: 100%;
                    cursor: pointer;
                    display: inline-block;
                    vertical-align: middle;
                    @mixin ellipsis 240px;
                }
                .active {
                    background-color: #eef6fe;
                    color: #3a84ff;
                }
            }
            .bk-button {
                border-radius: 2px 0 0 2px;
                border-right: none;
            }
            float: left;
        }
        .btn-text {
            width: 140px;
            text-align: left;
            display: inline-block;
            vertical-align: middle;
            @mixin ellipsis 150px;
        }
    }
</style>
