using Moq;
using Grpc.Core;
using System.Text.Json;
using Assert = NUnit.Framework.Assert;

[TestFixture]
public class DllServiceTests
{
    private GrpcDllServiceImpl _service;
    private Mock<DllLoaderService> _mockDllLoaderService;

    [SetUp]
    public void Setup()
    {
        // 模擬 DllLoaderService
        _mockDllLoaderService = new Mock<DllLoaderService>();

        // 如果需要模擬特定方法的行為，可以使用以下語法：
        // _mockDllLoaderService.Setup(x => x.SomeMethod()).Returns(someValue);

        // 初始化服務
        _service = new GrpcDllServiceImpl();
    }


    [Test]
    public async Task InvokeMethod_ShouldReturnExpectedResult()
    {
        // Arrange
        var mockContext = new Mock<ServerCallContext>(); // 模擬 ServerCallContext

        // 準備 MethodRequest 物件
        var request = new MethodRequest
        {
            DirectoryPath = @"C:\Git\Quality\Quality\BusinessLogicLayer\bin\Debug",
            DllPath = @"C:\Git\Quality\Quality\BusinessLogicLayer\bin\Debug\BusinessLogicLayer.dll",
            TypeName = "BusinessLogicLayer.Service.FinalInspection.QueryService",
            MethodName = "GetFinalInspectionReport",
        };

        // 將參數序列化為 JSON 並添加到 request.Parameters
        var objParameter = new QA_R51_ViewModel()
        {
            SP = "24XXX",
            IsBI = false,
            StartInspectionDate = DateTime.Now
        };
        var jsonParameter = $@"SPSCH25010622";
        request.Parameters.Add(jsonParameter);

        // Act
        var response = await _service.GrpcExecuteDllMethod(request, mockContext.Object);

        // Assert
        Assert.That(response, !Is.Null); // 驗證 response 不是 null
        //Assert.AreEqual("DLL: C:\\Server\\MyLibrary.dll, Method: MyMethod", response.Result);
    }
    public class QA_R51_ViewModel
    {
        /// <inheritdoc/>

        /// <inheritdoc/>
        public string SP { get; set; }

        /// <inheritdoc/>
        public bool IsBI { get; set; }

        /// <inheritdoc/>
        public DateTime? StartInspectionDate { get; set; }

    }
}
