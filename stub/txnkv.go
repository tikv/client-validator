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

// TxnClientStub can be used like a txnkv.Client while it redirects all function
// calls to an httpproxy server.
type TxnClientStub struct {
	client      http.Client
	proxyServer string
	id          string
}

// NewTxnClientStub creates a client for txnkv calls.
func NewTxnClientStub(proxyServer string, pdServers []string) (*TxnClientStub, error) {
	client := &TxnClientStub{
		client:      http.Client{Timeout: time.Second * 10},
		proxyServer: strings.TrimSuffix(proxyServer, "/"),
	}
	res, err := client.send("/txnkv/client/new", &TxnRequest{PDAddrs: pdServers})
	if err != nil {
		return nil, err
	}
	client.id = res.ID
	return client, nil
}

// Close closes the client and releases resources in proxy server.
func (c *TxnClientStub) Close() error {
	_, err := c.send(fmt.Sprintf("/txnkv/client/%s/close", c.id), &TxnRequest{})
	return err
}

// Begin creates a transaction for read/write.
func (c *TxnClientStub) Begin() (*TransactionStub, error) {
	res, err := c.send(fmt.Sprintf("/txnkv/client/%s/begin", c.id), &TxnRequest{})
	if err != nil {
		return nil, err
	}
	return &TransactionStub{
		client: c,
		id:     res.ID,
	}, nil
}

// BeginWithTS creates a transaction which is normally readonly.
func (c *TxnClientStub) BeginWithTS(ts uint64) (*TransactionStub, error) {
	res, err := c.send(fmt.Sprintf("/txnkv/client/%s/begin-with-ts", c.id), &TxnRequest{TS: ts})
	if err != nil {
		return nil, err
	}
	return &TransactionStub{
		client: c,
		id:     res.ID,
	}, nil
}

// GetTS returns a latest timestamp.
func (c *TxnClientStub) GetTS() (uint64, error) {
	res, err := c.send(fmt.Sprintf("/txnkv/client/%s/get-ts", c.id), &TxnRequest{})
	if err != nil {
		return 0, err
	}
	return res.TS, nil
}

// TransactionStub can be used like a txnkv.Trasaction while it redirects all
// function calls to an httpproxy server.
type TransactionStub struct {
	client *TxnClientStub
	id     string
}

func (txn *TransactionStub) String() string {
	return txn.id
}

// Get retrives the value for the given key.
func (txn *TransactionStub) Get(k []byte) ([]byte, error) {
	res, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/get", txn.id), &TxnRequest{Key: k})
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

// BatchGet gets a batch of values.
func (txn *TransactionStub) BatchGet(keys [][]byte) (map[string][]byte, error) {
	res, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/batch-get", txn.id), &TxnRequest{Keys: keys})
	if err != nil {
		return nil, err
	}
	m := make(map[string][]byte, len(res.Keys))
	for i := range res.Keys {
		m[string(res.Keys[i])] = res.Values[i]
	}
	return m, nil
}

// Set sets the value for key k as v.
func (txn *TransactionStub) Set(k []byte, v []byte) error {
	_, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/set", txn.id), &TxnRequest{Key: k, Value: v})
	return err
}

// Iter creates an Iterator positioned on the first entry that k <= entry's key.
func (txn *TransactionStub) Iter(k []byte, upperBound []byte) (*IteratorStub, error) {
	res, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/iter", txn.id), &TxnRequest{Key: k, UpperBound: upperBound})
	if err != nil {
		return nil, err
	}
	return &IteratorStub{
		client: txn.client,
		id:     res.ID,
	}, nil
}

// IterReverse creates a reversed Iterator positioned on the first entry which key is less than k.
func (txn *TransactionStub) IterReverse(k []byte) (*IteratorStub, error) {
	res, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/iter-reverse", txn.id), &TxnRequest{Key: k})
	if err != nil {
		return nil, err
	}
	return &IteratorStub{
		client: txn.client,
		id:     res.ID,
	}, nil
}

// IsReadOnly returns if there are pending key-value to commit in the transaction.
func (txn *TransactionStub) IsReadOnly() (bool, error) {
	res, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/readonly", txn.id), &TxnRequest{})
	if err != nil {
		return false, err
	}
	return res.IsReadOnly, nil
}

// Delete removes the entry for key k.
func (txn *TransactionStub) Delete(k []byte) error {
	_, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/delete", txn.id), &TxnRequest{Key: k})
	return err
}

// Commit commits the transaction operations.
func (txn *TransactionStub) Commit() error {
	_, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/commit", txn.id), &TxnRequest{})
	return err
}

// Rollback undoes the transaction operations.
func (txn *TransactionStub) Rollback() error {
	_, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/rollback", txn.id), &TxnRequest{})
	return err
}

// LockKeys tries to lock the entries with the keys.
func (txn *TransactionStub) LockKeys(keys ...[]byte) error {
	_, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/lock-keys", txn.id), &TxnRequest{Keys: keys})
	return err
}

// Valid returns if the transaction is valid.
// A transaction becomes invalid after commit or rollback.
func (txn *TransactionStub) Valid() (bool, error) {
	res, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/valid", txn.id), &TxnRequest{})
	if err != nil {
		return false, err
	}
	return res.IsValid, nil
}

// Len returns the count of key-value pairs in the transaction's memory buffer.
func (txn *TransactionStub) Len() (int, error) {
	res, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/len", txn.id), &TxnRequest{})
	if err != nil {
		return 0, err
	}
	return res.Length, nil
}

// Size returns the length (in bytes) of the transaction's memory buffer.
func (txn *TransactionStub) Size() (int, error) {
	res, err := txn.client.send(fmt.Sprintf("/txnkv/txn/%s/size", txn.id), &TxnRequest{})
	if err != nil {
		return 0, err
	}
	return res.Size, nil
}

// IteratorStub can be used like a txnkv.kv.Iterator while it redirects all
// function calls to an httpproxy server.
type IteratorStub struct {
	client *TxnClientStub
	id     string
}

// Valid returns if the iterator is valid to use.
func (iter *IteratorStub) Valid() (bool, error) {
	res, err := iter.client.send(fmt.Sprintf("/txnkv/iter/%s/valid", iter.id), &TxnRequest{})
	if err != nil {
		return false, nil
	}
	return res.IsValid, nil
}

// Key returns the key the iterator currently positioned at.
func (iter *IteratorStub) Key() ([]byte, error) {
	res, err := iter.client.send(fmt.Sprintf("/txnkv/iter/%s/key", iter.id), &TxnRequest{})
	if err != nil {
		return nil, err
	}
	return res.Key, nil
}

// Value returns the valid the iterator currently positioned at.
func (iter *IteratorStub) Value() ([]byte, error) {
	res, err := iter.client.send(fmt.Sprintf("/txnkv/iter/%s/value", iter.id), &TxnRequest{})
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

// Next move the iterator to next position.
func (iter *IteratorStub) Next() error {
	_, err := iter.client.send(fmt.Sprintf("/txnkv/iter/%s/next", iter.id), &TxnRequest{})
	return err
}

// Close releases the iterator.
func (iter *IteratorStub) Close() error {
	_, err := iter.client.send(fmt.Sprintf("/txnkv/iter/%s/close", iter.id), &TxnRequest{})
	return err
}

func (c *TxnClientStub) send(uri string, req *TxnRequest) (*TxnResponse, error) {
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
		var resp TxnResponse
		if err = json.Unmarshal(body, &resp); err != nil {
			return nil, errors.WithStack(err)
		}
		return &resp, nil
	default:
		return nil, errors.New(string(body))
	}
}
