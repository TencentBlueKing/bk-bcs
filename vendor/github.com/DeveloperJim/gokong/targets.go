package gokong

import (
	"encoding/json"
	"fmt"
)

type TargetClient struct {
	config *Config
}

type TargetRequest struct {
	Target string `json:"target" yaml:"target"`
	Weight int    `json:"weight" yaml:"weight"`
}

type Target struct {
	Id        *string  `json:"id,omitempty" yaml:"id,omitempty"`
	CreatedAt *float32 `json:"created_at" yaml:"created_at"`
	Target    *string  `json:"target" yaml:"target"`
	Weight    *int     `json:"weight" yaml:"weight"`
	Upstream  *Id      `json:"upstream" yaml:"upstream"`
	Health    *string  `json:"health" yaml:"health"`
}

type Targets struct {
	Data   []*Target `json:"data" yaml:"data"`
	Total  int       `json:"total,omitempty" yaml:"total,omitempty"`
	Next   string    `json:"next,omitempty" yaml:"next,omitempty"`
	NodeId string    `json:"node_id,omitempty" yaml:"node_id,omitempty"`
}

const TargetsPath = "/upstreams/%s/targets"

func (targetClient *TargetClient) CreateFromUpstreamName(name string, targetRequest *TargetRequest) (*Target, error) {
	return targetClient.CreateFromUpstreamId(name, targetRequest)
}

func (targetClient *TargetClient) CreateFromUpstreamId(id string, targetRequest *TargetRequest) (*Target, error) {
	r, body, errs := newPost(targetClient.config, targetClient.config.HostAddress+fmt.Sprintf(TargetsPath, id)).Send(targetRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not register the target, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	createdTarget := &Target{}
	err := json.Unmarshal([]byte(body), createdTarget)
	if err != nil {
		return nil, fmt.Errorf("could not parse target get response, error: %v", err)
	}

	if createdTarget.Id == nil {
		return nil, fmt.Errorf("could not register the target, error: %v", body)
	}

	return createdTarget, nil
}

func (targetClient *TargetClient) GetTargetsFromUpstreamName(name string) ([]*Target, error) {
	return targetClient.GetTargetsFromUpstreamId(name)
}

func (targetClient *TargetClient) GetTargetsFromUpstreamId(id string) ([]*Target, error) {
	targets := []*Target{}
	data := &Targets{}

	for {
		r, body, errs := newGet(targetClient.config, targetClient.config.HostAddress+fmt.Sprintf(TargetsPath, id)).End()
		if errs != nil {
			return nil, fmt.Errorf("could not get targets, error: %v", errs)
		}

		if r.StatusCode == 401 || r.StatusCode == 403 {
			return nil, fmt.Errorf("not authorised, message from kong: %s", body)
		}

		if r.StatusCode == 404 {
			return nil, fmt.Errorf("non existent upstream: %s", id)
		}

		err := json.Unmarshal([]byte(body), data)
		if err != nil {
			return nil, fmt.Errorf("could not parse target get response, error: %v", err)
		}

		targets = append(targets, data.Data...)

		if data.Next == "" {
			break
		}
	}
	return targets, nil
}

func (targetClient *TargetClient) DeleteFromUpstreamByHostPort(upstreamNameOrId string, hostPort string) error {
	return targetClient.DeleteFromUpstreamById(upstreamNameOrId, hostPort)
}

func (targetClient *TargetClient) DeleteFromUpstreamById(upstreamNameOrId string, id string) error {
	r, body, errs := newDelete(targetClient.config, targetClient.config.HostAddress+fmt.Sprintf(TargetsPath, upstreamNameOrId)+fmt.Sprintf("/%s", id)).End()
	if errs != nil {
		return fmt.Errorf("could not delete the target, result: %v error: %v", r, errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return fmt.Errorf("not authorised, message from kong: %s", body)
	}

	if r.StatusCode != 204 {
		return fmt.Errorf("Received unexpected response status code: %d. Body: %s", r.StatusCode, body)
	}

	return nil
}

func (targetClient *TargetClient) SetTargetFromUpstreamByHostPortAsHealthy(upstreamNameOrId string, hostPort string) error {
	return targetClient.SetTargetFromUpstreamByIdAsHealthy(upstreamNameOrId, hostPort)
}

func (targetClient *TargetClient) SetTargetFromUpstreamByIdAsHealthy(upstreamNameOrId string, id string) error {
	r, body, errs := newPost(targetClient.config, targetClient.config.HostAddress+fmt.Sprintf(TargetsPath, upstreamNameOrId)+fmt.Sprintf("/%s/healthy", id)).Send("").End()
	if errs != nil {
		return fmt.Errorf("could not set the target as healthy, result: %v error: %v", r, errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return fmt.Errorf("not authorised, message from kong: %s", body)
	}

	if r.StatusCode != 204 {
		return fmt.Errorf("Received unexpected response status code: %d. Body: %s", r.StatusCode, body)
	}

	return nil
}

func (targetClient *TargetClient) SetTargetFromUpstreamByHostPortAsUnhealthy(upstreamNameOrId string, hostPort string) error {
	return targetClient.SetTargetFromUpstreamByIdAsUnhealthy(upstreamNameOrId, hostPort)
}

func (targetClient *TargetClient) SetTargetFromUpstreamByIdAsUnhealthy(upstreamNameOrId string, id string) error {
	r, body, errs := newPost(targetClient.config, targetClient.config.HostAddress+fmt.Sprintf(TargetsPath, upstreamNameOrId)+fmt.Sprintf("/%s/unhealthy", id)).Send("").End()
	if errs != nil {
		return fmt.Errorf("could not set the target as unhealthy, result: %v error: %v", r, errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return fmt.Errorf("not authorised, message from kong: %s", body)
	}

	if r.StatusCode != 204 {
		return fmt.Errorf("Received unexpected response status code: %d. Body: %s", r.StatusCode, body)
	}

	return nil
}

func (targetClient *TargetClient) GetTargetsWithHealthFromUpstreamName(name string) ([]*Target, error) {
	return targetClient.GetTargetsWithHealthFromUpstreamId(name)
}

func (targetClient *TargetClient) GetTargetsWithHealthFromUpstreamId(id string) ([]*Target, error) {
	targets := []*Target{}
	data := &Targets{}

	for {
		r, body, errs := newGet(targetClient.config, targetClient.config.HostAddress+fmt.Sprintf("/upstreams/%s/health", id)).End()
		if errs != nil {
			return nil, fmt.Errorf("could not get targets, error: %v", errs)
		}

		if r.StatusCode == 401 || r.StatusCode == 403 {
			return nil, fmt.Errorf("not authorised, message from kong: %s", body)
		}

		if r.StatusCode == 404 {
			return nil, fmt.Errorf("non existent upstream: %s", id)
		}

		err := json.Unmarshal([]byte(body), data)
		if err != nil {
			return nil, fmt.Errorf("could not parse target get response, error: %v", err)
		}

		targets = append(targets, data.Data...)

		if data.Next == "" {
			break
		}
	}
	return targets, nil
}

// TODO: Implement List all Targets - https://docs.konghq.com/1.0.x/admin-api/#list-all-targets
// Note: JSON returned by this method has slightly different structure to the standard Get request. This method retruns "upstream_id": "07131005-ba30-4204-a29f-0927d53257b4" instead of "upstream": {"id":"127dfc88-ed57-45bf-b77a-a9d3a152ad31"},

// TODO: Implement other methods available by /targets paths
// https://docs.konghq.com/1.0.x/admin-api/#update-or-create-upstream
// https://docs.konghq.com/1.0.x/admin-api/#delete-upstream
