namespace SiriIceCream.Desktop.Models;

/// <summary>A shop customer (returned by the API).</summary>
public class Customer
{
    public int Id { get; set; }
    public string Name { get; set; } = "";
    public string Email { get; set; } = "";
    public string Phone { get; set; } = "";
}

/// <summary>The data sent to the API to create or update a customer.</summary>
public class CustomerInput
{
    public string Name { get; set; } = "";
    public string Email { get; set; } = "";
    public string Phone { get; set; } = "";
}
