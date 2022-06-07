const tencentCloud = () => import(/* webpackChunkName: 'cloud-token' */'@/views/cloudtoken/tencentCloud.vue')
export default [
    {
        path: ':projectCode/tencentCloud',
        name: 'tencentCloud',
        component: tencentCloud,
        meta: {
            title: 'Tencent Cloud'
        }
    }
]
