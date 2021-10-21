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

package v4http

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
)

func (s *Scheduler) CreateAdmissionWebhook(body []byte) (string, error) {

	blog.Info("create admissionwebhook data(%s)", string(body))

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := s.GetHost() + "/v1/admissionwebhook"
	blog.Info("post a request to url(%s), request:%s", url, string(body))

	reply, err := s.client.POST(url, nil, body)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) UpdateAdmissionWebhook(body []byte) (string, error) {

	blog.Info("update admissionwebhook data(%s)", string(body))

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := s.GetHost() + "/v1/admissionwebhook"
	blog.Info("put a request to url(%s), request:%s", url, string(body))

	reply, err := s.client.PUT(url, nil, body)
	if err != nil {
		blog.Error("put request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) DeleteAdmissionWebhook(ns string, name string) (string, error) {
	blog.Info("delete admissionwebhook(%s, %s)", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := s.GetHost() + "/v1/admissionwebhook/" + ns + "/" + name
	blog.Info("delete url(%s)", url)

	reply, err := s.client.DELETE(url, nil, nil)
	if err != nil {
		blog.Error("delete url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) FetchAdmissionWebhook(ns string, name string) (string, error) {
	blog.Info("fetch admissionwebhook(%s, %s)", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := s.GetHost() + "/v1/admissionwebhook/" + ns + "/" + name
	blog.Info("fetch url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("fetch url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) FetchAllAdmissionWebhooks() (string, error) {
	blog.Info("fetch all admissionwebhooks")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := s.GetHost() + "/v1/admissionwebhooks"
	blog.Info("fetch url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("fetch url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}
