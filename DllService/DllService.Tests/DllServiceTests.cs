using NUnit.Framework;
using Moq;
using Grpc.Core;
using System.Threading.Tasks;

[TestFixture]
public class DllServiceTests
{
    private DllServiceImpl _service;
    private Mock<DllLoaderService> _mockDllLoaderService;

    [SetUp]
    public void Setup()
    {
        // 模擬 DllLoaderService
        _mockDllLoaderService = new Mock<DllLoaderService>();

        // 如果需要模擬特定方法的行為，可以使用以下語法：
        // _mockDllLoaderService.Setup(x => x.SomeMethod()).Returns(someValue);

        // 初始化服務，傳入模擬對象
        //_service = new DllServiceImpl(_mockDllLoaderService.Object);
        var b = new DllLoaderService();
        b.LoadDllsFromDirectories(new List<string>() { @"D:\SourceCode\PMSToolApplication\SciApiGateway\SciApiGatewayTests45\bin\Debug" });
        _service = new DllServiceImpl(b);
    }

    [Test]
    public async Task InvokeMethod_ShouldReturnExpectedResult()
    {
        // Arrange
        var mockContext = new Mock<ServerCallContext>(); // 模擬 ServerCallContext
        var request = new MethodRequest
        {
            DllPath = @"D:\SourceCode\PMSToolApplication\SciApiGateway\SciApiGatewayTests45\bin\Debug\SciApiGateway.dll",
            TypeName = "SciApiGateway.Component.HashToolComponent",
            MethodName = "ComputeSHA3_256Hash",
            Parameters = { "param1" }
        };

        // Act
        var response = await _service.InvokeMethod(request, mockContext.Object);

        // Assert
        Assert.That(response, !Is.Null); // 驗證 response 不是 null
        //Assert.AreEqual("DLL: C:\\Server\\MyLibrary.dll, Method: MyMethod", response.Result);
    }
}
