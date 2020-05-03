package gokong

import (
	"encoding/json"
	"fmt"
)

type CertificateClient struct {
	config *Config
}

type CertificateRequest struct {
	Cert *string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key  *string `json:"key,omitempty" yaml:"key,omitempty"`
}

type Certificate struct {
	Id   *string `json:"id,omitempty" yaml:"id,omitempty"`
	Cert *string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key  *string `json:"key,omitempty" yaml:"key,omitempty"`
}

type Certificates struct {
	Results []*Certificate `json:"data,omitempty" yaml:"data,omitempty"`
	Total   int            `json:"total,omitempty" yaml:"total,omitempty"`
}

const CertificatesPath = "/certificates/"

func (certificateClient *CertificateClient) GetById(id string) (*Certificate, error) {

	r, body, errs := newGet(certificateClient.config, certificateClient.config.HostAddress+CertificatesPath+id).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get certificate, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	certificate := &Certificate{}
	err := json.Unmarshal([]byte(body), certificate)
	if err != nil {
		return nil, fmt.Errorf("could not parse certificate get response, error: %v", err)
	}

	if certificate.Id == nil {
		return nil, nil
	}

	return certificate, nil
}

func (certificateClient *CertificateClient) Create(certificateRequest *CertificateRequest) (*Certificate, error) {

	r, body, errs := newPost(certificateClient.config, certificateClient.config.HostAddress+CertificatesPath).Send(certificateRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not create new certificate, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	createdCertificate := &Certificate{}
	err := json.Unmarshal([]byte(body), createdCertificate)
	if err != nil {
		return nil, fmt.Errorf("could not parse certificate creation response, error: %v", err)
	}

	if createdCertificate.Id == nil {
		return nil, fmt.Errorf("could not create certificate, error: %v", body)
	}

	return createdCertificate, nil
}

func (certificateClient *CertificateClient) DeleteById(id string) error {

	r, body, errs := newDelete(certificateClient.config, certificateClient.config.HostAddress+CertificatesPath+id).End()
	if errs != nil {
		return fmt.Errorf("could not delete certificate, result: %v error: %v", r, errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return fmt.Errorf("not authorised, message from kong: %s", body)
	}

	return nil
}

func (certificateClient *CertificateClient) List() (*Certificates, error) {

	r, body, errs := newGet(certificateClient.config, certificateClient.config.HostAddress+CertificatesPath).End()
	if errs != nil {
		return nil, fmt.Errorf("could not get certificates, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	certificates := &Certificates{}
	err := json.Unmarshal([]byte(body), certificates)
	if err != nil {
		return nil, fmt.Errorf("could not parse certificates list response, error: %v", err)
	}

	return certificates, nil
}

func (certificateClient *CertificateClient) UpdateById(id string, certificateRequest *CertificateRequest) (*Certificate, error) {

	r, body, errs := newPatch(certificateClient.config, certificateClient.config.HostAddress+CertificatesPath+id).Send(certificateRequest).End()
	if errs != nil {
		return nil, fmt.Errorf("could not update certificate, error: %v", errs)
	}

	if r.StatusCode == 401 || r.StatusCode == 403 {
		return nil, fmt.Errorf("not authorised, message from kong: %s", body)
	}

	updatedCertificate := &Certificate{}
	err := json.Unmarshal([]byte(body), updatedCertificate)
	if err != nil {
		return nil, fmt.Errorf("could not parse certificate update response, error: %v", err)
	}

	if updatedCertificate.Id == nil {
		return nil, fmt.Errorf("could not update certificate, error: %v", body)
	}

	return updatedCertificate, nil
}
