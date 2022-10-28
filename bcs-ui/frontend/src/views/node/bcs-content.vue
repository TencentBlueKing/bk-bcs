<template>
  <div class="bcs-body">
    <slot name="header"></slot>
    <section class="content">
      <slot></slot>
    </section>
  </div>
</template>
<style lang="postcss" scoped>
.bcs-body {
    .content {
        padding: 24px;
    }
}
</style>
