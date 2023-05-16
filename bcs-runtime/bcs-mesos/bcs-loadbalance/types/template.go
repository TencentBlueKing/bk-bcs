/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package types

// TemplateData data holder for haproxy.cfg.template
type TemplateData struct {
	HTTP    HTTPServiceInfoList      // HTTP service info
	HTTPS   HTTPServiceInfoList      // HTTPS service info
	TCP     FourLayerServiceInfoList // TCP service info
	UDP     FourLayerServiceInfoList // UDP service info
	LogFlag bool                     // log flag, true will open log writer
	SSLCert string                   // SSL certificate path, true will listen https
}
