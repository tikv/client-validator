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

package stub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// RawClientStub can be used like a rawkv.Client while it redirects all function
// calls to an httpproxy server.
type RawClientStub struct {
	client      http.Client
	proxyServer string
	id          string
}

// NewRawClientStub creates a client for rawkv calls.
func NewRawClientStub(proxyServer string, pdServers []string) (*RawClientStub, error) {
	client := &RawClientStub{
		client:      http.Client{Timeout: time.Second * 10},
		proxyServer: strings.TrimSuffix(proxyServer, "/"),
	}
	res, err := client.send("/rawkv/client/new", &RawRequest{PDAddrs: pdServers})
	if err != nil {
		return nil, err
	}
	client.id = res.ID
	return client, nil
}

// Close closes the client and releases resources in proxy server.
func (c *RawClientStub) Close() error {
	_, err := c.send(fmt.Sprintf("/rawkv/client/%s/close", c.id), &RawRequest{})
	return err
}

// Get queries value with the key. When the key does not exist, it returns `nil, nil`.
func (c *RawClientStub) Get(key []byte) ([]byte, error) {
	res, err := c.send(fmt.Sprintf("/rawkv/client/%s/get", c.id), &RawRequest{Key: key})
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

// BatchGet queries values with the keys.
func (c *RawClientStub) BatchGet(keys [][]byte) ([][]byte, error) {
	res, err := c.send(fmt.Sprintf("/rawkv/client/%s/batch-get", c.id), &RawRequest{Keys: keys})
	if err != nil {
		return nil, err
	}
	return res.Values, nil
}

// Put stores a key-value pair to TiKV.
func (c *RawClientStub) Put(key, value []byte) error {
	_, err := c.send(fmt.Sprintf("/rawkv/client/%s/put", c.id), &RawRequest{Key: key, Value: value})
	return err
}

// BatchPut stores key-value pairs to TiKV.
func (c *RawClientStub) BatchPut(keys, values [][]byte) error {
	_, err := c.send(fmt.Sprintf("/rawkv/client/%s/batch-put", c.id), &RawRequest{Keys: keys, Values: values})
	return err
}

// Delete deletes a key-value pair from TiKV.
func (c *RawClientStub) Delete(key []byte) error {
	_, err := c.send(fmt.Sprintf("/rawkv/client/%s/delete", c.id), &RawRequest{Key: key})
	return err
}

// BatchDelete deletes key-value pairs from TiKV.
func (c *RawClientStub) BatchDelete(keys [][]byte) error {
	_, err := c.send(fmt.Sprintf("/rawkv/client/%s/batch-delete", c.id), &RawRequest{Keys: keys})
	return err
}

// DeleteRange deletes all key-value pairs in a range from TiKV.
func (c *RawClientStub) DeleteRange(startKey, endKey []byte) error {
	_, err := c.send(fmt.Sprintf("/rawkv/client/%s/delete-range", c.id), &RawRequest{StartKey: startKey, EndKey: endKey})
	return err
}

// Scan queries continuous kv pairs in range [startKey, endKey), up to limit pairs.
func (c *RawClientStub) Scan(startKey, endKey []byte, limit int) ([][]byte, [][]byte, error) {
	res, err := c.send(fmt.Sprintf("/rawkv/client/%s/scan", c.id), &RawRequest{StartKey: startKey, EndKey: endKey, Limit: limit})
	if err != nil {
		return nil, nil, err
	}
	return res.Keys, res.Values, nil
}

func (c *RawClientStub) send(uri string, req *RawRequest) (*RawResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res, err := c.client.Post(c.proxyServer+uri, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch {
	case res.StatusCode >= 200 && res.StatusCode < 300: // 2xx means OK.
		var resp RawResponse
		if err = json.Unmarshal(body, &resp); err != nil {
			return nil, errors.WithStack(err)
		}
		return &resp, nil
	default:
		return nil, errors.New(string(body))
	}
}
