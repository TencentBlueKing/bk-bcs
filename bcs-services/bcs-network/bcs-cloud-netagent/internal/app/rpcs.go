/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/cloudnetagent"
	pbcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/common"
)

func (s *Server) AllocIP(ctx context.Context, req *pb.AllocIPReq) (*pb.AllocIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("AllocIP seq[%d] input[%+v]", req.Seq, req)

	

}

func (s *Server) ReleaseIP(context.Context, req *pb.ReleaseIPReq) (*pb.ReleaseIPResp, error) {

}