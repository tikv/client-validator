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

package tests

import (
	"fmt"

	"github.com/tikv/client-validator/mocktikv"
	"github.com/tikv/client-validator/stub"
	"github.com/tikv/client-validator/validator"
)

type testRawKV struct{}

func (t testRawKV) newCluster(ctx validator.ExecContext) *mocktikv.Cluster {
	cluster, err := mocktikv.NewCluster(*mockTiKVAddr)
	ctx.AssertNil(err)
	return cluster
}

var _ = validator.Feature("rawkv.new", "create a rawkv client", nil, testRawKV{}.checkClientCreate)

func (t testRawKV) checkClientCreate(ctx validator.ExecContext) validator.FeatureStatus {
	cluster := t.newCluster(ctx)
	defer cluster.Close()
	client, err := stub.NewRawClientStub(*clientProxyAddr, cluster.PDAddrs())
	if err != nil {
		return errToFeatureStatus(err)
	}
	defer client.Close()
	return validator.FeaturePass
}

func (t testRawKV) newClient(ctx validator.ExecContext) (*mocktikv.Cluster, *stub.RawClientStub) {
	cluster := t.newCluster(ctx)
	client, err := stub.NewRawClientStub(*clientProxyAddr, cluster.PDAddrs())
	ctx.AssertNil(err)
	return cluster, client
}

var _ = validator.Feature("rawkv.close", "close a rawkv client", nil, testRawKV{}.checkClose)

func (t testRawKV) checkClose(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	err := client.Close()
	return errToFeatureStatus(err)
}

var _ = validator.Feature("rawkv.put", "store key-value pair in rawkv mod", []string{"rawkv.new"}, testRawKV{}.checkPut)

func (t testRawKV) checkPut(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	err := client.Put([]byte("k"), []byte("v"))
	return errToFeatureStatus(err)
}

var _ = validator.Feature("rawkv.get", "load key-value pair in raw mod", []string{"rawkv.new"}, testRawKV{}.checkGet)

func (t testRawKV) checkGet(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	val, err := client.Get([]byte("k"))
	if err != nil {
		return errToFeatureStatus(err)
	}
	ctx.Assert(len(val) == 0, "expect empty value")
	return validator.FeaturePass
}

var _ = validator.Feature("rawkv.delete", "delete key-value pair in raw mod", []string{"rawkv.put"}, testRawKV{}.checkDelete)

func (t testRawKV) checkDelete(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	err := client.Put([]byte("k"), []byte("v"))
	ctx.AssertNil(err)
	err = client.Delete([]byte("k"))
	return errToFeatureStatus(err)
}

// Ideally, `scan` does not have to depend on `put`. However, mock-tikv does not
// support inject data directly now, so we need `put` to prepare some data.
var _ = validator.Feature("rawkv.scan", "scan key-value pairs in raw mod", []string{"rawkv.put"}, testRawKV{}.checkScan)

func (t testRawKV) checkScan(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	err := client.Put([]byte("k1"), []byte("v1"))
	ctx.AssertNil(err)
	err = client.Put([]byte("k2"), []byte("v2"))
	ctx.AssertNil(err)
	keys, values, err := client.Scan(nil, nil, 2)
	if err != nil {
		return errToFeatureStatus(err)
	}
	ctx.AssertDeepEQ(keys, bss("k1", "k2"))
	ctx.AssertDeepEQ(values, bss("v1", "v2"))
	return validator.FeaturePass
}

var _ = validator.Story("basic rawkv client", "rawkv.new", "rawkv.close", "rawkv.get", "rawkv.put", "rawkv.delete", "rawkv.scan")

var _ = validator.Feature("rawkv.batch-get", "load rawkv in batches", []string{"rawkv.new"}, testRawKV{}.checkBatchGet)

func (t testRawKV) checkBatchGet(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	values, err := client.BatchGet(bss("k1", "k2"))
	if err != nil {
		return errToFeatureStatus(err)
	}
	ctx.AssertEQ(len(values), 2, "batch get should return equal lenght with keys (even if not exist)")
	ctx.AssertEQ(len(values[0]), 0)
	ctx.AssertEQ(len(values[1]), 0)
	return validator.FeaturePass
}

var _ = validator.Feature("rawkv.batch-put", "put rawkv in batches", []string{"rawkv.new"}, testRawKV{}.checkBatchPut)

func (t testRawKV) checkBatchPut(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	err := client.BatchPut(bss("k1", "k2"), bss("v1", "v2"))
	return errToFeatureStatus(err)
}

var _ = validator.Feature("rawkv.batch-delete", "delete rawkv in batches", []string{"rawkv.new"}, testRawKV{}.checkBatchDelete)

func (t testRawKV) checkBatchDelete(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	err := client.BatchDelete(bss("k1", "k2"))
	return errToFeatureStatus(err)
}

var _ = validator.Story("rawkv batch operations", "rawkv.batch-get", "rawkv.batch-put", "rawkv.batch-delete")

var _ = validator.Feature("rawkv.delete-range", "delete rawkv range", []string{"rawkv.new"}, testRawKV{}.checkDeleteRange)

func (t testRawKV) checkDeleteRange(ctx validator.ExecContext) validator.FeatureStatus {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	err := client.DeleteRange([]byte("k"), nil)
	return errToFeatureStatus(err)
}

var _ = validator.Test("simple rawkv get/put/delete", []string{"rawkv.get", "rawkv.put", "rawkv.delete"}, testRawKV{}.testSimple)

func (t testRawKV) testSimple(ctx validator.ExecContext) {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	t.mustNotExist(ctx, client, "key")
	t.mustPut(ctx, client, "key", "value")
	t.mustGet(ctx, client, "key", "value")
	t.mustDelete(ctx, client, "key")
	t.mustNotExist(ctx, client, "key")
}

var _ = validator.Test("put empty value is disallowed", []string{"rawkv.put"}, testRawKV{}.testPutEmptyValue)

func (t testRawKV) testPutEmptyValue(ctx validator.ExecContext) {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	err := client.Put([]byte("k"), []byte(""))
	ctx.AssertNotNil(err)
}

var _ = validator.Test("batch operations, batch-put/batch-get/batch-delete", []string{"rawkv.batch-put", "rawkv.batch-get", "rawkv.batch-delete"}, testRawKV{}.testBatch)

func (t testRawKV) testBatch(ctx validator.ExecContext) {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	var n, size int
	var keys, values []string
	for ; size/16*1024 < 4; n++ {
		key := fmt.Sprint("key", n)
		size += len(key)
		keys = append(keys, key)
		value := fmt.Sprint("value", n)
		size += len(value)
		values = append(values, value)
		t.mustNotExist(ctx, client, key)
	}
	t.mustSplit(ctx, cluster, "", fmt.Sprint("key", n/2))
	t.mustBatchPut(ctx, client, keys, values)
	t.mustBatchGet(ctx, client, keys, values)
	t.mustBatchDelete(ctx, client, keys)
	t.mustBatchNotExist(ctx, client, keys)
}

var _ = validator.Test("batch get with partial result", []string{"rawkv.batch-put", "rawkv.batch-get"}, testRawKV{}.testBatchGetPartial)

func (t testRawKV) testBatchGetPartial(ctx validator.ExecContext) {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	t.mustBatchPut(ctx, client, []string{"k1", "k3"}, []string{"v1", "v3"})
	t.mustBatchGet(ctx, client, []string{"k1", "k2", "k3", "k4"}, []string{"v1", "", "v3", ""})
}

var _ = validator.Test("region split", []string{"rawkv.get", "rawkv.put"}, testRawKV{}.testRegionSplit)

func (t testRawKV) testRegionSplit(ctx validator.ExecContext) {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	t.mustPut(ctx, client, "k1", "v1")
	t.mustPut(ctx, client, "k3", "v3")
	t.mustSplit(ctx, cluster, "k", "k2")
	t.mustGet(ctx, client, "k1", "v1")
	t.mustGet(ctx, client, "k3", "v3")
}

var _ = validator.Test("scan", []string{"rawkv.batch-put", "rawkv.scan"}, testRawKV{}.testScan)

func (t testRawKV) testScan(ctx validator.ExecContext) {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	t.mustBatchPut(ctx, client, []string{"k1", "k3", "k5", "k7"}, []string{"v1", "v3", "v5", "v7"})
	check := func() {
		t.mustScan(ctx, client, "", "", 1, "k1", "v1")
		t.mustScan(ctx, client, "k1", "", 2, "k1", "v1", "k3", "v3")
		t.mustScan(ctx, client, "", "", 10, "k1", "v1", "k3", "v3", "k5", "v5", "k7", "v7")
		t.mustScan(ctx, client, "k2", "", 2, "k3", "v3", "k5", "v5")
		t.mustScan(ctx, client, "k2", "", 3, "k3", "v3", "k5", "v5", "k7", "v7")
		t.mustScan(ctx, client, "", "k1", 1)
		t.mustScan(ctx, client, "k1", "k3", 2, "k1", "v1")
		t.mustScan(ctx, client, "k1", "k5", 10, "k1", "v1", "k3", "v3")
		t.mustScan(ctx, client, "k1", "k5\x00", 10, "k1", "v1", "k3", "v3", "k5", "v5")
		t.mustScan(ctx, client, "k5\x00", "k5\x00\x00", 10)
	}

	check()
	t.mustSplit(ctx, cluster, "k", "k2")
	check()
	t.mustSplit(ctx, cluster, "k2", "k5")
	check()
}

var _ = validator.Test("delete range", []string{"rawkv.batch-put", "rawkv.scan", "rawkv.delete-range"}, testRawKV{}.testDeleteRange)

func (t testRawKV) testDeleteRange(ctx validator.ExecContext) {
	cluster, client := t.newClient(ctx)
	defer cluster.Close()
	defer client.Close()

	var keys, values []string
	for _, i := range []byte("abcd") {
		for j := byte('0'); j < byte('9'); j++ {
			keys = append(keys, string([]byte{i, j}))
			values = append(values, string([]byte{'v', i, j}))
		}
	}

	list := func() []string {
		var kv []string
		for i := range keys {
			kv = append(kv, keys[i], values[i])
		}
		return kv
	}

	cut := func(start, end string) {
		k, v := keys[:0], values[:0]
		for i := range keys {
			if keys[i] < start || keys[i] >= end {
				k = append(k, keys[i])
				v = append(v, values[i])
			}
		}
		keys, values = k, v
	}

	t.mustBatchPut(ctx, client, keys, values)

	check := func(start, end string) {
		t.mustDeleteRange(ctx, client, start, end)
		cut(start, end)
		t.mustScan(ctx, client, "", "", len(keys), list()...)
	}

	check("0", "1")
	check("b", "c0")
	check("c11", "c12")
	check("d0", "d0")
	check("c5", "d5")
	check("a", "z")
}

func (t testRawKV) mustNotExist(ctx validator.ExecContext, client *stub.RawClientStub, key string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	v, err := client.Get([]byte(key))
	ctx.AssertNil(err)
	ctx.AssertEQ(len(v), 0)
}

func (t testRawKV) mustBatchNotExist(ctx validator.ExecContext, client *stub.RawClientStub, keys []string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	values, err := client.BatchGet(bss(keys...))
	ctx.AssertNil(err)
	ctx.AssertEQ(len(values), len(keys), "length of values does not match keys")
	for _, v := range values {
		ctx.AssertEQ(len(v), 0)
	}
}

func (t testRawKV) mustGet(ctx validator.ExecContext, client *stub.RawClientStub, key, value string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	val, err := client.Get([]byte(key))
	ctx.AssertNil(err)
	ctx.AssertEQ(string(val), value)
}

func (t testRawKV) mustBatchGet(ctx validator.ExecContext, client *stub.RawClientStub, keys, values []string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	vals, err := client.BatchGet(bss(keys...))
	ctx.AssertNil(err)
	ctx.AssertEQ(len(vals), len(values))
	for i, val := range vals {
		ctx.AssertEQ(string(val), values[i])
	}
}

func (t testRawKV) mustPut(ctx validator.ExecContext, client *stub.RawClientStub, key, value string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	err := client.Put([]byte(key), []byte(value))
	ctx.AssertNil(err)
}

func (t testRawKV) mustBatchPut(ctx validator.ExecContext, client *stub.RawClientStub, keys, values []string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	err := client.BatchPut(bss(keys...), bss(values...))
	ctx.AssertNil(err)
}

func (t testRawKV) mustDelete(ctx validator.ExecContext, client *stub.RawClientStub, key string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	err := client.Delete([]byte(key))
	ctx.AssertNil(err)
}

func (t testRawKV) mustBatchDelete(ctx validator.ExecContext, client *stub.RawClientStub, keys []string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	err := client.BatchDelete(bss(keys...))
	ctx.AssertNil(err)
}

func (t testRawKV) mustScan(ctx validator.ExecContext, client *stub.RawClientStub, start, end string, limit int, expect ...string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	keys, values, err := client.Scan([]byte(start), []byte(end), limit)
	ctx.AssertNil(err)
	ctx.AssertEQ(len(keys)*2, len(expect))
	for i := range keys {
		ctx.AssertEQ(string(keys[i]), expect[i*2])
		ctx.AssertEQ(string(values[i]), expect[i*2+1])
	}
}

func (t testRawKV) mustDeleteRange(ctx validator.ExecContext, client *stub.RawClientStub, start, end string) {
	ctx.AddCallerDepth(1)
	defer ctx.AddCallerDepth(-1)
	err := client.DeleteRange([]byte(start), []byte(end))
	ctx.AssertNil(err)
}

func (t testRawKV) mustSplit(ctx validator.ExecContext, cluster *mocktikv.Cluster, start, end string) {
	// TODO: Not supported by mock-tikv now.
}

func bss(ss ...string) [][]byte {
	bss := make([][]byte, len(ss))
	for i := range ss {
		bss[i] = []byte(ss[i])
	}
	return bss
}
