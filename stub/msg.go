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

// RawRequest is the structure of a rawkv request that the http proxy accepts.
// It should be kept synced with the proxy server.
type RawRequest struct {
	PDAddrs  []string `json:"pd_addrs,omitempty"`  // for new
	Key      []byte   `json:"key,omitempty"`       // for get, put, delete
	Keys     [][]byte `json:"keys,omitempty"`      // for batchGet, batchPut, batchDelete
	Value    []byte   `json:"value,omitempty"`     // for put
	Values   [][]byte `json:"values,omitmepty"`    // for batchPut
	StartKey []byte   `json:"start_key,omitempty"` // for scan, deleteRange
	EndKey   []byte   `json:"end_key,omitempty"`   // for scan, deleteRange
	Limit    int      `json:"limit,omitempty"`     // for scan
}

// RawResponse is the structure of a rawkv response that the http proxy sends.
// It should be kept synced with the proxy server.
type RawResponse struct {
	ID     string   `json:"id,omitempty"`     // for new
	Value  []byte   `json:"value,omitempty"`  // for get
	Keys   [][]byte `json:"keys,omitempty"`   // for scan
	Values [][]byte `json:"values,omitempty"` // for batchGet
}

// TxnRequest is the structure of a txnkv request that the http proxy accepts.
// It should be kept synced with the proxy server.
type TxnRequest struct {
	PDAddrs    []string `json:"pd_addrs,omitempty"`    // for new
	TS         uint64   `json:"ts,omitempty"`          // for beginWithTS
	Key        []byte   `json:"key,omitempty"`         // for get, set, delete, iter, iterReverse
	Value      []byte   `json:"value,omitempty"`       // for set
	Keys       [][]byte `json:"keys,omitempty"`        // for batchGet, lockKeys
	UpperBound []byte   `json:"upper_bound,omitempty"` // for iter
}

// TxnResponse is the structure of a txnkv response that the http proxy sends.
// It should be kept synced with the proxy server.
type TxnResponse struct {
	ID         string   `json:"id,omitempty"`          // for new, begin, beginWithTS, iter, iterReverse
	TS         uint64   `json:"ts,omitempty"`          // for getTS
	Key        []byte   `json:"key,omitempty"`         // for iterKey
	Value      []byte   `json:"value,omitempty"`       // for get, iterValue
	Keys       [][]byte `json:"keys,omitempty"`        // for batchGet
	Values     [][]byte `json:"values,omitempty"`      // for batchGet
	IsValid    bool     `json:"is_valid,omitempty"`    // for valid, iterValid
	IsReadOnly bool     `json:"is_readonly,omitempty"` // for isReadOnly
	Size       int      `json:"size,omitempty"`        // for size
	Length     int      `json:"length,omitempty"`      // for length
}
