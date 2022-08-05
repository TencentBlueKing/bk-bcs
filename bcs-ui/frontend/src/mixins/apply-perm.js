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

// 功能、角色对应 map
const POLICY_ROLE_MAP = {
  // 创建 -> 创建者角色
  create: 'creator',
  // 删除 -> 拥有者角色
  delete: 'bcs_manager',
  // 列表 -> 拥有者角色
  list: 'bcs_manager',
  // 查看 -> 拥有者角色
  view: 'bcs_manager',
  // 编辑 -> 拥有者角色
  edit: 'bcs_manager',
  // 使用 -> 拥有者角色
  use: 'bcs_manager',
};

export default {
  methods: {
    createApplyPermUrl({ policy, projectCode, idx }) {
      const url = `${DEVOPS_BCS_API_URL}/api/perm/apply/subsystem/?client_id=bcs-web-backend&service_code=bcs`
                + `&project_code=${projectCode}&role_${POLICY_ROLE_MAP[policy]}=${idx}`;
      return url;
    },
  },
};
