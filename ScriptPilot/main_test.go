// server_test.go
package main

import (
	"ScriptPilot/proto/taskexecutor"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testServer struct {
	taskexecutor.UnimplementedTaskExecutorServer
}

func TestExecuteTask(t *testing.T) {
	server := &testServer{}

	t.Run("Success", func(t *testing.T) {
		req := &taskexecutor.TaskRequest{
			FactoryId: "F123",
			TaskName:  "TestTask",
		}
		resp, err := server.ExecuteTask(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "Task executed successfully", resp.Message)
		assert.Equal(t, "Task output", resp.Output)
	})

	t.Run("Missing TaskName", func(t *testing.T) {
		req := &taskexecutor.TaskRequest{
			FactoryId: "F123",
		}
		resp, err := server.ExecuteTask(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "Task name is required", resp.Message)
		assert.Equal(t, "Invalid request", resp.Error)
	})
}
