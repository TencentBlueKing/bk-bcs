<template>
    <div class="biz-content" v-bkloading="{ isLoading: showLoading, opacity: 1 }">
        <div class="biz-top-bar">
            <div class="biz-image-detail-title">
                <span class="bcs-icon bcs-icon-arrows-left" style="color: #3a84ff; cursor: pointer; font-weight: 600;" @click="backImageLibrary"></span>
                {{imageName}}
            </div>
        </div>
        <div class="biz-image-detail-content-wrapper">
            <div class="biz-header-content">
                <div class="left-wrapper">
                    <div class="logo">
                        <img src="@/images/default_logo_normal.jpg" />
                    </div>
                    <div class="left-content">
                        <p class="image-name" :title="imageName">{{imageName}}</p>
                        <p class="download-count">{{downloadCount + $t('次下载')}}</p>
                    </div>
                </div>
                <div class="right-wrapper">
                    <div class="top-content">
                        <div class="updator">
                            <p>{{$t('最近更新人')}}</p>
                            <p>{{modifiedBy}}</p>
                        </div>
                        <div class="update-date">
                            <p>{{$t('最近更新时间')}}</p>
                            <p>{{modified}}</p>
                        </div>
                    </div>
                    <div class="bottom-content">
                        <p>{{$t('仓库相对地址')}}</p>
                        <p>{{imagePath}}</p>
                    </div>
                </div>
            </div>
            <div class="tag-count" v-if="dataList.length"><span>TAG</span>{{'(' + tagCount + ')'}}</div>
            <table class="bk-table has-table-hover biz-table biz-image-detail-table" :style="{ minHeight: `${tableHeight - 40}px` }">
                <thead>
                    <tr>
                        <th style="width: 15%; text-align: left;padding-left: 30px;">
                            {{$t('名称')}}
                        </th>
                        <th>{{$t('大小')}}</th>
                        <th>{{$t('最近更新时间')}}</th>
                        <th>{{$t('地址')}}</th>
                    </tr>
                </thead>
                <tbody>
                    <template v-if="dataList.length">
                        <tr v-for="(item, index) in dataList" :key="index">
                            <td style="text-align: left;padding-left: 30px;">
                                {{item.tag || '--'}}
                            </td>
                            <td>{{item.size || '--'}}</td>
                            <td>{{item.modified || '--'}}</td>
                            <td>
                                <template v-if="$INTERNAL">
                                    <template v-if="item.artifactorys.length">
                                        <bk-tag type="filled" :theme="art === 'DEV' ? 'warning' : 'success'" v-for="(art, inx) in item.artifactorys" :key="inx">
                                            {{art === 'DEV' ? $t('研发仓库') : $t('生产仓库')}}
                                        </bk-tag>
                                    </template>
                                    <template v-else>
                                        --
                                    </template>
                                </template>
                                <template v-else>
                                    {{item.image || '--'}}
                                </template>
                            </td>
                        </tr>
                        <tr v-if="showScrollLoading">
                            <td colspan="4">
                                <div class="loading-row" v-bkloading="{ isLoading: true }"></div>
                            </td>
                        </tr>
                        <tr v-if="!hasNext" class="empty-row">
                            <td colspan="4" class="tc">
                                {{$t('没有更多TAG')}}
                            </td>
                        </tr>
                    </template>
                    <template v-else>
                        <tr class="no-hover">
                            <td colspan="4">
                                <div class="bk-message-box">
                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                </div>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>
    </div>
</template>

<script>
    import { getScrollHeight, getScrollTop, getWindowHeight } from '@/common/util'

    export default {
        props: {
            imageRepo: {
                type: String,
                default: '',
                required: true
            }
        },
        data () {
            return {
                showLoading: true,
                showScrollLoading: false,
                hasNext: true,
                hasPrevious: false,
                isBeforeDestroy: false,
                bkMessageInstance: null,
                imageName: '',
                repo: '',
                dataList: [],
                imageDetailList: [],
                tagCount: 0,
                created: '',
                modified: '',
                modifiedBy: '',
                imagePath: '',
                downloadCount: 0,
                tableHeight: window.innerHeight - 20 - 219 - 20 - 20 - 10,
                pageSize: 0,
                curPage: 1
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            backRouterName () {
                return localStorage.getItem('backRouterName')
            },
            curImage () {
                return this.$store.state.depot.curImage
            }
        },
        async created () {
            if (!this.imageRepo) {
                this.$router.push({
                    name: this.backRouterName
                })
            } else {
                this.pageSize = Math.ceil(this.tableHeight / 41 + 10)
                this.isBeforeDestroy = false
                await this.fetchImageLibraryData(this.imageRepo, (this.curPage - 1) * this.pageSize, this.pageSize)
            }
        },
        async mounted () {
            self.addEventListener('resize', () => {
                if (this.tagCount < 10) {
                    this.tableHeight = this.tagCount * 41
                } else {
                    this.tableHeight = window.innerHeight - 20 - 219 - 20 - 20 - 10
                }
            })
            self.addEventListener('scroll', async e => {
                if (this.showScrollLoading || this.isBeforeDestroy || this.tagCount < 10) {
                    return
                }
                if (getScrollTop() + getWindowHeight() >= getScrollHeight()) {
                    if (this.hasNext) {
                        this.showScrollLoading = true
                        this.curPage = this.curPage + 1
                        await this.fetchImageLibraryData(this.imageRepo, (this.curPage - 1) * this.pageSize, this.pageSize)
                    }
                }
            })
        },
        beforeDestroy () {
            this.isBeforeDestroy = true
        },
        destroyed () {
            this.bkMessageInstance && this.bkMessageInstance.close()
        },
        methods: {
            backImageLibrary () {
                this.$router.push({
                    name: this.backRouterName
                })
            },
            reset () {
                this.imageDetailList.splice(0, this.imageDetailList.length, ...[])
                this.tagCount = 0
                this.imageName = ''
                this.created = ''
                this.modified = ''
                this.modifiedBy = ''
                this.imagePath = ''
                this.downloadCount = 0
            },

            /**
             * 获取详情数据
             */
            async fetchImageLibraryData (repo = '', offset = 0, limit = 10) {
                if (!this.showScrollLoading) {
                    this.reset()
                    this.showLoading = true
                }
                const params = {
                    image_repo: repo,
                    offset: offset,
                    limit: limit,
                    projectId: this.projectId
                }
                try {
                    const res = await this.$store.dispatch('depot/getImageLibraryDetail', params)

                    if (this.showScrollLoading) {
                        this.hasNext = res.data.has_next
                        this.hasPrevious = res.data.has_previous
                    } else {
                        this.tagCount = res.data.tagCount || 0
                        if (this.tagCount < 10) {
                            this.tableHeight = this.tagCount * 41
                            this.hasNext = true
                        }
                        this.imageName = res.data.imageName || '--'
                        this.created = res.data.created || '--'
                        this.modified = res.data.modified || '--'
                        this.modifiedBy = res.data.modifiedBy || '--'
                        this.imagePath = res.data.repo || '--'
                        this.downloadCount = res.data.downloadCount || 0
                    }
                    this.imageDetailList.splice(0, this.imageDetailList.length, ...(res.data.tags || []))
                    if (this.imageDetailList.length) {
                        this.imageDetailList.forEach(item => {
                            this.dataList.push(item)
                        })
                    }
                } catch (e) {
                    console.warn(333, e)
                } finally {
                    this.showLoading = false
                    const scrollTop = getScrollTop()
                    if (scrollTop !== 0) {
                        window.scrollTo(0, getScrollTop() - 40)
                    }
                    setTimeout(() => {
                        this.showScrollLoading = false
                    }, 100)
                }
            }
        }
    }
</script>

<style scoped>
    @import './image-detail.css';
</style>
