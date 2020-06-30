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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
)

func (s *Scheduler) RegisterCustomResource(body []byte) (string, error) {

	blog.Info("Register custom resource data(%s)", string(body))

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := s.GetHost() + "/v1/crr/register"
	blog.Info("post a request to url(%s), request:%s", url, string(body))

	reply, err := s.client.POST(url, nil, body)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) CreateCustomResource(ns, kind string, body []byte) (string, error) {

	blog.Info("Create custom resource(%s %s) data(%s)", ns, kind, string(body))

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/crd/namespaces/%s/%s", s.GetHost(), ns, kind)
	blog.Info("post a request to url(%s), request:%s", url, string(body))

	reply, err := s.client.POST(url, nil, body)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) UpdateCustomResource(ns, kind string, body []byte) (string, error) {

	blog.Info("Update custom resource(%s %s) data(%s)", ns, kind, string(body))

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/crd/namespaces/%s/%s", s.GetHost(), ns, kind)
	blog.Info("post a request to url(%s), request:%s", url, string(body))

	reply, err := s.client.PUT(url, nil, body)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) DeleteCustomResource(ns, kind, name string) (string, error) {

	blog.Info("Delete custom resource(%s %s %s)", ns, kind, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/crd/namespaces/%s/%s/%s", s.GetHost(), ns, kind, name)
	blog.Info("post a request to url(%s), request:%s", url)

	reply, err := s.client.DELETE(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) ListCustomResource(ns, kind string) (string, error) {

	blog.Info("List custom resource(%s %s)", ns, kind)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/crd/namespaces/%s/%s", s.GetHost(), ns, kind)
	blog.Info("post a request to url(%s), request:%s", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) ListAllCustomResource(kind string) (string, error) {

	blog.Info("List all custom resource(%s)", kind)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/crd/%s", s.GetHost(), kind)
	blog.Info("post a request to url(%s), request:%s", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}

func (s *Scheduler) GetCustomResource(ns, kind, name string) (string, error) {

	blog.Info("Get custom resource(%s %s %s)", ns, kind, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		return err.Error(), err
	}

	url := fmt.Sprintf("%s/v1/crd/namespaces/%s/%s/%s", s.GetHost(), ns, kind, name)
	blog.Info("post a request to url(%s), request:%s", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("post request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		return err.Error(), err
	}

	return string(reply), nil
}
