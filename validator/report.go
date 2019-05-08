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

// FeatureReport is test report for a feature.
type FeatureReport struct {
	Key         string        `json:"key"`
	Description string        `json:"description"`
	Status      FeatureStatus `json:"status"`
	Records     []Recorder    `json:"records"`
}

// StoryReport is the test report for a story.
type StoryReport struct {
	Description string          `json:"description,omitempty"`
	Features    []FeatureReport `json:"features,omitempty"`
}

// Report contains test results.
type Report struct {
	Stories  []StoryReport   `json:"stories,omitempty"`
	Features []FeatureReport `json:"features,omitempty"`
}

func (r *testRunner) report() Report {
	var report Report
	reported := make(map[string]struct{})
	for _, s := range r.stories {
		report.Stories = append(report.Stories, r.reportStory(s))
		for _, k := range s.features {
			reported[k] = struct{}{}
		}
	}
	for _, f := range r.features {
		if _, ok := reported[f.conf.key]; !ok {
			report.Features = append(report.Features, r.reportFeature(f.conf.key))
		}
	}
	return report
}

func (r *testRunner) reportStory(conf storyConf) StoryReport {
	report := StoryReport{Description: conf.description}
	for _, f := range conf.features {
		report.Features = append(report.Features, r.reportFeature(f))
	}
	return report
}

func (r *testRunner) reportFeature(key string) FeatureReport {
	f := r.featuresMap[key]
	if f == nil {
		panic("feature not found: " + key)
	}

	return FeatureReport{
		Key:         f.conf.key,
		Description: f.conf.description,
		Status:      f.status,
		Records:     f.records,
	}
}
