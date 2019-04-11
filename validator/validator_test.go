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

package validator_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/tikv/client-validator/validator"
)

var _ = validator.Feature("A", "describe A", nil, func(_ validator.ExecContext) validator.FeatureStatus {
	return validator.FeaturePass
})

var _ = validator.Feature("B", "describe B", nil, func(_ validator.ExecContext) validator.FeatureStatus {
	return validator.FeatureNotImplemented
})

var _ = validator.Feature("C", "describe C", nil, func(_ validator.ExecContext) validator.FeatureStatus {
	return validator.FeatureFail
})

var _ = validator.Feature("D", "describe D", []string{"A"}, func(_ validator.ExecContext) validator.FeatureStatus {
	return validator.FeaturePass
})

var _ = validator.Feature("E", "describe E", []string{"A", "B"}, func(_ validator.ExecContext) validator.FeatureStatus {
	return validator.FeatureNotImplemented
})

var _ = validator.Feature("F", "describe F", []string{"A", "D"}, func(ctx validator.ExecContext) validator.FeatureStatus {
	ctx.Log("foobar")
	return validator.FeaturePass
})

var _ = validator.Story("E and F", "E", "F")

var _ = validator.Feature("G", "describe G", nil, func(_ validator.ExecContext) validator.FeatureStatus {
	return validator.FeaturePass
})

var _ = validator.Test("test G", []string{"G"}, func(ctx validator.ExecContext) {
	ctx.Fail("G has a bug")
})

func TestValidator(t *testing.T) {
	validator.LogTimeFormat = "[TIME]"
	validator.LogFileLine = false

	expect := validator.Report{
		Stories: []validator.StoryReport{
			{
				Description: "E and F",
				Features: []validator.FeatureReport{
					{
						Key:         "E",
						Description: "describe E",
						Status:      validator.FeatureSkip,
					},
					{
						Key:         "F",
						Description: "describe F",
						Status:      validator.FeaturePass,
						Records: []validator.Recorder{
							{
								Description: "check F(describe F)",
								Logs: []string{
									"[TIME] foobar",
									"[TIME] check finish. success=true, feature.status=PASS",
								},
								Success: true,
							},
						},
					},
				},
			},
		},
		Features: []validator.FeatureReport{
			{
				Key:         "A",
				Description: "describe A",
				Status:      validator.FeaturePass,
				Records: []validator.Recorder{
					{
						Description: "check A(describe A)",
						Logs: []string{
							"[TIME] check finish. success=true, feature.status=PASS",
						},
						Success: true,
					},
				},
			},
			{
				Key:         "B",
				Description: "describe B",
				Status:      validator.FeatureNotImplemented,
				Records: []validator.Recorder{
					{
						Description: "check B(describe B)",
						Logs: []string{
							"[TIME] check finish. success=true, feature.status=NOT_IMPL",
						},
						Success: true,
					},
				},
			},
			{
				Key:         "C",
				Description: "describe C",
				Status:      validator.FeatureFail,
				Records: []validator.Recorder{
					{
						Description: "check C(describe C)",
						Logs: []string{
							"[TIME] check finish. success=false, feature.status=FAIL",
						},
						Success: false,
					},
				},
			},
			{
				Key:         "D",
				Description: "describe D",
				Status:      validator.FeaturePass,
				Records: []validator.Recorder{
					{
						Description: "check D(describe D)",
						Logs: []string{
							"[TIME] check finish. success=true, feature.status=PASS",
						},
						Success: true,
					},
				},
			},
			{
				Key:         "G",
				Description: "describe G",
				Status:      validator.FeatureDefect,
				Records: []validator.Recorder{
					{
						Description: "check G(describe G)",
						Logs: []string{
							"[TIME] check finish. success=true, feature.status=PASS",
						},
						Success: true,
					},
					{
						Description: "test G",
						Logs: []string{
							"[TIME] G has a bug",
							"[TIME] test finish. success=false, feature.status=DEFECT",
						},
					},
				},
			},
		},
	}

	expectJson, _ := json.Marshal(expect)
	reportJson, _ := json.Marshal(validator.RunAll())

	if !bytes.Equal(expectJson, reportJson) {
		t.Logf("expect: %s", expectJson)
		t.Logf("got   : %s", reportJson)
		t.FailNow()
	}
}
