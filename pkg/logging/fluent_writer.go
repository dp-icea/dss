package logging

import (
	"github.com/fluent/fluent-logger-golang/fluent"
)

type FluentWriter struct {
	*fluent.Fluent
}

func (fw *FluentWriter) Write(p []byte) (n int, err error) {
	//TODO: use tag from env file
	msg := map[string]string{"msg": string(p)}
	err = fw.Post("dss.logs", msg)

	//TODO: implement descriptive message for when post is not successful
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
