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
 */

// Package httpsvr is http server package
// NOCC:tosa/comment_ratio(设计如此)
package httpsvr

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/emicklei/go-restful"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

// NetIPAllocateReponseData represents allocate BCSNetIP response
type NetIPAllocateReponseData struct {
	Host         string `json:"host"`
	ContainerID  string `json:"containerID"`
	IPAddr       string `json:"ipAddr"`
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
	Gateway      string `json:"gateway"`
	Mask         int    `json:"mask"`
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

// get allocate response data object from request, BCSNetIP and BCSNetPool
func getAllocateResponseData(
	req *NetIPAllocateRequest, bcsip *v1.BCSNetIP, bcspool *v1.BCSNetPool) *NetIPAllocateReponseData {
	return &NetIPAllocateReponseData{
		Host:         req.Host,
		ContainerID:  req.Host,
		IPAddr:       bcsip.Name,
		PodName:      req.Host,
		PodNamespace: req.Host,
		Gateway:      bcspool.Spec.Gateway,
		Mask:         bcspool.Spec.Mask,
	}
}

// get common response data
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
	ws.Route(ws.POST("/v1/allocator").To(httpServerClient.AllocateIP))
	ws.Route(ws.DELETE("/v1/allocator").To(httpServerClient.DeleteIP))
}

// AllocateIP do ip allocation
// nolint funlen
func (c *HttpServerClient) AllocateIP(request *restful.Request, response *restful.Response) {
	requestID := request.Request.Header.Get("X-Request-Id")
	netIPReq := &NetIPAllocateRequest{}
	// ReadEntity checks the Accept header and reads the content into the entityPointer.
	if err := request.ReadEntity(netIPReq); err != nil {
		blog.Errorf("decode json request failed, %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := validateAllocateNetIPReq(netIPReq); err != nil {
		response.WriteEntity(responseData(1, err.Error(), false, requestID, nil))
		return
	}

	// get available ips
	netPoolList := &v1.BCSNetPoolList{}
	if err := c.K8SClient.List(context.Background(), netPoolList); err != nil {
		message := fmt.Sprintf("get BCSNetPool list failed, %s", err.Error())
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}
	availableIP, err := c.getAvailableIPs(netPoolList, netIPReq)
	if err != nil {
		response.WriteEntity(responseData(2, err.Error(), false, requestID, nil))
		return
	}

	// get claim info from pod annotations
	claimName, _, err := c.getIPClaimAndDuration(netIPReq.PodNamespace, netIPReq.PodName)
	if err != nil {
		message := fmt.Sprintf("check BCSNetIP [%s] fixed status failed, %s", netIPReq.IPAddr, err.Error())
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil))
		return
	}

	// match IP by claim or podName and podNamespace
	if claimName != "" {
		ipClaim, gerr := c.getClaim(netIPReq.PodNamespace, claimName)
		if gerr != nil {
			message := fmt.Sprintf("get BCSNetIPClaim %s/%s failed, err %s",
				netIPReq.PodNamespace, claimName, gerr.Error())
			blog.Errorf(message)
			response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
			return
		}
		if ipClaim.DeletionTimestamp != nil {
			message := fmt.Sprintf("BCSNetIPClaim %s/%s is deleting", netIPReq.PodNamespace, claimName)
			blog.Errorf(message)
			response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
			return
		}
		if ipClaim.Status.Phase == "" {
			message := fmt.Sprintf("BCSNetIPClaim %s/%s empty phase, wait to pending", netIPReq.PodNamespace, claimName)
			blog.Errorf(message)
			response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
			return
		}
		if ipClaim.Status.Phase == constant.BCSNetIPClaimExpiredStatus {
			message := fmt.Sprintf("BCSNetIPClaim %s/%s is expired", netIPReq.PodNamespace, claimName)
			blog.Errorf(message)
			response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
			return
		}
		if ipClaim.Status.Phase == constant.BCSNetIPClaimPendingStatus {
			// allocate ip for pending ip claim
			targetIP, bcspool, aerr := c.allocateNewIPForClaim(netIPReq, availableIP, ipClaim)
			if aerr != nil {
				response.WriteEntity(responseData(2, aerr.Error(), false, requestID, nil)) // nolint
				return
			}
			data := getAllocateResponseData(netIPReq, targetIP, bcspool)
			response.WriteEntity(responseData(0, "success", true, requestID, data))
			return
		} else if ipClaim.Status.Phase == constant.BCSNetIPClaimBoundedStatus {
			// get ip from ip claim
			bcsNetIP, bcsNetPool, aerr := c.allocateIPByClaim(netIPReq, ipClaim)
			if aerr != nil {
				response.WriteEntity(responseData(2, aerr.Error(), false, requestID, nil)) // nolint
				return
			}
			data := getAllocateResponseData(netIPReq, bcsNetIP, bcsNetPool)
			message := fmt.Sprintf("allocate IP [%s] from BCSNetIPClaim %s/%s for Host %s success",
				bcsNetIP.Name, ipClaim.GetNamespace(), ipClaim.GetName(), netIPReq.Host)
			blog.Infof(message)
			response.WriteEntity(responseData(0, message, true, requestID, data)) // nolint
			return
		}
		message := fmt.Sprintf("invalid available BCSNetIPClaim %s/%s phase %s",
			ipClaim.GetNamespace(), ipClaim.GetName(), ipClaim.Status.Phase)
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
		return
	}
	// allocate available unfixed ip
	if len(availableIP) == 0 {
		message := fmt.Sprintf("no available IP for pod %s/%s", netIPReq.PodNamespace, netIPReq.PodName)
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
		return
	}
	targetIP := availableIP[0]
	if uerr := c.updateIPStatus(targetIP, netIPReq, "", "", false); uerr != nil {
		response.WriteEntity(responseData(2, uerr.Error(), false, requestID, nil)) // nolint
		return
	}
	message := fmt.Sprintf("allocate IP [%s] for Host %s success", targetIP.Name, netIPReq.Host)
	blog.Infof(message)
	bcspool, err := c.getPoolByIP(targetIP)
	if err != nil {
		response.WriteEntity(responseData(2, err.Error(), false, requestID, nil)) // nolint
		return
	}
	data := getAllocateResponseData(netIPReq, targetIP, bcspool)
	response.WriteEntity(responseData(0, message, true, requestID, data)) // nolint
}

// allocateNewIPForClaim xxx
func (c *HttpServerClient) allocateNewIPForClaim(
	netIPReq *NetIPAllocateRequest, availableIP []*v1.BCSNetIP, ipClaim *v1.BCSNetIPClaim) (
	*v1.BCSNetIP, *v1.BCSNetPool, error) {
	// do fixed ip bound
	if len(availableIP) == 0 {
		message := fmt.Sprintf("no available IP for pod %s/%s", netIPReq.PodNamespace, netIPReq.PodName)
		blog.Errorf(message)
		return nil, nil, fmt.Errorf(message)
	}
	// update ip status
	targetIP := availableIP[0]
	if err := c.updateIPStatus(targetIP, netIPReq,
		utils.GetNamespacedNameKey(netIPReq.PodNamespace, ipClaim.GetName()), "", true); err != nil {
		message := fmt.Sprintf("update IP %s status, failed, err %s", targetIP.GetName(), err.Error())
		blog.Errorf(message)
		return nil, nil, fmt.Errorf(message)
	}
	// bound ip claim
	if err := c.boundClaimIP(ipClaim, targetIP); err != nil {
		message := fmt.Sprintf("bound BCSNetIP %s to BCSNetIPClaim %s/%s failed, err %s",
			targetIP.GetName(), ipClaim.GetNamespace(), ipClaim.GetName(), err.Error())
		blog.Errorf(message)
		return nil, nil, fmt.Errorf(message)
	}
	message := fmt.Sprintf("allocate fixed IP [%s] for Host %s success", targetIP.Name, netIPReq.Host)
	blog.Infof(message)
	// get pool by ip
	bcspool, err := c.getPoolByIP(targetIP)
	if err != nil {
		message := fmt.Sprintf("get pool failed, err %s", err.Error())
		blog.Errorf(message)
		return nil, nil, fmt.Errorf(message)
	}
	return targetIP, bcspool, nil
}

// allocateIPByClaim xxx
func (c *HttpServerClient) allocateIPByClaim(
	netIPReq *NetIPAllocateRequest, ipClaim *v1.BCSNetIPClaim) (*v1.BCSNetIP, *v1.BCSNetPool, error) {
	ipName := ipClaim.Status.BoundedIP
	bcsNetIP, bcsNetPool, err := c.getIPAndPool(ipName)
	if err != nil {
		message := fmt.Sprintf("get BCSNetIP and BCSNetPool by BCSNetIP name %s failed, err %s",
			ipName, err.Error())
		blog.Errorf(message)
		return nil, nil, fmt.Errorf(message)
	}
	if bcsNetIP.Status.Phase != constant.BCSNetIPReservedStatus {
		message := fmt.Sprintf(
			"BCSNetIP %s bound with BCSNetIPClaim %s/%s is not in reserved status, BCSNetIP status %v",
			bcsNetIP.Name, ipClaim.Name, ipClaim.Namespace, bcsNetIP.Status)
		blog.Errorf(message)
		return nil, nil, fmt.Errorf(message)
	}
	// update ip status
	if err := c.updateIPStatus(
		bcsNetIP, netIPReq, utils.GetNamespacedNameKey(ipClaim.GetNamespace(),
			ipClaim.GetName()), ipClaim.Spec.ExpiredDuration, true); err != nil {
		message := fmt.Sprintf("update BCSNetIP %s status failed, err %s",
			ipName, err.Error())
		blog.Errorf(message)
		return nil, nil, fmt.Errorf(message)
	}
	return bcsNetIP, bcsNetPool, nil
}

// update IP Status
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

// get Available IPs
func (c *HttpServerClient) getAvailableIPs(netPoolList *v1.BCSNetPoolList, netIPReq *NetIPAllocateRequest) (
	[]*v1.BCSNetIP, error) {
	var availableIP []*v1.BCSNetIP
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
			}
		}
	}
	// if host not found in pools, return error
	if !found {
		message := fmt.Sprintf("host %s does not exist in pools", netIPReq.Host)
		blog.Errorf(message)
		return nil, errors.New(message)
	}

	return availableIP, nil
}

// get Pool By IP
func (c *HttpServerClient) getPoolByIP(bcsip *v1.BCSNetIP) (*v1.BCSNetPool, error) {
	if len(bcsip.GetLabels()) == 0 {
		return nil, fmt.Errorf("BCSNetIP %s has no labels", bcsip.GetName())
	}
	poolName, ok := bcsip.GetLabels()[constant.PodLabelKeyForPool]
	if !ok {
		return nil, fmt.Errorf("BCSNetIP %s has no pool label", bcsip.GetName())
	}
	netPool := &v1.BCSNetPool{}
	if err := c.K8SClient.Get(context.Background(), types.NamespacedName{Name: poolName}, netPool); err != nil {
		blog.Warnf("get BCSNetPool [%s] failed, %s", poolName, err.Error())
		return nil, err
	}
	return netPool, nil
}

// get Claim
func (c *HttpServerClient) getClaim(ns, name string) (*v1.BCSNetIPClaim, error) {
	retClaim := &v1.BCSNetIPClaim{}
	if err := c.K8SClient.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, retClaim); err != nil {
		blog.Warnf("get BCSNetIPClaim by ns %s name %s failed, err %s", ns, name, err.Error())
		return nil, fmt.Errorf("get BCSNetIPClaim by ns %s name %s failed, err %s", ns, name, err.Error())
	}
	return retClaim, nil
}

// bound Claim IP
func (c *HttpServerClient) boundClaimIP(claim *v1.BCSNetIPClaim, netIP *v1.BCSNetIP) error {
	claim.Status.BoundedIP = netIP.Name
	claim.Status.Phase = constant.BCSNetIPClaimBoundedStatus
	if err := c.K8SClient.Status().Update(context.Background(), claim); err != nil {
		blog.Errorf("update BCSNetIPClaim %s/%s status failed, err %v", claim.Namespace, claim.Name, err)
		return fmt.Errorf("update BCSNetIPClaim %s/%s status failed, err %v", claim.Namespace, claim.Name, err)
	}
	return nil
}

// DeleteIP do ip release
func (c *HttpServerClient) DeleteIP(request *restful.Request, response *restful.Response) {
	requestID := request.Request.Header.Get("X-Request-Id")
	netIPReq := &NetIPDeleteRequest{}
	if err := request.ReadEntity(netIPReq); err != nil {
		blog.Errorf("decode json request failed, %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error()) // nolint
		return
	}
	if err := validateDeleteNetIPReq(netIPReq); err != nil {
		response.WriteEntity(responseData(1, err.Error(), false, requestID, nil)) // nolint
		return
	}

	// list all BCSNetIP
	netIPList := &v1.BCSNetIPList{}
	if err := c.K8SClient.List(context.Background(), netIPList); err != nil {
		message := fmt.Sprintf("get BCSNetIP list failed, %s", err.Error())
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
		return
	}
	var netIP *v1.BCSNetIP
	for _, ip := range netIPList.Items {
		if ip.Status.ContainerID == netIPReq.ContainerID && ip.Status.PodNamespace == netIPReq.PodNamespace &&
			ip.Status.PodName == netIPReq.PodName && ip.Status.Host == netIPReq.Host {
			netIP = &ip
			break
		}
	}
	if netIP == nil {
		message := fmt.Sprintf("didn't find related BCSNetIP instance for container %s", netIPReq.ContainerID)
		blog.Errorf(message)
		response.WriteEntity(responseData(0, message, true, requestID, nil)) // nolint
		return
	}
	claimKey := netIP.Status.IPClaimKey
	if len(claimKey) != 0 {
		claim := &v1.BCSNetIPClaim{}
		// ParseNamespacedNameKey return key by namespace and name
		podNamespace, podName, err := utils.ParseNamespacedNameKey(claimKey)
		if err != nil {
			message := fmt.Sprintf("invalid IPClaimKey %s of BCSNetIP %s instance", claimKey, netIP.GetName())
			blog.Errorf(message)
			response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
			return
		}
		err = c.K8SClient.Get(context.Background(), types.NamespacedName{Name: podName, Namespace: podNamespace}, claim)
		if err != nil {
			message := fmt.Sprintf("get IPClaim by IPClaimKey %s of BCSNetIP %s instance", claimKey, netIP.GetName())
			blog.Errorf(message)
			response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
			return
		}
		netIP.Status = v1.BCSNetIPStatus{
			Phase:      constant.BCSNetIPReservedStatus,
			IPClaimKey: claimKey,
			UpdateTime: metav1.Now(),
		}
	} else {
		netIP.Status = v1.BCSNetIPStatus{
			Phase:      constant.BCSNetIPAvailableStatus,
			UpdateTime: metav1.Now(),
		}
	}

	// update BCSNetIP status
	if err := c.K8SClient.Status().Update(context.Background(), netIP); err != nil {
		message := fmt.Sprintf("update IP [%s] status failed", netIP.Name)
		blog.Errorf(message)
		response.WriteEntity(responseData(2, message, false, requestID, nil)) // nolint
		return
	}
	message := fmt.Sprintf("deactive IP [%s] success, it's available now", netIP.Name)
	blog.Infof(message)
	response.WriteEntity(responseData(0, message, true, requestID, netIPReq)) // nolint
}

// get IP And Pool
func (c *HttpServerClient) getIPAndPool(ip string) (*v1.BCSNetIP, *v1.BCSNetPool, error) {
	bcsNetIP := &v1.BCSNetIP{}
	if err := c.K8SClient.Get(context.Background(), types.NamespacedName{Name: ip}, bcsNetIP); err != nil {
		return nil, nil, err
	}
	bcsNetPool, err := c.getPoolByIP(bcsNetIP)
	if err != nil {
		return nil, nil, err
	}
	return bcsNetIP, bcsNetPool, nil
}

// get IP Claim And Duration
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
	claim := &v1.BCSNetIPClaim{}
	err = c.K8SClient.Get(context.Background(), types.NamespacedName{Name: claimValue, Namespace: namespace}, claim)
	if err != nil {
		return claimValue, "", err
	}
	return claimValue, claim.Spec.ExpiredDuration, nil
}

// validate Allocate Net IP Req
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

// validate Delete Net IP Req
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
