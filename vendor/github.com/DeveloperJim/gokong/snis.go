package gokong

import (
	"encoding/json"
	"fmt"
)

type SnisClient struct {
	config *Config
}

type SnisRequest struct {
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	CertificateId *Id    `json:"certificate,omitempty" yaml:"certificate,omitempty"`
}

type Sni struct {
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	CertificateId *Id    `json:"certificate,omitempty" yaml:"certificate,omitempty"`
}

type Snis struct {
	Results []*Sni `json:"data,omitempty" yaml:"data,omitempty"`
	Total   int    `json:"total,omitempty" yaml:"total,omitempty"`
}

const SnisPath = "/snis/"

func (snisClient *SnisClient) Create(snisRequest *SnisRequest) (*Sni, error) {

	r, body, errs := newPost(snisClient.config, snisClient.config.HostAddress+SnisPath).Send(snisRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not create new sni, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	sni := &Sni{}
	err := json.Unmarshal([]byte(body), sni)
	if err != nil {
		return nil, fmt.Errorf("could not parse sni creation response, error: %v", err)
	}

	if sni.CertificateId == nil {
		return nil, fmt.Errorf("could not create sni, error: %v", body)
	}

	return sni, nil
}

func (snisClient *SnisClient) GetByName(name string) (*Sni, error) {

	r, body, errs := newGet(snisClient.config, snisClient.config.HostAddress+SnisPath+name).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get sni, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	sni := &Sni{}
	err := json.Unmarshal([]byte(body), sni)
	if err != nil {
		return nil, fmt.Errorf("could not parse sni get response, error: %v", err)
	}

	if sni.Name == "" {
		return nil, nil
	}

	return sni, nil
}

func (snisClient *SnisClient) List() (*Snis, error) {

	r, body, errs := newGet(snisClient.config, snisClient.config.HostAddress+SnisPath).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get snis, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	snis := &Snis{}
	err := json.Unmarshal([]byte(body), snis)
	if err != nil {
		return nil, fmt.Errorf("could not parse snis list response, error: %v", err)
	}

	return snis, nil
}

func (snisClient *SnisClient) DeleteByName(name string) error {

	r, body, errs := newDelete(snisClient.config, snisClient.config.HostAddress+SnisPath+name).End()
	if errs != nil {
		return fmt.Errorf("could not delete sni, result: %v error: %v", r, errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return fmt.Errorf("not authorised, message from kong: %s", body)
	}

	return nil
}

func (snisClient *SnisClient) UpdateByName(name string, snisRequest *SnisRequest) (*Sni, error) {

	r, body, errs := newPatch(snisClient.config, snisClient.config.HostAddress+SnisPath+name).Send(snisRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not update sni, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	updatedSni := &Sni{}
	err := json.Unmarshal([]byte(body), updatedSni)
	if err != nil {
		return nil, fmt.Errorf("could not parse sni update response, error: %v", err)
	}

	if updatedSni.CertificateId == nil {
		return nil, fmt.Errorf("could not update sni, error: %v", body)
	}

	return updatedSni, nil
}
