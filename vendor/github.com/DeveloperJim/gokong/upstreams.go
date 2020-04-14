package gokong

import (
	"encoding/json"
	"fmt"
)

type UpstreamClient struct {
	config *Config
}

type UpstreamRequest struct {
	Name               string               `json:"name" yaml:"name"`
	Slots              int                  `json:"slots,omitempty" yaml:"slots,omitempty"`
	HashOn             string               `json:"hash_on,omitempty" yaml:"hash_on,omitempty"`
	HashFallback       string               `json:"hash_fallback,omitempty" yaml:"hash_fallback,omitempty"`
	HashOnHeader       string               `json:"hash_on_header,omitempty" yaml:"hash_on_header,omitempty"`
	HashFallbackHeader string               `json:"hash_fallback_header,omitempty" yaml:"hash_fallback_header,omitempty"`
	HashOnCookie       string               `json:"hash_on_cookie,omitempty" yaml:"hash_on_cookie,omitempty"`
	HashOnCookiePath   string               `json:"hash_on_cookie_path,omitempty" yaml:"hash_on_cookie_path,omitempty"`
	HealthChecks       *UpstreamHealthCheck `json:"healthchecks,omitempty" yaml:"healthchecks,omitempty"`
	Tags               []*string            `json:"tags" yaml:"tags"`
}

type UpstreamHealthCheck struct {
	Active  *UpstreamHealthCheckActive  `json:"active,omitempty" yaml:"active,omitempty"`
	Passive *UpstreamHealthCheckPassive `json:"passive,omitempty" yaml:"passive,omitempty"`
}

type UpstreamHealthCheckActive struct {
	Type                   string           `json:"type,omitempty" yaml:"type,omitempty"`
	Concurrency            int              `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	Healthy                *ActiveHealthy   `json:"healthy,omitempty" yaml:"healthy,omitempty"`
	HttpPath               string           `json:"http_path,omitempty" yaml:"http_path,omitempty"`
	HttpsVerifyCertificate bool             `json:"https_verify_certificate" yaml:"https_verify_certificate"`
	HttpsSni               *string          `json:"https_sni,omitempty" yaml:"https_sni,omitempty"`
	Timeout                int              `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Unhealthy              *ActiveUnhealthy `json:"unhealthy,omitempty" yaml:"unhealthy,omitempty"`
}

type ActiveHealthy struct {
	HttpStatuses []int `json:"http_statuses,omitempty" yaml:"http_statuses,omitempty"`
	Interval     int   `json:"interval" yaml:"interval"`
	Successes    int   `json:"successes" yaml:"successes"`
}

type ActiveUnhealthy struct {
	HttpFailures int   `json:"http_failures" yaml:"http_failures"`
	HttpStatuses []int `json:"http_statuses,omitempty" yaml:"http_statuses,omitempty"`
	Interval     int   `json:"interval" yaml:"interval"`
	TcpFailures  int   `json:"tcp_failures" yaml:"tcp_failures"`
	Timeouts     int   `json:"timeouts" yaml:"timeouts"`
}

type UpstreamHealthCheckPassive struct {
	Type      string            `json:"type,omitempty" yaml:"type,omitempty"`
	Healthy   *PassiveHealthy   `json:"healthy,omitempty yaml:"healthy,omitempty"`
	Unhealthy *PassiveUnhealthy `json:"unhealthy,omitempty yaml:"unhealthy,omitempty"`
}

type PassiveHealthy struct {
	HttpStatuses []int `json:"http_statuses,omitempty" yaml:"http_statuses,omitempty"`
	Successes    int   `json:"successes" yaml:"successes"`
}

type PassiveUnhealthy struct {
	HttpFailures int   `json:"http_failures" yaml:"http_failures"`
	HttpStatuses []int `json:"http_statuses,omitempty" yaml:"http_statuses,omitempty"`
	TcpFailures  int   `json:"tcp_failures" yaml:"tcp_failures"`
	Timeouts     int   `json:"timeouts" yaml:"timeouts"`
}

type Upstream struct {
	Id string `json:"id,omitempty" yaml:"id,omitempty"`
	UpstreamRequest
}

type Upstreams struct {
	Results []*Upstream `json:"data,omitempty" yaml:"data,omitempty"`
	Next    string      `json:"next,omitempty" yaml:"next,omitempty"`
}

const UpstreamsPath = "/upstreams/"

func (upstreamClient *UpstreamClient) GetByName(name string) (*Upstream, error) {
	return upstreamClient.GetById(name)
}

func (upstreamClient *UpstreamClient) GetById(id string) (*Upstream, error) {

	r, body, errs := newGet(upstreamClient.config, upstreamClient.config.HostAddress+UpstreamsPath+id).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get upstream, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	upstream := &Upstream{}
	err := json.Unmarshal([]byte(body), upstream)
	if err != nil {
		return nil, fmt.Errorf("could not parse upstream get response, error: %v", err)
	}

	if upstream.Id == "" {
		return nil, nil
	}

	return upstream, nil
}

func (upstreamClient *UpstreamClient) Create(upstreamRequest *UpstreamRequest) (*Upstream, error) {

	r, body, errs := newPost(upstreamClient.config, upstreamClient.config.HostAddress+UpstreamsPath).Send(upstreamRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not create new upstream, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	createdUpstream := &Upstream{}
	err := json.Unmarshal([]byte(body), createdUpstream)
	if err != nil {
		return nil, fmt.Errorf("could not parse upstream creation response, error: %v", err)
	}

	if createdUpstream.Id == "" {
		return nil, fmt.Errorf("could not create update, error: %v", body)
	}

	return createdUpstream, nil
}

func (upstreamClient *UpstreamClient) DeleteByName(name string) error {
	return upstreamClient.DeleteById(name)
}

func (upstreamClient *UpstreamClient) DeleteById(id string) error {

	r, body, errs := newDelete(upstreamClient.config, upstreamClient.config.HostAddress+UpstreamsPath+id).End()
	if errs != nil {
		return fmt.Errorf("could not delete upstream, result: %v error: %v", r, errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return fmt.Errorf("not authorised, message from kong: %s", body)
	}

	return nil
}

func (upstreamClient *UpstreamClient) List() (*Upstreams, error) {

	r, body, errs := newGet(upstreamClient.config, upstreamClient.config.HostAddress+UpstreamsPath).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get upstreams, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	upstreams := &Upstreams{}
	err := json.Unmarshal([]byte(body), upstreams)
	if err != nil {
		return nil, fmt.Errorf("could not parse upstreams list response, error: %v", err)
	}

	return upstreams, nil
}

func (upstreamClient *UpstreamClient) UpdateByName(name string, upstreamRequest *UpstreamRequest) (*Upstream, error) {
	return upstreamClient.UpdateById(name, upstreamRequest)
}

func (upstreamClient *UpstreamClient) UpdateById(id string, upstreamRequest *UpstreamRequest) (*Upstream, error) {

	r, body, errs := newPatch(upstreamClient.config, upstreamClient.config.HostAddress+UpstreamsPath+id).Send(upstreamRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not update upstream, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	updatedUpstream := &Upstream{}
	err := json.Unmarshal([]byte(body), updatedUpstream)
	if err != nil {
		return nil, fmt.Errorf("could not parse upstream update response, error: %v", err)
	}

	if updatedUpstream.Id == "" {
		return nil, fmt.Errorf("could not update upstream, error: %v", body)
	}

	return updatedUpstream, nil
}
