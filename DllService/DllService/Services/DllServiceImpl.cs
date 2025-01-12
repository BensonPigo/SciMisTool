using Grpc.Core;
using Microsoft.AspNetCore.Razor.TagHelpers;
using System.Threading.Tasks;

public class DllServiceImpl : DllService.DllServiceBase
{
    private readonly DllLoaderService _dllLoaderService;

    public DllServiceImpl(DllLoaderService dllLoaderService)
    {
        _dllLoaderService = dllLoaderService;
    }

    public override Task<MethodResponse> InvokeMethod(MethodRequest request,ServerCallContext context)
    {
        var result = _dllLoaderService.InvokeMethod(
            request.DllPath,
            request.TypeName,
            request.MethodName,
            request.Parameters.ToArray()
        );

        return Task.FromResult(
            new MethodResponse
            {
                Result = result?.ToString() ?? "null"
            }
        );
    }
}