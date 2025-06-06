# ScriptPilot

This service exposes a gRPC API for executing scripts.

## RPC Methods

### GetServiceLog

Fetch log content produced by another service.

```
rpc GetServiceLog(GetServiceLogRequest) returns (GetServiceLogResponse)
```

`GetServiceLogRequest` requires the service name and a date (used to build the log filename). The response contains the raw log content.
