/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/

export default {
  data() {
    return {
      isBatchRemoving: false,
      curSelectedData: [],
      batchDialogConfig: {
        isShow: false,
        list: [],
        data: [],
      },
    };
  },
  computed: {
    isCheckCurPageAll() {
      if (this.curPageData.length) {
        const list = this.curPageData;
        const selectList = list.filter(item => item.isChecked === true);
        const canSelectList = list.filter(item => item.can_delete);
        if (selectList.length && (selectList.length === canSelectList.length)) {
          return true;
        }
        return false;
      }
      return false;
    },
  },
  methods: {

    /**
         * 每行的多选框点击事件
         */
    rowClick() {
      this.$nextTick(() => {
        this.alreadySelectedNums = this.serviceList.filter(item => item.isChecked).length;
      });
    },

    /**
         * 选择当前页数据
         */
    selectServices() {
      const list = this.curPageData;
      const selectList = list.filter(item => item.isChecked === true);
      this.curSelectedData.splice(0, this.curSelectedData.length, ...selectList);
    },

    /**
         * 清空当前页选择
         */
    clearSelectServices() {
      this.serviceList.forEach((item) => {
        item.isChecked = false;
      });
    },

    /**
         * 确认批量删除service
         */
    async removeServices() {
      const data = [];
      const names = [];

      this.serviceSelectedList.forEach((item) => {
        data.push({
          cluster_id: item.clusterId,
          namespace: item.namespace,
          name: item.resourceName,
        });
        names.push(`${item.cluster_id} / ${item.namespace} / ${item.resourceName}`);
      });
      if (!data.length) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.msg.selectDeleteService'),
        });
        return false;
      }

      this.batchDialogConfig.list = names;
      this.batchDialogConfig.data = data;
      this.batchDialogConfig.isShow = true;
    },

    /**
         * 批量删除service
         * @param  {object} data services
         */
    async deleteServices(data) {
      this.batchDialogConfig.isShow = false;
      this.isPageLoading = true;
      const { projectId } = this;

      try {
        await this.$store.dispatch('network/deleteServices', {
          projectId,
          data,
        });

        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.delete'),
        });
        this.initPageConf();
        this.getServiceList();
      } catch (e) {
        // 4004，已经被删除过，但接口不能立即清除，防止重复删除
        if (e.code === 4004) {
          this.initPageConf();
          this.getServiceList();
        }
        this.$bkMessage({
          theme: 'error',
          delay: 8000,
          hasCloseIcon: true,
          message: e.message,
        });
        this.isPageLoading = false;
      }
    },
  },
};
