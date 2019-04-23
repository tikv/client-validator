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

package validator

import "sync"

var (
	// LogTimeFormat defines the time format in checker/test execute log.
	LogTimeFormat = "2006-01-02 15:04:05.999"
	// LogFileLine indicates if output code position of caller in log.
	LogFileLine = true
)

// FeatureStatus is the status of a feature.
type FeatureStatus string

// Feature status.
const (
	// The feature has passed checkers and all tests.
	FeaturePass FeatureStatus = "PASS"
	// The feature is not implemented.
	FeatureNotImplemented FeatureStatus = "NOT_IMPL"
	// The feature has incorrect behavior. (failed to pass checker)
	FeatureFail FeatureStatus = "FAIL"
	// The feature has passed the checker but failed in some tests. (may have bugs)
	FeatureDefect FeatureStatus = "DEFECT"
	// The feature is not tested. (prerequisite features not supported)
	FeatureSkip FeatureStatus = "SKIP"
)

// RegisterFeature defines a feature. All features should be registered before main().
func RegisterFeature(key, description string, requireFeatures []string, checkF func(ExecContext) FeatureStatus) string {
	confMu.Lock()
	defer confMu.Unlock()
	for _, feature := range featureConfs {
		if feature.key == key {
			panic("duplicated feature key: " + key)
		}
	}
	featureConfs = append(featureConfs, featureConf{
		key:              key,
		description:      description,
		requiredFeatures: requireFeatures,
		checkF:           checkF,
	})
	return key
}

// RegisterStory is a list of features. The features will be placed together in the test report.
func RegisterStory(description string, features ...string) struct{} {
	confMu.Lock()
	defer confMu.Unlock()
	storyConfs = append(storyConfs, storyConf{description: description, features: features})
	return struct{}{}
}

// RegisterTest defines a test case. If testF panics, all features will be marked as
// Defect. All tests should be registered before main().
func RegisterTest(description string, features []string, testF func(ExecContext)) struct{} {
	confMu.Lock()
	defer confMu.Unlock()
	testConfs = append(testConfs, testConf{
		description: description,
		features:    features,
		testF:       testF,
	})
	return struct{}{}
}

type featureConf struct {
	key              string
	description      string
	requiredFeatures []string
	checkF           func(ExecContext) FeatureStatus
}

type storyConf struct {
	description string
	features    []string
}

type testConf struct {
	description string
	features    []string
	testF       func(ExecContext)
}

var (
	confMu       sync.RWMutex
	featureConfs []featureConf
	testConfs    []testConf
	storyConfs   []storyConf
)
