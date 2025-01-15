using CommandService.Services;

var builder = WebApplication.CreateBuilder(args);

//builder.Services.AddGrpcClient<DllService.DllServiceClient>(o =>
//{
//    o.Address = new Uri("http://localhost:5239"); // 替換為 gRPC 服務的真實地址
//});


// Add services to the container.
builder.Services.AddControllers();
//builder.Services.AddDbContext<CommandContext>(opt=>opt.);

builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

// 註冊 Service 到 DI 容器
builder.Services.AddScoped<ExecuteService>();

var app = builder.Build();

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    // app.UseSwaggerUI();
    app.UseSwaggerUI(options =>
    {
        options.SwaggerEndpoint("/swagger/v1/swagger.json", "My API v1");
        options.RoutePrefix = string.Empty; // 根路徑打開 Swagger UI
    });
}

app.UseHttpsRedirection();
app.UseAuthorization();

app.MapControllers();
app.Run();