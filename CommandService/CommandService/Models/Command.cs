namespace CommandService.Models
{
    public class Command
    {
        public long Id { get; set; }
        public string? Fty { get; set; }
        public string? IpAddress { get; set; }
        public string? Sender { get; set; }
    }
}