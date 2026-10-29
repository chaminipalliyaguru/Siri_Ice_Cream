namespace SiriIceCream.Desktop.Models;

/// <summary>The API's answer to a successful login.</summary>
public class LoginResult
{
    public string Token { get; set; } = "";
    public User User { get; set; } = new();
}
