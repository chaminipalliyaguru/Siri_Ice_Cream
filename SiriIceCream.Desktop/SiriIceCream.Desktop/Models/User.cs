namespace SiriIceCream.Desktop.Models;

/// <summary>A staff member who can log in (returned by the API).</summary>
public class User
{
    public int Id { get; set; }
    public string Name { get; set; } = "";
    public string Username { get; set; } = "";
    public string Role { get; set; } = "";
}
