using Microsoft.AspNetCore.Mvc;
using System.Threading.Tasks;
using CommandService.Services;
using CommandService.Models;

[ApiController]
[Route("api/[controller]")]
public class CommandController : ControllerBase
{
    private readonly ExecuteService _grpcInvokeService;

    public CommandController(ExecuteService grpcInvokeService)
    {
        _grpcInvokeService = grpcInvokeService;
    }


    [HttpPost]
    public async Task<IActionResult> InvokeMethod([FromBody] MethodRequestDto requestDto)
    {
        //var request = new MethodRequest
        //{
        //    DllPath = requestDto.DllPath ?? string.Empty,
        //    TypeName = requestDto.TypeName?? string.Empty,
        //    MethodName = requestDto.MethodName?? string.Empty,
        //};
        //request.Parameters.AddRange(requestDto.Parameters == null ? new List<string>() : requestDto.Parameters.ToArray());

        //// 調用 Service 層的 gRPC 邏輯
        //var result = await _grpcInvokeService.InvokeMethodAsync(
        //    request.DllPath,
        //    request.TypeName,
        //    request.MethodName,
        //    request.Parameters.ToArray()
        //);

        // 返回結果給客戶端
        //return Ok(new { Result = result });
        return null;
    }
}