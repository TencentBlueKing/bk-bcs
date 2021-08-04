// Copyright 2021 The BCS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scalercore

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/uuid"

	autoscalingv1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-general-pod-autoscaler/pkg/requests"
)

var client = http.Client{
	Timeout: 15 * time.Second,
}

var _ Scaler = &WebhookScaler{}

type WebhookScaler struct {
	modeConfig *autoscalingv1.WebhookMode
	name       string
}

func NewWebhookScaler(modeConfig *autoscalingv1.WebhookMode) Scaler {
	return &WebhookScaler{modeConfig: modeConfig, name: Webhook}
}

func (s *WebhookScaler) GetReplicas(gpa *autoscalingv1.GeneralPodAutoscaler, currentReplicas int32) (int32, error) {
	if s.modeConfig == nil {
		return 0, errors.New("webhookPolicy parameter must not be nil")
	}

	u, err := s.buildURLFromWebhookPolicy()
	if err != nil {
		return 0, err
	}
	req := requests.AutoscaleReview{
		Request: &requests.AutoscaleRequest{
			UID:  uuid.NewUUID(),
			Name: gpa.Spec.ScaleTargetRef.Name,
			// gpa and workload must deploy in the same namespace
			Namespace:       gpa.Namespace,
			Parameters:      s.modeConfig.Parameters,
			CurrentReplicas: currentReplicas,
		},
		Response: nil,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}

	res, err := client.Post(
		u.String(),
		"application/json",
		strings.NewReader(string(b)),
	)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			if err != nil {
				err = errors.Wrap(err, cerr.Error())
			} else {
				err = cerr
			}
		}
	}()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status code %d from the server: %s", res.StatusCode, u.String())
	}
	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var faResp requests.AutoscaleReview
	err = json.Unmarshal(result, &faResp)
	if err != nil {
		return 0, err
	}
	if faResp.Response == nil {
		return 0, fmt.Errorf("received empty reponse")
	}
	if faResp.Response.Scale {
		return faResp.Response.Replicas, nil
	}
	return currentReplicas, nil

}

func (s *WebhookScaler) ScalerName() string {
	return s.name
}

// buildURLFromWebhookPolicy - build URL for Webhook and set CARoot for client Transport
func (s *WebhookScaler) buildURLFromWebhookPolicy() (u *url.URL, err error) {
	w := s.modeConfig
	if w.URL != nil && w.Service != nil {
		return nil, errors.New("service and URL cannot be used simultaneously")
	}

	scheme := "http"
	if w.CABundle != nil {
		scheme = "https"

		if err := setCABundle(w.CABundle); err != nil {
			return nil, err
		}
	}

	if w.URL != nil {
		if *w.URL == "" {
			return nil, errors.New("URL was not provided")
		}

		return url.ParseRequestURI(*w.URL)
	}

	if w.Service == nil {
		return nil, errors.New("service was not provided, either URL or Service must be provided")
	}

	if w.Service.Name == "" {
		return nil, errors.New("service name was not provided")
	}

	if w.Service.Path == nil {
		empty := ""
		w.Service.Path = &empty
	}

	if w.Service.Namespace == "" {
		w.Service.Namespace = "default"
	}

	return createURL(scheme, w.Service.Name, w.Service.Namespace, *w.Service.Path, w.Service.Port), nil
}

// moved to a separate method to cover it with unit tests and check that URL corresponds to a proper pattern
func createURL(scheme, name, namespace, path string, port *int32) *url.URL {
	var hostPort int32 = 8000
	if port != nil {
		hostPort = *port
	}

	return &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s.%s.svc:%d", name, namespace, hostPort),
		Path:   path,
	}
}

func setCABundle(caBundle []byte) error {
	// We can have multiple fleetautoscalers with different CABundles defined,
	// so we switch client.Transport before each POST request
	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(caBundle); !ok {
		return errors.New("no certs were appended from caBundle")
	}
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: rootCAs,
		},
	}
	return nil
}
