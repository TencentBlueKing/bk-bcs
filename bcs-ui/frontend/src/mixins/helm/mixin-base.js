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
/* eslint-disable no-prototype-builtins */
import bkKeyer from '@/components/keyer';
import ace from '@/components/ace-editor';
import bkFormCreater from '@/views/helm/components/form-creater';

export default {
  components: {
    bkKeyer,
    ace,
    bkFormCreater,
  },
  data() {
    return {
      curProjectId: '',
    };
  },
  computed: {
    curProject() {
      const project = this.$store.state.curProject;
      return project;
    },
    projectId() {
      this.curProjectId = this.$route.params.projectId;
      return this.curProjectId;
    },
  },
  watch: {
    // 如果不是k8s类型的项目，无法访问些页面，重定向回集群首页
    curProjectId() {
      if (this.curProject && this.curProject.kind !== PROJECT_K8S) {
        this.$router.push({
          name: 'clusterMain',
          params: {
            projectId: this.projectId,
            projectCode: this.projectCode,
          },
        });
      }
    },
  },
  methods: {
    /**
         * 简单判断是否为图片
         * @param  {string}  img 图片url
         * @return {boolean} true/false
         */
    isImage(img) {
      if (!img) {
        return false;
      }
      if (img.startsWith('http://') || img.startsWith('https://') || img.startsWith('data:image/')) {
        return true;
      }
      return false;
    },

    /**
         * 根据path（eg: a.b.c）设置对象属性
         * @param {object} obj 对象
         * @param {string} path  路径
         * @param {string number...} value 值
         */
    setProperty(obj, path, value) {
      const paths = path.split('.');
      let temp = obj;
      const pathLength = paths.length;
      if (pathLength) {
        for (let i = 0; i < pathLength; i++) {
          const item = paths[i];
          if (temp.hasOwnProperty(item)) {
            if (i === (pathLength - 1)) {
              temp[item] = value;
            } else {
              temp = temp[item];
            }
          }
        }
      }
    },

    /**
         * 根据path（eg: a.b.c）判断对象属性是否存在
         * @param {object} obj 对象
         * @param {string} path  路径
         * @return {boolean} true/false
         */
    hasProperty(obj, path) {
      const paths = path.split('.');
      let temp = obj;
      const pathLength = paths.length;
      if (pathLength) {
        for (let i = 0; i < pathLength; i++) {
          const item = paths[i];
          if (temp.hasOwnProperty(item)) {
            temp = temp[item];
          } else {
            return false;
          }
        }
        return true;
      }
      return false;
    },

    /**
         * 根据path（eg: a.b.c）获取对象属性
         * @param {object} obj 对象
         * @param {string} path  路径
         * @return {string number...} value
         */
    getProperty(obj, path) {
      const paths = path.split('.');
      let temp = obj;
      if (paths.length) {
        for (const item of paths) {
          if (temp.hasOwnProperty(item)) {
            temp = temp[item];
          } else {
            return undefined;
          }
        }
        return temp;
      }
      return undefined;
    },

    /**
         * 返回
         */
    goBack() {
      window.history.go(-1);
    },
  },
};
