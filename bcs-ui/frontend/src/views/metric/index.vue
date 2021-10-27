<template>
    <wlist-metric></wlist-metric>
</template>

<script>
    const wlistMetric = () => import(
        /* webpackChunkName: 'metric' */'./main'
    )

    export default {
        components: {
            wlistMetric
        }
    }
</script>
