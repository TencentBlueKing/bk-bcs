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

const Depot = () => import(/* webpackChunkName: 'depot' */'@/views/depot');
const ImageLibrary = () => import(/* webpackChunkName: 'depot' */'@/views/depot/image-library');
const ImageDetail = () => import(/* webpackChunkName: 'depot' */'@/views/depot/image-detail');
const ProjectImage = () => import(/* webpackChunkName: 'depot' */'@/views/depot/project-image');

const childRoutes = [
  // 这里没有把 depot 作为 cluster 的 children
  // 是因为如果把 depot 作为 cluster 的 children，那么必须要在 Cluster 的 component 中
  // 通过 router-view 来渲染子组件，但在业务逻辑中，depot 和 cluster 是平级的
  {
    path: ':projectCode/depot',
    name: 'depotMain',
    component: Depot,
    children: [
      // domain/bcs/projectCode/depot => domain/bcs/projectCode/depot/image-library
      {
        path: 'image-library',
        component: ImageLibrary,
        name: 'imageLibrary',
        alias: '',
      },
      {
        path: 'image-detail/:imageRepo',
        component: ImageDetail,
        name: 'imageDetail',
        alias: '',
        props: true,
        meta: {
          menuId: 'imageLibrary',
        },
      },
      {
        path: 'project-image',
        name: 'projectImage',
        component: ProjectImage,
      },
    ],
  },
];

export default childRoutes;
