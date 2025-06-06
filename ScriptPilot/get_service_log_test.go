package main

import (
	pb "ScriptPilot/proto/taskexecutor"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyServer struct{ server }

func TestGetServiceLog_NotFound(t *testing.T) {
	config.ServiceLogPaths = map[string]string{}
	s := &dummyServer{}
	_, err := s.GetServiceLog(context.Background(), &pb.GetServiceLogRequest{ServiceName: "FtyBitoTpe", Date: "20250604"})
	assert.Error(t, err)
}
