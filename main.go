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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
	_ "github.com/tikv/client-validator/tests" // import tests
	"github.com/tikv/client-validator/validator"
)

var (
	showRecord  = flag.String("record", "failed", "none | failed | all")
	showLog     = flag.Bool("show-log", false, "show test logs in report")
	outputStyle = flag.String("output", "console", "console | text | json")
)

func main() {
	flag.Parse()
	report := validator.RunAll()
	trimReport(&report)
	printReport(&report)
}

func trimReport(report *validator.Report) {
	for i := range report.Stories {
		for j := range report.Stories[i].Features {
			trimFeatureReport(&report.Stories[i].Features[j])
		}
	}
	for i := range report.Features {
		trimFeatureReport(&report.Features[i])
	}
}

func trimFeatureReport(report *validator.FeatureReport) {
	records := report.Records[:0]
	for _, r := range report.Records {
		if *showRecord == "all" || (*showRecord == "failed" && !r.Success) {
			if !*showLog {
				r.Logs = nil
			}
			records = append(records, r)
		}
	}
	report.Records = records
}

func printReport(report *validator.Report) {
	switch *outputStyle {
	case "json":
		printJSON(report)
	case "text":
		printText(report)
	default:
		printConsole(report)
	}
}

func printJSON(report *validator.Report) {
	data, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(data))
}

func printText(report *validator.Report) {
	hr := strings.Repeat("-", 80)

	printFeature := func(feature *validator.FeatureReport) {
		fmt.Printf("  + [%v] %v: %v\n", feature.Status, feature.Key, feature.Description)
		for _, r := range feature.Records {
			stateText := "PASS"
			if !r.Success {
				stateText = "FAIL"
			}
			fmt.Printf("    - [%v] %v\n", stateText, r.Description)
			for _, l := range r.Logs {
				fmt.Printf("      $ %s\n", l)
			}
		}
	}

	for _, s := range report.Stories {
		fmt.Println(hr)
		fmt.Println("# " + s.Description)
		for _, feature := range s.Features {
			printFeature(&feature)
		}
	}
	for _, feature := range report.Features {
		fmt.Println(hr)
		printFeature(&feature)
	}
}

func printConsole(report *validator.Report) {
	hr := strings.Repeat("-", 80)

	printFeature := func(feature *validator.FeatureReport) {
		info := aurora.Bold(aurora.Blue(fmt.Sprintf("%s: %s", feature.Key, feature.Description)))
		fmt.Printf("  [%v] %v\n", colorizeStatus(string(feature.Status)), info)
		for _, r := range feature.Records {
			stateText := "PASS"
			if !r.Success {
				stateText = "FAIL"
			}
			fmt.Printf("    [%v] %v\n", colorizeStatus(stateText), aurora.Bold(aurora.Cyan(r.Description)))
			for _, l := range r.Logs {
				fmt.Print("      ")
				fmt.Println(aurora.Gray(l))
			}
		}
	}

	for _, s := range report.Stories {
		fmt.Println(hr)
		fmt.Println(aurora.Bold(aurora.Magenta("# " + s.Description)))
		for _, feature := range s.Features {
			printFeature(&feature)
		}
	}
	for _, feature := range report.Features {
		fmt.Println(hr)
		printFeature(&feature)
	}
}

func colorizeStatus(text string) aurora.Value {
	switch text {
	case "PASS":
		return aurora.Green(text)
	case "NOT_IMPL", "SKIP":
		return aurora.Gray(text)
	case "FAIL", "DEFECT":
		return aurora.Red(text)
	}
	return nil
}
