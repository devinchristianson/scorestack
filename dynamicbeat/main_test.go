package main

// This file is mandatory as otherwise the dynamicbeat.test binary is not generated correctly.

import (
	"flag"
	"testing"

	"github.com/scorestack/scorestack/dynamicbeat/cmd"
)

var systemTest *bool

func init() {
	systemTest = flag.Bool("systemTest", false, "Set to true when running system tests")

	cmd.RootCmd.PersistentFlags().AddGoFlag(flag.CommandLine.Lookup("systemTest"))
}

// Test started when the test binary is started. Only calls main.
func TestSystem(t *testing.T) {

	if *systemTest {
		main()
	}
}
