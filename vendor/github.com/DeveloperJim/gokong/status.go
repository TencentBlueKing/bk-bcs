package gokong

import (
	"encoding/json"
	"errors"
	"fmt"
)

type StatusClient struct {
	config *Config
}

type Status struct {
	Server   serverStatus   `json:"server" yaml:"server"`
	Database databaseStatus `json:"database" yaml:"database"`
}

type serverStatus struct {
	TotalRequests       int `json:"total_requests" yaml:"total_requests"`
	ConnectionsActive   int `json:"connections_active" yaml:"connections_active"`
	ConnectionsAccepted int `json:"connections_accepted" yaml:"connections_accepted"`
	ConnectionsHandled  int `json:"connections_handled" yaml:"connections_handled"`
	ConnectionsReading  int `json:"connections_reading" yaml:"connections_reading"`
	ConnectionsWriting  int `json:"connections_writing" yaml:"connections_writing"`
	ConnectionsWaiting  int `json:"connections_waiting" yaml:"connections_waiting"`
}

type databaseStatus struct {
	Reachable bool `json:"reachable" yaml:"reachable"`
}

func (statusClient *StatusClient) Get() (*Status, error) {

	_, body, errs := newGet(statusClient.config, statusClient.config.HostAddress+"/status").End()
	if errs != nil {
		return nil, errors.New(fmt.Sprintf("Could not call get status, error: %v", errs))
	}

	status := &Status{}
	err := json.Unmarshal([]byte(body), status)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse status response, error: %v", err))
	}

	return status, nil

}
