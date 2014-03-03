package picolog

import (
	"io/ioutil"
	"strings"
	"testing"
	"regexp"
	"log/syslog"
	"os"
)

func TestLogger(t *testing.T) {
	fo, err := ioutil.TempFile(".", "picolog_test_out")
	fname := fo.Name()
	defer os.Remove(fname)
	if err != nil {
		t.Errorf("Could not open tempfile: %v", err)
	}
	l := NewLogger(syslog.LOG_INFO, "test", fo)
	l.Infof("logging things")
	fo.Seek(0, 0)
	out, err := ioutil.ReadAll(fo)
	if err != nil {
		t.Errorf("Could not read tempfile: %v", err)
	}
	pattern := regexp.MustCompile(`\[test\]\s+[\s\d:/.]+logging things\s+`)
	if !pattern.Match(out) {
		t.Errorf("Wanted a match for %s, got %s", pattern, out)
	}
}

func TestParseLogLevel(t *testing.T) {
	var logLevelValid = []struct {
		in string
		out syslog.Priority
	}{
		{"debug", syslog.LOG_DEBUG},
		{"info", syslog.LOG_INFO},
		{"warning", syslog.LOG_WARNING},
		{"emerg", syslog.LOG_EMERG},
		{"notice", syslog.LOG_NOTICE},
	}
	testOneLevel := func(in string, out syslog.Priority) {
		res, err := ParseLogLevel(in)
		if err != nil {
			t.Errorf("%v", err)
		}
		if res != out {
			t.Errorf("Wanted %v, got %v", out, res)
		}
	}
	for _, test := range logLevelValid {
		testOneLevel(test.in, test.out)
		testOneLevel(strings.ToUpper(test.in), test.out)
	}
	_, err := ParseLogLevel("invalid")
	if err == nil {
		t.Errorf("Successfully parsed invalid log level.")
	}
}
