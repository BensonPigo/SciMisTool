using Microsoft.AspNetCore.Mvc;
using CommandService.Models;
using Grpc.Net.Client;
using GrpcServiceClient;

namespace CommandService.Services
{
    public class ExecuteService
    {
        //private readonly DllService.DllServiceClient _grpcClient;

        public ExecuteService()
        {
            //_grpcClient = grpcClient;
        }

        public  Task<MethodResponse> ExecuteCommand(string directoryPath, string dllPath, string typeName, string methodName, string[] parameters)
         {
            using var channel = GrpcChannel.ForAddress("http://localhost:5239");

            // 2. 初始化服務的 Client
            var client = new GrpcDllService.GrpcDllServiceClient(channel);

            //  Web API 的 DTO Mapping至 gRPC 的 Request，對應的是dll 的method、參數等等
            var grpcRequest = new MethodRequest
            {
                DirectoryPath = directoryPath,
                DllPath = dllPath,
                TypeName = typeName,
                MethodName = methodName,
            };
            grpcRequest.Parameters.AddRange(parameters);

            // 3. 呼叫服務端的方法
            var reply =  client.GrpcExecuteDllMethod(grpcRequest);


            if (reply == null)
            {
                //var r = reply.Result;
            }
            // 返回 gRPC 方法的結果
            return null;

        }
    }
}
