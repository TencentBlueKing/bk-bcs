import { defineComponent, computed } from '@vue/composition-api'

import './dashboard-top-actions.css'

export default defineComponent({
    name: 'DashboardTopActions',
    setup (props, ctx) {
        const { $router, $route } = ctx.root

        const projectId = computed(() => $route.params.projectId).value
        const projectCode = computed(() => $route.params.projectCode).value

        const goBCS = () => {
            $router.push({
                name: 'clusterMain',
                params: {
                    projectId: projectId,
                    projectCode: projectCode
                }
            })
        }

        return {
            goBCS
        }
    },
    render () {
        return (
            <div class="dashboard-top-actions">
                {
                    this.$INTERNAL
                        ? <a href={this.PROJECT_CONFIG.doc.contact} class="bk-text-button" v-bk-tooltips_top={this.$t('蓝鲸容器助手')}>{this.$t('联系我们')}</a>
                        : null
                }
                <a href={this.PROJECT_CONFIG.doc.help} target="_blank" class="bk-text-button">{this.$t('帮助')}</a>
            </div>
        )
    }
})
