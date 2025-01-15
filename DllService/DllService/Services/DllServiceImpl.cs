using Grpc.Core;
using Google.Protobuf.WellKnownTypes;
using Microsoft.AspNetCore.Razor.TagHelpers;
using System.Threading.Tasks;
using Grpc.Net.Client.Configuration;
using System.Reflection.Metadata;
using System.Reflection;
using System.Security.Cryptography;
using System.Text.Json;

public class GrpcDllServiceImpl : GrpcDllService.GrpcDllServiceBase
{
    public GrpcDllServiceImpl()
    {
    }

    public override Task<MethodResponse> GrpcExecuteDllMethod(MethodRequest request, ServerCallContext context)
    {
        DllLoaderService _dllLoaderService = new DllLoaderService();
        _dllLoaderService.LoadDllsFromDirectories(new List<string>() { request.DllPath, });

        var method = _dllLoaderService.GetMethod(
            request.DllPath,
            request.TypeName,
            request.MethodName);

        if (method == null)
        {
            throw new ArgumentException($"Method {request.MethodName} not found in type {request.TypeName}.");
        }

        var parametersInfo = method.GetParameters();
        if (parametersInfo.Length != request.Parameters.Count)
        {
            throw new ArgumentException($"Parameter count mismatch. Expected: {parametersInfo.Length}, Provided: {request.Parameters.Count}");
        }

        // 動態解析 JSON 參數並轉換為所需類型
        var parameters = new object[parametersInfo.Length];
        for (int i = 0; i < parametersInfo.Length; i++)
        {
            var parameterType = parametersInfo[i].ParameterType;
            var jsonValue = request.Parameters[i];

            try
            {
                // 判斷參數是string int這種系統類別，還是自定義的類別
                if (parameterType.IsClass && parameterType.Namespace != "System")
                {
                    parameters[i] = string.IsNullOrEmpty(jsonValue)
                        ? null
                        : JsonSerializer.Deserialize(jsonValue, parameterType);
                }
                else
                {
                    parameters[i] = jsonValue;
                }
            }
            catch (Exception ex)
            {
                throw new ArgumentException(
                    $"Failed to deserialize parameter {i} for method '{request.MethodName}' in type '{request.TypeName}'. " +
                    $"Parameter value: {jsonValue}. Expected type: {parameterType}.",
                    ex
                );
            }
        }

        // 調用方法
        var result = _dllLoaderService.InvokeMethod(
            request.DllPath,
            request.TypeName,
            request.MethodName,
            parameters
        );

        return Task.FromResult(
            new MethodResponse
            {
                Result = result?.ToString() ?? "null",
                Timestamp = Timestamp.FromDateTime(DateTime.UtcNow)
            });
    }

}