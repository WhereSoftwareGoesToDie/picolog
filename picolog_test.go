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
	l := NewLogger(LogInfo, "test", fo)
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

func TestSubLogger(t *testing.T) {
	fo, err := ioutil.TempFile(".", "picolog_sublogger_test_out")
	fname := fo.Name()
	defer os.Remove(fname)
	if err != nil {
		t.Errorf("Could not open tempfile: %v", err)
	}
	l := NewLogger(LogInfo, "test1", fo)
	l2 := l.NewSubLogger("test2")
	l3 := l2.NewSubLogger("test3")
	// Ordering is not a bug
	l.Infof("one")
	l3.Infof("two")
	l2.Infof("three")
	fo.Seek(0, 0)
	out, err := ioutil.ReadAll(fo)
	if err != nil {
		t.Errorf("Could not read tempfile: %v", err)
	}
	pattern := `\[test1\]\s+[\s\d:/.]+one\s+`
	pattern += `\[test1\]\[test2\]\[test3\]\s+[\s\d:/.]+two\s+`
	pattern += `\[test1\]\[test2\]\s+[\s\d:/.]+three\s+`
	re := regexp.MustCompile(pattern)
	if !re.Match(out) {
		t.Errorf("Wanted a match for %s, got %s", pattern, out)
	}
}

func TestParseLogLevel(t *testing.T) {
	var logLevelValid = []struct {
		in string
		out LogLevel
	}{
		{"debug", LogLevel(syslog.LOG_DEBUG)},
		{"info", LogLevel(syslog.LOG_INFO)},
		{"warning", LogLevel(syslog.LOG_WARNING)},
		{"emerg", LogLevel(syslog.LOG_EMERG)},
		{"notice", LogLevel(syslog.LOG_NOTICE)},
	}
	testOneLevel := func(in string, out LogLevel) {
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
