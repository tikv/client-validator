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
	"flag"
	"strings"

	"github.com/tikv/client-validator/validator"
)

var (
	mockTiKVAddr    = flag.String("mock-tikv", "http://127.0.0.1:2378", "mock-tikv server address")
	clientProxyAddr = flag.String("client-proxy", "http://127.0.0.1:8080", "client proxy server address")
)

func errToFeatureStatus(err error) validator.FeatureStatus {
	if err == nil {
		return validator.FeaturePass
	}
	if strings.Index(strings.ToLower(err.Error()), "not implemented") > 0 {
		return validator.FeatureNotImplemented
	}
	return validator.FeatureFail
}
