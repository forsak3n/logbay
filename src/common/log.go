package common

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
)

var fields = make([]string, 0)

func init() {
	textFormatter := new(prefixed.TextFormatter)
	textFormatter.DisableColors = true
	textFormatter.ForceFormatting = true
	textFormatter.FullTimestamp = true

	log.SetFormatter(textFormatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
}

func SetContextFields(s []string) {
	fields = s
}

func ContextLogger(ctx context.Context) *log.Entry {

	if len(fields) == 0 {
		return log.NewEntry(log.StandardLogger())
	}

	logFields := make(log.Fields)

	for _, v := range fields {
		if fieldValue, ok := ctx.Value(v).(string); ok && len(fieldValue) > 0 {
			logFields[v] = fieldValue
		}
	}

	return log.WithFields(logFields)
}
