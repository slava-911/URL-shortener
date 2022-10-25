package logging

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Entry
}

func (l *Logger) ExtraFields(fields map[string]interface{}) *Logger {
	return &Logger{l.WithFields(fields)}
}

func GetLogger(level string) *Logger {
	//once.Do(func() {
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.Fatalln(err)
	}

	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
		},
		DisableColors: false,
		FullTimestamp: true,
	}

	l.SetOutput(os.Stdout)
	l.SetLevel(logrusLevel)

	return &Logger{logrus.NewEntry(l)}
	//})
}
