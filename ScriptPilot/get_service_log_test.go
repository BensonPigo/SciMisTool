package main

import (
	pb "ScriptPilot/proto/taskexecutor"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type dummyServer struct{ server }

func TestGetServiceLog_NotFound(t *testing.T) {
	config.ServiceLogPaths = map[string]string{}
	s := &dummyServer{}
	_, err := s.GetServiceLog(context.Background(), &pb.GetServiceLogRequest{ServiceName: "foo", Date: "20210101"})
	assert.Error(t, err)
}
