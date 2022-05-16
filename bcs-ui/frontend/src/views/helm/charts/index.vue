<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-tpl-title">
                Charts
            </div>
            <bk-guide>
                <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="showGuide">{{$t('如何推送Helm Chart到项目仓库？')}}</a>
            </bk-guide>
        </div>

        <guide ref="clusterGuide"></guide>

        <div class="biz-content-wrapper biz-tpl-wrapper" style="padding: 0; margin: 0;" v-bkloading="{ isLoading: showLoading, opacity: 0.1 }">
            <template v-if="!showLoading">
                <div class="biz-panel-header" style="padding: 20px;">
                    <div class="right right-flex">
                        <!-- <bcs-select v-model="selectedName" :clearable="false">
                            <bcs-option
                                v-for="item in chartTypeList"
                                :key="item.id"
                                :id="item.id"
                                :name="item.name"
                            >
                            </bcs-option>
                        </bcs-select> -->
                        <div class="biz-search-input mr10" style="width: 300px;">
                            <bk-input right-icon="bk-icon icon-search"
                                :placeholder="$t('输入关键字，按Enter搜索')"
                                clearable
                                v-model="searchKeyword"
                                @enter="search"
                                @clear="clearSearch" />
                        </div>
                        <i class="bcs-icon bcs-icon-list toggle-layout-icon" v-if="isCardLayout" @click="handleToggleLayout"></i>
                        <i class="bcs-icon bcs-icon-apps toggle-layout-icon" v-else @click="handleToggleLayout"></i>
                    </div>
                </div>
                <component
                    :is="layoutComponent"
                    :selected-name="selectedName"
                    :tpl-list="tplList"
                    @fetchList="getTplList"
                >
                </component>
            </template>
        </div>
        <svg style="display: none;">
            <title>{{$t('模板集默认图标')}}</title>
            <symbol id="biz-set-icon" viewBox="0 0 60 60">
                <g id="图层_6">
                    <g id="图层_32_1_">
                        <path class="st0" d="M12,8v4H8c-1.1,0-2,0.9-2,2v42c0,1.1,0.9,2,2,2h42c1.1,0,2-0.9,2-2v-4h4c1.1,0,2-0.9,2-2V8c0-1.1-0.9-2-2-2
                            H14C12.9,6,12,6.9,12,8z M48,48v4v2H10V16h2h4h32V48z M54,48h-2V14c0-1.1-0.9-2-2-2H16v-2h38V48z" />
                    </g>
                    <path class="st1" d="M45.7,33.7h-1.8l-3.4-8.3l1.3-1.3c0.5-0.5,0.5-1.3,0-1.8l0,0c-0.5-0.5-1.3-0.5-1.8,0l-1.3,1.3l-8.4-3.5v-1.8
                        c0-0.7-0.6-1.3-1.3-1.3l0,0c-0.7,0-1.3,0.6-1.3,1.3V20l-8.4,3.5l-1.2-1.2c-0.5-0.5-1.3-0.5-1.8,0l0,0c-0.5,0.5-0.5,1.3,0,1.8
                        l1.2,1.2L14,33.7h-1.8c-0.7,0-1.3,0.6-1.3,1.3l0,0c0,0.7,0.6,1.3,1.3,1.3H14l3.5,8.4L16.2,46c-0.5,0.5-0.5,1.3,0,1.8l0,0
                        c0.5,0.5,1.3,0.5,1.8,0l1.3-1.3l8.3,3.4v1.8c0,0.7,0.6,1.3,1.3,1.3l0,0c0.7,0,1.3-0.6,1.3-1.3v-1.9l8.3-3.4l1.3,1.3
                        c0.5,0.5,1.3,0.5,1.8,0l0,0c0.5-0.5,0.5-1.3,0-1.8l-1.3-1.3l3.4-8.3h1.9c0.7,0,1.3-0.6,1.3-1.3l0,0C47,34.3,46.4,33.7,45.7,33.7z
                            M30.3,23.4l6,2.5l-4.6,4.6c-0.4-0.2-0.9-0.4-1.3-0.6v-6.5H30.3z M27.7,23.4V30c-0.5,0.1-0.9,0.3-1.4,0.6l-4.7-4.7L27.7,23.4z
                            M19.9,27.7l4.7,4.7c-0.2,0.4-0.4,0.9-0.5,1.3h-6.6L19.9,27.7z M17.4,36.3H24c0.1,0.5,0.3,0.9,0.6,1.3l-4.7,4.7L17.4,36.3z
                            M27.7,46.5l-6-2.5l4.7-4.7c0.4,0.2,0.8,0.4,1.3,0.5V46.5z M29,37.5c-1.4,0-2.6-1.2-2.6-2.6c0-1.4,1.2-2.6,2.6-2.6s2.6,1.2,2.6,2.6
                        C31.6,36.4,30.4,37.5,29,37.5z M30.3,46.5v-6.6c0.5-0.1,0.9-0.3,1.3-0.5l4.6,4.6L30.3,46.5z M38,42.2l-4.6-4.6
                        c0.2-0.4,0.4-0.8,0.6-1.3h6.5L38,42.2z M34,33.7c-0.1-0.5-0.3-0.9-0.5-1.3l4.6-4.6l2.5,6H34V33.7z" />
                    <g class="st2">
                        <path class="st3" d="M41,49H17c-1.1,0-2-0.9-2-2V23c0-1.1,0.9-2,2-2h24c1.1,0,2,0.9,2,2v24C43,48.1,42.1,49,41,49z" />
                    </g>
                    <g>
                        <path class="st0" d="M42.2,25c-1.9,0-2.9,0.5-2.9,1.5v17.1c0,1,1,1.5,2.9,1.5v1.8H31.4V45c2,0,3-0.5,3-1.5v-8H23.6v8
                            c0,1,1,1.5,3,1.5v1.8H15.8V45c1.9,0,2.8-0.5,2.8-1.5V26.4c0-1-0.9-1.5-2.8-1.5V23h10.8v2c-2,0-3,0.5-3,1.5v6.8h10.8v-6.8
                            c0-1-1-1.5-3-1.5v-1.9h10.8V25z" />
                    </g>
                </g>
            </symbol>
        </svg>
    </div>
</template>

<script>
    import { ref, reactive, computed, watch, onMounted } from '@vue/composition-api'
    import tplListTable from './tpl-list-table'
    import tplListCard from './tpl-list-card'
    import Guide from './guide'

    const cardLayout = 'IS_CARD_LAYOUT'
    export default {
        name: 'Charts',
        components: {
            Guide,
            tplListTable,
            tplListCard
        },
        setup (props, ctx) {
            const { $i18n, $store, $route } = ctx.root

            const selectedName = ref('privateRepo')
            const isCardLayout = ref(false)
            const chartTypeList = reactive([
                { id: 'privateRepo', name: $i18n.t('项目仓库') },
                { id: 'publicRepo', name: $i18n.t('公共仓库') }
            ])
            const curProjectId = ref('')
            const showLoading = ref(false)
            const isSearchMode = ref(false)
            const publicTplList = ref([])
            const privateTplList = ref([])
            const tplList = ref([])
            const tplListCache = ref([])
            const searchKeyword = ref('')
            const clusterGuide = ref(null)
        
            const layoutComponent = computed(() => {
                return isCardLayout.value ? tplListCard : tplListTable
            })

            // const curProject = computed(() => {
            //     return this.$store.state.curProject
            // })

            const projectCode = computed(() => {
                return $store.state.curProjectCode
            })

            const projectId = computed(() => {
                curProjectId.value = $route.params.projectId
                return curProjectId.value
            })
            console.log(projectId)

            watch(selectedName, (val) => {
                setTplList()
            })

            watch(searchKeyword, (newVal, oldVal) => {
                if (oldVal && !newVal) {
                    search()
                }
            })

            const handleToggleLayout = () => {
                isCardLayout.value = !isCardLayout.value
                localStorage.setItem(cardLayout, isCardLayout.value)
            }

            /**
             * 获取模板列表
             */
            const getTplList = async () => {
                const operator = $store.state.user.username
                showLoading.value = true

                const res = await $store.dispatch('helm/getTplList', {
                    $projectId: projectCode.value,
                    $repository: projectCode.value,
                    page: 1,
                    size: 1500,
                    operator
                }).catch(() => false)

                showLoading.value = false
                if (!res) return

                tplList.value = res.data
            }

            /**
             * 根据当前显示公有/私有模板
             */
            const setTplList = () => {
                clearSearch()
                if (selectedName.value === 'privateRepo') {
                    tplList.value = JSON.parse(JSON.stringify(privateTplList.value))
                } else {
                    tplList.value = JSON.parse(JSON.stringify(publicTplList.value))
                }
                tplListCache.value = JSON.parse(JSON.stringify(tplList.value))
            }

            /**
             * 清除搜索
             */
            const clearSearch = () => {
                searchKeyword.value = ''
                search()
            }

            /**
             * 搜索
             */
            const search = () => {
                const keyword = searchKeyword.value
                if (keyword) {
                    const results = tplListCache.value.filter(item => item.name.indexOf(keyword) > -1)
                    tplList.value.splice(0, tplList.value.length, ...results)
                    isSearchMode.value = true
                } else {
                    tplList.value.splice(0, tplList.value.length, ...tplListCache.value)
                    isSearchMode.value = false
                }
            }
            
            /**
             * 显示引导层(如何推送Helm Chart到项目仓库？)
             */
            const showGuide = () => {
                clusterGuide.value.show()
            }

            onMounted(() => {
                if (localStorage.getItem(cardLayout) && localStorage.getItem(cardLayout) === 'true') {
                    isCardLayout.value = true
                }
                getTplList()
            })

            return {
                tplList,
                searchKeyword,
                showLoading,
                isSearchMode,
                selectedName,
                isCardLayout,
                chartTypeList,
                layoutComponent,
                clusterGuide,
                search,
                showGuide,
                clearSearch,
                getTplList,
                handleToggleLayout
            }
        }
    }
</script>

<style lang="postcss" scoped>
    @import '@/css/variable.css';
    @import '@/css/mixins/ellipsis.css';
    @import '@/css/mixins/clearfix.css';
    @import '@/css/mixins/scroller.css';
    
    .biz-tpl-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 20px;
    }
    .right-flex {
        display: flex;
        align-items: center;
    }
    .toggle-layout-icon {
        cursor: pointer;
    }
</style>
