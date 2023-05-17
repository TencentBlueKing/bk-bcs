<template>
  <div class="flex justify-center relative top-[20%]">
    <div class="bg-[#fff] px-[76px] pt-[64px] pb-[40px] max-w-[640px]" v-if="$INTERNAL">
      <div class="text-[#313238] text-[28px] font-bold flex justify-center">{{ $t('欢迎使用 TKEx-IEG 容器平台') }}</div>
      <div class="text-[16px] mt-[24px] leading-[24px]">
        {{ $t('TKEx-IEG 是公司 K8S Oteam 指定的 IEG 容器服务平台。') }}
        <br />
        {{ $t('TKEx-IEG 依托于 TKE，聚焦打造游戏一站式微服务部署治理方案，并基于 IEG 各类游戏的特点，提供游戏特色的容器管理服务。') }}
      </div>
      <div class="flex justify-center mt-[32px]">
        <bk-button
          theme="primary"
          v-authority="{
            actionId: 'project_create',
            permCtx: {}
          }"
          @click="handleGotoProjectManage">{{ $t('创建项目') }}</bk-button>
        <bk-button @click="handleGotoIAM">{{ $t('加入项目') }}</bk-button>
      </div>
      <div class="flex justify-center mt-[16px]">
        <bk-button text @click="handleGotoDoc">
          <span class="relative top-[-1px]">
            <i class="bcs-icon bcs-icon-question-circle"></i>
          </span>
          {{ $t('游戏接入指南') }}
        </bk-button>
      </div>
    </div>
    <div class="bg-[#fff] px-[76px] pt-[64px] pb-[40px] max-w-[640px]" v-else>
      <div class="text-[#313238] text-[28px] font-bold flex justify-center">
        {{ $t('欢迎使用蓝鲸容器管理平台') }}
      </div>
      <div class="text-[16px] mt-[24px] leading-[24px]">
        {{ $t('蓝鲸容器管理平台（Blueking Container Service）定位于打造云原生技术和业务实际应用场景之间的桥梁；') }}
        <br />
        {{ $t('聚焦于复杂应用场景的容器化部署技术方案的研发、整合和产品化；') }}
        <br />
        {{ $t('致力于为游戏等复杂应用提供一站式、低门槛的容器编排和服务治理服务。') }}
      </div>
      <div class="flex justify-center mt-[32px]">
        <bk-button
          theme="primary"
          v-authority="{
            actionId: 'project_create',
            permCtx: {}
          }"
          @click="handleGotoProjectManage">{{ $t('创建项目') }}</bk-button>
        <bk-button @click="handleGotoIAM">{{ $t('加入项目') }}</bk-button>
      </div>
      <div class="flex justify-center mt-[16px]">
        <bk-button text @click="handleGotoDoc">
          <span class="relative top-[-1px]">
            <i class="bcs-icon bcs-icon-question-circle"></i>
          </span>
          {{ $t('游戏接入指南') }}
        </bk-button>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, computed, toRef, reactive } from 'vue';
import $router from '@/router';

export default defineComponent({
  name: 'ProjectGuide',
  setup() {
    const $route = computed(() => toRef(reactive($router), 'currentRoute').value);

    function handleGotoIAM() {
      window.open(`${window.BK_IAM_APP_URL}apply-join-user-group?system_id=bk_bcs_app`);
    }
    function handleGotoProjectManage() {
      if (window.REGION === 'ieod') {
        window.open(`${window.DEVOPS_HOST}/console/pm`);
      } else {
        if ($route.value.name === 'projectManage') return;
        $router.push({
          name: 'projectManage',
        });
      }
    }
    function handleGotoDoc() {
      window.open(window.BCS_CONFIG?.help);
    }

    return {
      handleGotoIAM,
      handleGotoProjectManage,
      handleGotoDoc,
    };
  },
});
</script>
