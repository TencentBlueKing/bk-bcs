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

package httpsvr

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/utils"
)

// HttpServerClient http server client
type HttpServerClient struct {
	K8SClient client.Client
}

// NetIPAllocateRequest represents allocate BCSNetIP request
type NetIPAllocateRequest struct {
	Host         string `json:"host"`
	ContainerID  string `json:"containerID"`
	IPAddr       string `json:"ipAddr,omitempty"`
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
}

// NetIPDeleteRequest represents delete BCSNetIP request
type NetIPDeleteRequest struct {
	Host         string `json:"host"`
	ContainerID  string `json:"containerID"`
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
}

// NetIPResponse represents allocate/delete BCSNetIP response
type NetIPResponse struct {
	Code      uint        `json:"code"`
	Message   string      `json:"message"`
	Result    bool        `json:"result"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
}

func responseData(code uint, m string, result bool, reqID string, data interface{}) *NetIPResponse {
	return &NetIPResponse{
		Code:      code,
		Message:   m,
		Result:    result,
		RequestID: reqID,
		Data:      data,
	}
}

// InitRouters init router
func InitRouters(ws *restful.WebService, httpServerClient *HttpServerClient) {
	ws.Route(ws.POST("/v1/allocator").To(httpServerClient.allocateIP))
	ws.Route(ws.DELETE("/v1/allocator").To(httpServerClient.deleteIP))
}

func (c *HttpServerClient) allocateIP(request *restful.Request, response *restful.Response) {
	requestID := request.Request.Header.Get("X-Request-Id")
	netIPReq := &NetIPAllocateRequest{}
	if err := request.ReadEntity(netIPReq); err != nil {
		blog.Errorf("decode json request failed, %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := validateAllocateNetIPReq(netIPReq); err != nil {
		response.WriteEntity(responseData(1, err.Error(), false, requestID, nil))
		return
	}

	if netIPReq.IPAddr != "" {
		netIP, err := c.getIPFromRequest(netIPReq)
		if err != nil {
			response.WriteEntity(responseData(2, err.Error(), false, requestID, nil))
		}
		if err := c.K8SClient.Status().Update(context.Background(), netIP); err != nil {
			message := fmt.Sprintf("update IP [%s] status failed", netIPReq.IPAddr)
			blog.Errorf(message)
			response.WriteEntity(responseData(2, message, false, requestID, nil))
			return
		}
		message := fmt.Sprintf("allocate ip [%s] for container %s success", netIPReq.IPAddr, netIPReq.ContainerID)
		blog.Infof(message)
		response.WriteEntity(responseData(0, message, true, requestID, netIPReq))
		return
	}

	// ip address not exists in request
	netPoolList := &v1.BCSNetPoolList{}
	if err := c.K8SClient.List(context.Background(), netPoolList); err != nil {
		message := fmt.Sprintf("get BCSNetPool list failed, %s", err.Error())
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}

	availableIP, reservedIP, err := c.getAvailableIPs(netPoolList, netIPReq)
	if err != nil {
		response.WriteEntity(responseData(2, err.Error(), false, requestID, nil))
		return
	}

	// get claim info from pod annotations
	claimKey, ExpiredDuration, err := c.getIPClaimAndDuration(netIPReq.PodNamespace, netIPReq.PodName)
	if err != nil {
		message := fmt.Sprintf("check BCSNetIP [%s] fixed status failed, %s", netIPReq.IPAddr, err.Error())
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}

	// match IP by claim or podName and podNamespace
	if claimKey != "" {
		for _, ip := range reservedIP {
			if ip.Status.IPClaimKey == claimKey {
				if err := c.updateIPStatus(ip, netIPReq, claimKey, ExpiredDuration, true); err != nil {
					response.WriteEntity(responseData(2, err.Error(), false, requestID, nil))
					return
				}
				data := netIPReq
				data.IPAddr = ip.Name
				message := fmt.Sprintf("allocate IP [%s] for Host %s success", ip.Name, netIPReq.Host)
				blog.Infof(message)
				response.WriteEntity(responseData(0, message, true, requestID, data))
				return
			}
		}
		message := fmt.Sprintf("claim %s is not bounding a valid IP", claimKey)
		blog.Infof(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}

	if len(availableIP) == 0 {
		message := fmt.Sprintf("no available IP for pod %s/%s", netIPReq.PodNamespace, netIPReq.PodName)
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}
	if err := c.updateIPStatus(availableIP[0], netIPReq, "", "", false); err != nil {
		response.WriteEntity(responseData(2, err.Error(), false, requestID, nil))
		return
	}
	message := fmt.Sprintf("allocate IP [%s] for Host %s success", availableIP[0].Name, netIPReq.Host)
	blog.Infof(message)
	data := netIPReq
	data.IPAddr = availableIP[0].Name
	response.WriteEntity(responseData(0, message, true, requestID, data))
}

func (c *HttpServerClient) updateIPStatus(ip *v1.BCSNetIP, netIPReq *NetIPAllocateRequest, claimKey, duration string,
	fixed bool) error {
	ip.Status = v1.BCSNetIPStatus{
		Phase:        constant.BCSNetIPActiveStatus,
		Host:         netIPReq.Host,
		ContainerID:  netIPReq.ContainerID,
		Fixed:        fixed,
		IPClaimKey:   claimKey,
		PodNamespace: netIPReq.PodNamespace,
		PodName:      netIPReq.PodName,
		UpdateTime:   metav1.Now(),
		KeepDuration: duration,
	}
	if err := c.K8SClient.Status().Update(context.Background(), ip); err != nil {
		message := fmt.Sprintf("update IP [%s] status failed, err %s", netIPReq.IPAddr, err.Error())
		blog.Errorf(message)
		return errors.New(message)
	}
	return nil
}

func (c *HttpServerClient) getAvailableIPs(netPoolList *v1.BCSNetPoolList, netIPReq *NetIPAllocateRequest) (
	[]*v1.BCSNetIP, []*v1.BCSNetIP, error) {
	var availableIP, reservedIP []*v1.BCSNetIP
	found := false
	for _, pool := range netPoolList.Items {
		if utils.StringInSlice(pool.Spec.Hosts, netIPReq.Host) {
			found = true
			for _, v := range pool.Spec.AvailableIPs {
				netIP := &v1.BCSNetIP{}
				if err := c.K8SClient.Get(context.Background(), types.NamespacedName{Name: v}, netIP); err != nil {
					blog.Warnf("get BCSNetIP [%s] failed, %s", v, err.Error())
					continue
				}
				if netIP.Status.Phase == constant.BCSNetIPAvailableStatus {
					availableIP = append(availableIP, netIP)
				}
				if netIP.Status.Phase == constant.BCSNetIPReservedStatus {
					reservedIP = append(reservedIP, netIP)
				}
			}
		}
	}
	if !found {
		message := fmt.Sprintf("host %s does not exist in pools", netIPReq.Host)
		blog.Errorf(message)
		return nil, nil, errors.New(message)
	}

	return availableIP, reservedIP, nil
}

func (c *HttpServerClient) getIPFromRequest(netIPReq *NetIPAllocateRequest) (*v1.BCSNetIP, error) {
	netIP := &v1.BCSNetIP{}
	if err := c.K8SClient.Get(context.Background(), types.NamespacedName{Name: netIPReq.IPAddr}, netIP); err != nil {
		message := fmt.Sprintf("get BCSNetIP [%s] failed, %s", netIPReq.IPAddr, err.Error())
		blog.Errorf(message)
		return nil, errors.New(message)
	}
	if netIP.Status.Phase == constant.BCSNetIPActiveStatus {
		message := fmt.Sprintf("the requested IP [%s] is in use", netIPReq.IPAddr)
		blog.Errorf(message)
		return nil, errors.New(message)
	}
	claimKey, keepDuration, err := c.getIPClaimAndDuration(netIPReq.PodNamespace, netIPReq.PodName)
	if err != nil {
		message := fmt.Sprintf("check BCSNetIP [%s] fixed status failed, %s", netIPReq.IPAddr, err.Error())
		blog.Errorf(message)
		return nil, errors.New(message)
	}
	netIP.Status = v1.BCSNetIPStatus{
		Phase:        constant.BCSNetIPActiveStatus,
		Host:         netIPReq.Host,
		ContainerID:  netIPReq.ContainerID,
		IPClaimKey:   claimKey,
		PodNamespace: netIPReq.PodNamespace,
		PodName:      netIPReq.PodName,
		Fixed: func(s string) bool {
			if s != "" {
				return true
			}
			return false
		}(claimKey),
		UpdateTime:   metav1.Now(),
		KeepDuration: keepDuration,
	}
	return netIP, nil
}

func (c *HttpServerClient) deleteIP(request *restful.Request, response *restful.Response) {
	requestID := request.Request.Header.Get("X-Request-Id")
	netIPReq := &NetIPDeleteRequest{}
	if err := request.ReadEntity(netIPReq); err != nil {
		blog.Errorf("decode json request failed, %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := validateDeleteNetIPReq(netIPReq); err != nil {
		response.WriteEntity(responseData(1, err.Error(), false, requestID, nil))
		return
	}

	netIPList := &v1.BCSNetIPList{}
	if err := c.K8SClient.List(context.Background(), netIPList); err != nil {
		message := fmt.Sprintf("get BCSNetIP list failed, %s", err.Error())
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}
	var netIP *v1.BCSNetIP
	for _, ip := range netIPList.Items {
		if ip.Status.ContainerID == netIPReq.ContainerID && ip.Status.PodNamespace == netIPReq.PodNamespace &&
			ip.Status.PodName == netIPReq.PodName {
			netIP = &ip
			break
		}
	}
	if netIP == nil {
		message := fmt.Sprintf("didn't find related BCSNetIP instance for container %s", netIPReq.ContainerID)
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}
	claimKey, _, err := c.getIPClaimAndDuration(netIPReq.PodNamespace, netIPReq.PodName)
	if err != nil {
		message := fmt.Sprintf("check BCSNetIP [%s] fixed status failed, %s", netIP.Name, err.Error())
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}
	if claimKey != "" {
		netIP.Status.Phase = constant.BCSNetIPReservedStatus
		netIP.Status.UpdateTime = metav1.Now()
	} else {
		netIP.Status = v1.BCSNetIPStatus{
			Phase:      constant.BCSNetIPAvailableStatus,
			UpdateTime: metav1.Now(),
		}
	}

	if err := c.K8SClient.Status().Update(context.Background(), netIP); err != nil {
		message := fmt.Sprintf("update IP [%s] status failed", netIP.Name)
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}
	message := fmt.Sprintf("deactive IP [%s] success, it's available now", netIP.Name)
	blog.Errorf(message)
	response.WriteEntity(responseData(0, message, true, requestID, netIPReq))
}

func (c *HttpServerClient) getIPClaimAndDuration(namespace, name string) (string, string, error) {
	pod := &coreV1.Pod{}
	err := c.K8SClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, pod)
	if err != nil {
		return "", "", err
	}

	claimValue, ok := pod.ObjectMeta.Annotations[constant.PodAnnotationKeyForIPClaim]
	if !ok {
		return "", "", nil
	}
	claimKey := fmt.Sprintf("%s/%s", namespace, claimValue)
	claim := &v1.BCSNetIPClaim{}
	err = c.K8SClient.Get(context.Background(), types.NamespacedName{Name: claimValue, Namespace: namespace}, claim)
	if err != nil {
		return claimKey, "", err
	}
	return claimKey, claim.Spec.ExpiredDuration, nil
}

func validateAllocateNetIPReq(netIPReq *NetIPAllocateRequest) error {
	var message string
	if netIPReq == nil {
		message = "lost request body for allocating ip"
		blog.Errorf(message)
		return errors.New(message)
	}
	if netIPReq.Host == "" || netIPReq.ContainerID == "" {
		message = "lost Host/ContainerID info in request"
		blog.Errorf(message)
		return errors.New(message)
	}
	if netIPReq.PodNamespace == "" || netIPReq.PodName == "" {
		message = "lost PodNamespace/PodName info in request"
		blog.Errorf(message)
		return errors.New(message)
	}
	return nil
}

func validateDeleteNetIPReq(netIPReq *NetIPDeleteRequest) error {
	var message string
	if netIPReq == nil {
		message = "lost request body for allocating ip"
		blog.Errorf(message)
		return errors.New(message)
	}
	if netIPReq.Host == "" || netIPReq.ContainerID == "" {
		message = "lost Host/ContainerID info in request"
		blog.Errorf(message)
		return errors.New(message)
	}
	if netIPReq.PodNamespace == "" || netIPReq.PodName == "" {
		message = "lost PodNamespace/PodName info in request"
		blog.Errorf(message)
		return errors.New(message)
	}
	return nil
}
