package logging

import (
	"github.com/fluent/fluent-logger-golang/fluent"
)

type FluentWriter struct {
	*fluent.Fluent
}

const (
	ERROR_DEFAULT_VALUE int = 0
)

func (fw *FluentWriter) Write(p []byte) (n int, err error) {
	msg := map[string]string{"msg": string(p)}
	err = fw.Post("dss.logs", msg)

	if err != nil {
		return ERROR_DEFAULT_VALUE, err
	}
	return len(p), nil
}
