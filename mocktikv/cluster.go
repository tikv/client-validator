// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package mocktikv

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var httpClient = http.Client{
	Timeout: 10 * time.Second,
}

// MockCluster contains the mock cluster info respond from mock-tikv.
// It should be kept synced with mock-tikv.
type MockCluster struct {
	ID      uint64        `json:"id"`
	Members []*MockMember `json:"members"`
}

// MockMember contains the mock PD member respond from mock-tikv.
// It should be kept synced with mock-tikv.
type MockMember struct {
	Name       string   `json:"name"`
	MemberID   uint64   `json:"member_id"`
	ClientUrls []string `json:"client_urls"`
}

// Cluster represents a mock cluster in mock-tikv server.
type Cluster struct {
	mockServer string
	clusterID  uint64
	pdAddrs    []string
}

// NewCluster creates a mock cluster in mock-tikv server.
func NewCluster(mockServer string) (*Cluster, error) {
	res, err := httpClient.Post(mockServer+"/mock-tikv/api/v1/clusters", "application/json", strings.NewReader("{}"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var cluster MockCluster
	err = json.Unmarshal(data, &cluster)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var addrs []string
	for _, m := range cluster.Members {
		for _, url := range m.ClientUrls {
			addrs = append(addrs, strings.TrimPrefix(url, "http://"))
		}
	}
	return &Cluster{
		mockServer: mockServer,
		clusterID:  cluster.ID,
		pdAddrs:    addrs,
	}, nil
}

// ClusterID returns the mock cluster's ID.
func (c *Cluster) ClusterID() uint64 {
	return c.clusterID
}

// PDAddrs returns the mock cluster's PD addresses.
func (c *Cluster) PDAddrs() []string {
	return c.pdAddrs
}

// Close releases the mock cluster.
func (c *Cluster) Close() error {
	req, err := http.NewRequest("DELETE", c.mockServer+"/mock-tikv/api/v1/clusters", nil)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = httpClient.Do(req)
	return errors.WithStack(err)
}
