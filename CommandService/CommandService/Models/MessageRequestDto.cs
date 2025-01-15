namespace CommandService.Models
{
    public class MethodRequestDto
    {
        public string? DllPath { get; set; }
        public string? TypeName { get; set; }
        public string? MethodName { get; set; }
        public List<string>? Parameters { get; set; }
    }
}
