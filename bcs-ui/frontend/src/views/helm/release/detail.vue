<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-helm-title">
                <a class="bcs-icon bcs-icon-arrows-left back" href="javascript:void(0);" @click="goToHelmIndex"></a>
                appName
            </div>
        </div>
        <div class="biz-content-wrapper" v-bkloading="{ isLoading: isLoading }">
            <div class="basic-info">
                <div class="title">{{ $t('基本信息') }}</div>
                <div class="info">
                    <span class="appName">{{ releaseDetail.name }}</span>
                    <span class="status info-item">
                        <span class="key">{{ $t('状态') }}</span>
                        <span class="value">{{ $t('已部署') }}</span>
                    </span>
                    <span class="namespace info-item">
                        <span class="key">{{ $t('命名空间') }}</span>
                        <span class="value">{{ releaseDetail.namespace }}</span>
                    </span>
                    <span class="update info-item">
                        <span class="key">{{ $t('更新时间') }}</span>
                        <span class="value">{{ releaseDetail.updateTime && releaseDetail.updateTime.slice(0, 19) }}</span>
                    </span>
                    <!-- <span class="last info-item">
                        <span class="key">{{ $t('最后部署') }}</span>
                        <span class="value">10天</span>
                    </span> -->
                </div>
                <div class="desc">
                    <span class="key">{{ $t('简介') }}</span>
                    <span class="value">Ingress controller for Kubernetes using NGINX as a reverse proxy and load balancer</span>
                </div>
            </div>
            <div class="notes">
                <div class="title">Notes</div>
                <div class="notes-content">
                    图表名称：kubeapps 图表版本：7.8.7 应用程序版本：2.4.3** 在部署图表时请耐心等待** 提示： 使用以下命令查看部署状态：kubectl get pods -w --namespace kubeapps 可以通过集群内以下 DNS 名称上的端口 80 访问 Kubeapps： export CLUSTER_IP=$(minikube ip) # 在 Minikube 上。使用：kubectl cluster-info在其他 K8s 集群上 echo "Kubeapps URL:
                    <br>
                    <br>
                    <br>
                    图表名称：kubeapps 图表版本：7.8.7 应用程序版本：2.4.3** 在部署图表时请耐心等待** 提示： 使用以下命令查看部署状态：kubectl get pods -w --namespace kubeapps 可以通过集群内以下 DNS 名称上的端口 80 访问 Kubeapps： export CLUSTER_IP=$(minikube ip) # 在 Minikube 上。使用：kubectl cluster-info在其他 K8s 集群上 echo "Kubeapps URL:
                    图表名称：kubeapps 图表版本：7.8.7 应用程序版本：2.4.3** 在部署图表时请耐心等待** 提示： 使用以下命令查看部署状态：kubectl get pods -w --namespace kubeapps 可以通过集群内以下 DNS 名称上的端口 80 访问 Kubeapps： export CLUSTER_IP=$(minikube ip) # 在 Minikube 上。使用：kubectl cluster-info在其他 K8s 集群上 echo "Kubeapps URL:
                    <br>
                    图表名称：kubeapps 图表版本：7.8.7 应用程序版本：2.4.3** 在部署图表时请耐心等待** 提示： 使用以下命令查看部署状态：kubectl get pods -w --namespace kubeapps 可以通过集群内以下 DNS 名称上的端口 80 访问 Kubeapps： export CLUSTER_IP=$(minikube ip) # 在 Minikube 上。使用：kubectl cluster-info在其他 K8s 集群上 echo "Kubeapps URL:
                    <br>
                    图表名称：kubeapps 图表版本：7.8.7 应用程序版本：2.4.3** 在部署图表时请耐心等待** 提示： 使用以下命令查看部署状态：kubectl get pods -w --namespace kubeapps 可以通过集群内以下 DNS 名称上的端口 80 访问 Kubeapps： export CLUSTER_IP=$(minikube ip) # 在 Minikube 上。使用：kubectl cluster-info在其他 K8s 集群上 echo "Kubeapps URL:
                    <br>
                    图表名称：kubeapps 图表版本：7.8.7 应用程序版本：2.4.3** 在部署图表时请耐心等待** 提示： 使用以下命令查看部署状态：kubectl get pods -w --namespace kubeapps 可以通过集群内以下 DNS 名称上的端口 80 访问 Kubeapps： export CLUSTER_IP=$(minikube ip) # 在 Minikube 上。使用：kubectl cluster-info在其他 K8s 集群上 echo "Kubeapps URL:
                    <br>
                </div>
            </div>
            <div class="application">
                <div class="title">Application Resources</div>
                <bcs-tab class="mt20" type="card" :label-height="40">
                    <bcs-tab-panel name="Deployments" label="Deployments">
                        <bk-table>
                            <bk-table-column label="Key" prop="key"></bk-table-column>
                            <bk-table-column label="Value" prop="value"></bk-table-column>
                        </bk-table>
                    </bcs-tab-panel>
                    <bcs-tab-panel name="StatefulSets" label="StatefulSets">
                        <bk-table>
                            <bk-table-column label="Key" prop="key"></bk-table-column>
                            <bk-table-column label="Value" prop="value"></bk-table-column>
                        </bk-table>
                    </bcs-tab-panel>
                    <bcs-tab-panel name="Secrets" label="Secrets">
                        <bk-table>
                            <bk-table-column label="Key" prop="key"></bk-table-column>
                            <bk-table-column label="Value" prop="value"></bk-table-column>
                        </bk-table>
                    </bcs-tab-panel>
                    <bcs-tab-panel name="Services" label="Services">
                        <bk-table>
                            <bk-table-column label="Key" prop="key"></bk-table-column>
                            <bk-table-column label="Value" prop="value"></bk-table-column>
                        </bk-table>
                    </bcs-tab-panel>
                </bcs-tab>
            </div>
            <div class="installation">
                <div class="title">Installation Values</div>
                <ace
                    class="installation-editor"
                    :read-only="true"
                >
                </ace>
            </div>
        </div>
    </div>
</template>

<script>
    // import { onMounted, watch, computed, ref, reactive, onBeforeUnmount } from '@vue/composition-api'
    import { computed, onMounted, ref } from '@vue/composition-api'
    import ace from '@/components/ace-editor'

    export default {
        name: 'ReleaseStatus',
        components: {
            ace
        },

        setup (props, ctx) {
            const { $route, $router, $store } = ctx.root
            const activeTab = ref('mission')
            const isLoading = ref(false)
            const releaseDetail = ref({})

            const projectId = computed(() => {
                return $route.params.projectId
            })
            const projectCode = computed(() => {
                return $route.params.projectCode
            })

            /**
             * 返回Helm应用首页
             */
            const goToHelmIndex = () => {
                $router.push({
                    name: 'helms',
                    params: {
                        projectId: projectId.value,
                        projectCode: projectCode.value
                    }
                })
            }

            const fetchReleaseDetail = async () => {
                const { clusterId, namespace, name } = $route.params
                isLoading.value = true
                const res = await $store.dispatch('helm/getReleaseDetail', {
                    $clusterId: clusterId,
                    $namespace: namespace,
                    $name: name
                })
                isLoading.value = false
                releaseDetail.value = res
            }

            onMounted(() => {
                fetchReleaseDetail()
            })

            return {
                isLoading,
                activeTab,
                releaseDetail,
                goToHelmIndex
            }
        }
    }

</script>

<style lang="postcss" scoped>
    .biz-helm-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 20px;
    }
    .basic-info {
        width: 100%;
        height: 150px;
        background-color: #fff;
        padding: 16px 24px;
        
        .info {
            display: flex;
            margin-top: 16px;
            .appName {
                width: 200px;
            }
            .info-item {
                padding-left: 88px;
            }
        }
        .desc {
            margin-top: 26px;
        }
        .key {
            font-size: 12px;
            font-weight: 400;
            color: #979ba5;
        }
    }

    .notes {
        margin-top: 20px;
        width: 100%;
        height: 300px;
        padding: 16px 24px;
        background-color: #fff;
        .notes-content {
            width: 100%;
            height: 224px;
            background-color: #f0f1f5;
            border: 1px solid #dcdee5;
            border-radius: 2px;
            overflow-y: scroll;
            padding: 16px;
            font-size: 12px;
            margin-top: 12px;
        }
    }

    .application {
        margin-top: 20px;
        width: 100%;
        padding: 16px 24px;
        background-color: #fff;
    }

    .installation {
        width: 100%;
        margin-top: 20px;
        background-color: #fff;
        .title {
            padding: 16px 24px;
        }
        .installation-editor {
            width: 100% !important;
            height: 340px !important;
        }
    }
    
    .title {
        font-size: 16px;
        color: #63656e;
        font-weight: 700;
    }
</style>
