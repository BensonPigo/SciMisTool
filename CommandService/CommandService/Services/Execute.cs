using Microsoft.AspNetCore.Mvc;

namespace CommandService.Services
{
    public class Execute
    {
        private readonly DllService.DllServiceClient _grpcClient;

        public Execute(DllService.DllServiceClient grpcClient)
        {
            _grpcClient = grpcClient;
        }

        [HttpPost("invoke")]
        public async Task<IActionResult> InvokeMethod([FromBody] MethodRequest request)
        {
            // 調用 gRPC 服務
            var response = await _grpcClient.InvokeMethodAsync(request);

            // 返回結果給 Web API 的客戶端
            //return Ok(response.Result);
            return null;

        }
    }
}
