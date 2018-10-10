package globals

import "os"
import "path"

import "github.com/romana/rlog"
import "github.com/sirupsen/logrus"

type RLOG_HOOK struct { //  rlog HOOK
	dummy     string
	formatter *logrus.TextFormatter
}

var HOOK RLOG_HOOK

// setup default logging to current directory
func Init_rlog() {

	HOOK.formatter = new(logrus.TextFormatter)
	HOOK.formatter.DisableColors = true
	HOOK.formatter.DisableTimestamp = true

	if os.Getenv("RLOG_LOG_LEVEL") == "" {
		os.Setenv("RLOG_LOG_LEVEL", "WARN") // default logging in debug mode
	}

	if os.Getenv("RLOG_LOG_FILE") == "" {
		exename, _ := os.Executable()
		filename := path.Base(exename) + ".log"
		os.Setenv("RLOG_LOG_FILE", filename) // default log file name
	}

	if os.Getenv("RLOG_LOG_STREAM") == "" {
		os.Setenv("RLOG_LOG_STREAM", "NONE") // do not log to stdout/stderr
	}

	if os.Getenv("RLOG_CALLER_INFO") == "" {
		os.Setenv("RLOG_CALLER_INFO", "RLOG_CALLER_INFO") // log caller info
	}

	//os.Setenv("RLOG_TRACE_LEVEL", "10") user can request tracing
	//os.Setenv("RLOG_LOG_LEVEL", "DEBUG")
	rlog.UpdateEnv()

}

// log logrus messages to rlog
func (hook *RLOG_HOOK) Fire(entry *logrus.Entry) error {
	msg, err := hook.formatter.Format(entry)
	if err == nil {
		rlog.Infof(string(msg)) // log to file
	}
	return nil
}

// Levels returns configured log levels., we log everything
func (hook *RLOG_HOOK) Levels() []logrus.Level {
	return logrus.AllLevels
}
