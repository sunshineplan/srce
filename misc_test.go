package main

import (
	"reflect"
	"testing"
)

func TestParseCmd(t *testing.T) {
	testcase := []struct {
		cmd string
		res []string
	}{
		{
			"/service/srce/start",
			[]string{"service", "srce", "start"},
		},
		{
			"/service+srce_start",
			[]string{"service", "srce", "start"},
		},
		{
			"/service%20srce%20start",
			[]string{"service", "srce", "start"},
		},
		{
			"/~bin~bash",
			[]string{"/bin/bash"},
		},
	}

	for _, tc := range testcase {
		res, err := parseCmd(tc.cmd)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(tc.res, res) {
			t.Errorf("expected %q; got %q", tc.res, res)
		}
	}
}
