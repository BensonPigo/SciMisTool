using System.Threading.Tasks;
using Moq;
using NUnit.Framework;
using Assert = NUnit.Framework.Assert;
using GrpcServiceClient;
using Grpc.Net.Client;

namespace CommandService.Services.Tests
{
    [TestClass()]
    public class ExecuteServiceTests
    {
        private ExecuteService _executeService;
        private Mock<GrpcDllService.GrpcDllServiceClient> _mockGrpcClient;


        [SetUp]
        public void Setup()
        {
            // 初始化 Mock gRPC Client
            //_mockGrpcClient = new Mock<DllService.DllServiceClient>();

            // 初始化要測試的服務
            _executeService = new ExecuteService();
        }

        [Test]
        public void ExecuteCommandTest()
        {
            //using var channel = GrpcChannel.ForAddress("http://localhost:5239");
            //var client = new DllService.DllServiceClient(channel);

            // Arrange
            var directoryPath = @"C:\Git\Quality\Quality\BusinessLogicLayer\bin\Debug";
            var dllPath = @"C:\Git\Quality\Quality\BusinessLogicLayer\bin\Debug\BusinessLogicLayer.dll";
            var typeName = "BusinessLogicLayer.Service.FinalInspection.QueryService";
            var methodName = "GetFinalInspectionReport";
            var parameters = new[]{"SPSCH25010622"};

            //var parameters = new[]
            //            { $@"
            //{{
            //   ""SP"":""24XXX"",
            //   ""IsBI"":false,
            //   ""StartInspectionDate"":""2025-01-15T10:47:37.9982734+08:00""
            //}}
            //"
            //            };

            ////  Web API 的 DTO Mapping至 gRPC 的 Request，對應的是dll 的method、參數等等
            //var grpcRequest = new MethodRequest
            //{
            //    DirectoryPath = directoryPath,
            //    DllPath = dllPath,
            //    TypeName = typeName,
            //    MethodName = methodName,
            //};
            //grpcRequest.Parameters.AddRange(parameters);

            // 3. 呼叫服務端的方法
            //var reply = client.ExecuteDllMethod(grpcRequest);

            var reply = _executeService.ExecuteCommand(directoryPath, dllPath, typeName, methodName, parameters);


            Assert.Pass();
            //Assert.Fail();
        }
    }
}