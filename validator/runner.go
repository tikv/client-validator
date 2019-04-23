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

import "fmt"

// RunAll runs all registered checkers and tests then determine status of
// features.
func RunAll() Report {
	runner := newTestRunner()
	runner.run()
	return runner.report()
}

type featureInfo struct {
	conf    featureConf
	status  FeatureStatus
	records []Recorder
}

type testRunner struct {
	features    []*featureInfo
	featuresMap map[string]*featureInfo
	stories     []storyConf
	tests       []testConf
}

func newTestRunner() *testRunner {
	confMu.Lock()
	defer confMu.Unlock()

	runner := &testRunner{
		featuresMap: make(map[string]*featureInfo),
		stories:     append(storyConfs[:0:0], storyConfs...),
		tests:       append(testConfs[:0:0], testConfs...),
	}

	for _, conf := range featureConfs {
		f := &featureInfo{
			conf:   conf,
			status: FeatureSkip,
		}
		runner.features = append(runner.features, f)
		runner.featuresMap[conf.key] = f
	}
	return runner
}

func (r *testRunner) run() {
	for f := r.nextFeature(); f != nil; f = r.nextFeature() {
		r.runFeatureChecker(f)
	}
	for _, t := range r.tests {
		if r.checkRequiredFeatures(t.features) {
			r.runTest(t)
		}
	}
}

func (r *testRunner) nextFeature() *featureInfo {
	for _, f := range r.features {
		if f.status == FeatureSkip && r.checkRequiredFeatures(f.conf.requiredFeatures) {
			return f
		}
	}
	return nil
}

func (r *testRunner) checkRequiredFeatures(features []string) bool {
	for _, key := range features {
		f, ok := r.featuresMap[key]
		if !ok {
			panic("required feature not found: " + key)
		}
		if f.status != FeaturePass && f.status != FeatureDefect {
			return false
		}
	}
	return true
}

func (r *testRunner) runFeatureChecker(f *featureInfo) {
	recorder := newRecorder(fmt.Sprintf("check %s(%s)", f.conf.key, f.conf.description))
	f.status = r.callChecker(recorder, f.conf.checkF)
	recorder.Log("check finish. success=%v, feature.status=%s", recorder.Success, f.status)
	f.records = append(f.records, *recorder)
}

func (r *testRunner) callChecker(recorder *Recorder, f func(ExecContext) FeatureStatus) (status FeatureStatus) {
	defer func() {
		if err := recover(); err != nil {
			recorder.log(false, "%v", err)
			status = FeatureFail
		}
	}()

	if f != nil {
		status = f(execContext{Recorder: recorder, asserter: &asserter{}})
		recorder.Success = (status == FeaturePass || status == FeatureNotImplemented)
	} else {
		status = FeatureSkip
	}
	return
}

func (r *testRunner) runTest(t testConf) {
	recorder := newRecorder(t.description)
	r.callTest(recorder, t.testF)
	for _, key := range t.features {
		f := r.featuresMap[key]
		if !recorder.Success && f.status == FeaturePass {
			f.status = FeatureDefect
		}
		recorder2 := recorder.copy()
		recorder2.Log("test finish. success=%v, feature.status=%s", recorder2.Success, f.status)
		f.records = append(f.records, *recorder2)
	}
}

func (r *testRunner) callTest(recorder *Recorder, f func(ExecContext)) {
	defer func() {
		if err := recover(); err != nil {
			recorder.log(false, "%v", err)
		}
	}()

	if f != nil {
		f(execContext{Recorder: recorder, asserter: &asserter{}})
		recorder.Success = true // Success if not panic
	}
}
