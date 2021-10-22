/**
 * @file app router 配置
 */

const App = () => import(/* webpackChunkName: 'app-entry' */'@open/views/app')

const childRoutes = [
    // domain/bcs/projectId/app 应用页面
    {
        path: ':projectCode/app',
        component: App,
        children: [
            // k8s deployments 应用
            {
                path: 'deployments',
                name: 'deployments',
                children: [
                    // k8s deployments 应用里的实例详情页面
                    {
                        path: ':instanceId',
                        name: 'deploymentsInstanceDetail'
                    },
                    {
                        path: ':instanceName/:instanceNamespace/:instanceCategory',
                        name: 'deploymentsInstanceDetail2'
                    },
                    // k8s deployments 应用里的容器详情页面
                    {
                        path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
                        name: 'deploymentsContainerDetail'
                    },
                    {
                        path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
                        name: 'deploymentsContainerDetail2'
                    },
                    // k8s deployments 应用里的应用实例化页面
                    {
                        path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
                        name: 'deploymentsInstantiation'
                    }
                ]
            },
            // k8s daemonset 应用
            {
                path: 'daemonset',
                name: 'daemonset',
                children: [
                    // k8s daemonset 应用里的实例详情页面
                    {
                        path: ':instanceId',
                        name: 'daemonsetInstanceDetail'
                    },
                    {
                        path: ':instanceName/:instanceNamespace/:instanceCategory',
                        name: 'daemonsetInstanceDetail2'
                    },
                    // k8s daemonset 应用里的容器详情页面
                    {
                        path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
                        name: 'daemonsetContainerDetail'
                    },
                    {
                        path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
                        name: 'daemonsetContainerDetail2'
                    },
                    // k8s daemonset 应用里的应用实例化页面
                    {
                        path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
                        name: 'daemonsetInstantiation'
                    }
                ]
            },
            // k8s job 应用
            {
                path: 'job',
                name: 'job',
                children: [
                    // k8s job 应用里的实例详情页面
                    {
                        path: ':instanceId',
                        name: 'jobInstanceDetail'
                    },
                    {
                        path: ':instanceName/:instanceNamespace/:instanceCategory',
                        name: 'jobInstanceDetail2'
                    },
                    // k8s job 应用里的容器详情页面
                    {
                        path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
                        name: 'jobContainerDetail'
                    },
                    {
                        path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
                        name: 'jobContainerDetail2'
                    },
                    // k8s job 应用里的应用实例化页面
                    {
                        path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
                        name: 'jobInstantiation'
                    }
                ]
            },
            // k8s statefulset 应用
            {
                path: 'statefulset',
                name: 'statefulset',
                children: [
                    // k8s statefulset 应用里的实例详情页面
                    {
                        path: ':instanceId',
                        name: 'statefulsetInstanceDetail'
                    },
                    {
                        path: ':instanceName/:instanceNamespace/:instanceCategory',
                        name: 'statefulsetInstanceDetail2'
                    },
                    // k8s statefulset 应用里的容器详情页面
                    {
                        path: ':instanceId/taskgroups/:taskgroupName/containers/:containerId',
                        name: 'statefulsetContainerDetail'
                    },
                    {
                        path: ':instanceName/:instanceNamespace/:instanceCategory/taskgroups/:taskgroupName/containers/:containerId',
                        name: 'statefulsetContainerDetail2'
                    },
                    // k8s statefulset 应用里的应用实例化页面
                    {
                        path: ':templateId/instantiation/:category/:tmplAppName/:tmplAppId',
                        name: 'statefulsetInstantiation'
                    }
                ]
            },
            // k8s gamestatefulset 应用
            {
                path: 'gamestatefulset',
                name: 'gamestatefulset'
            },
            // k8s gamedeployments 应用
            {
                path: 'gamedeployments',
                name: 'gamedeployments'
            },
            // k8s gamestatefulset 应用
            {
                path: 'customobjects',
                name: 'customobjects'
            }
        ]
    }
]

export default childRoutes
