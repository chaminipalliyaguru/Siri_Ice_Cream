using SiriIceCream.Desktop.Helpers;

namespace SiriIceCream.Desktop.Models;

public class Customer
{
    public int Id { get; set; }
    public string Name { get; set; } = "";
    public string Email { get; set; } = "";
    public string Phone { get; set; } = "";

    // ---- Values calculated for the screen (not sent by the API) ----

    /// <summary>Letters for the avatar circle in the table, for example "NP".</summary>
    public string Initials => NameHelper.Initials(Name);

    /// <summary>True when a phone number was entered.</summary>
    public bool HasPhone => !string.IsNullOrWhiteSpace(Phone);

    /// <summary>The phone number, or a dash when there is none.</summary>
    public string PhoneDisplay => HasPhone ? Phone : "—";
}

/// <summary>The data sent to the API to create or update a customer.</summary>
public class CustomerInput
{
    public string Name { get; set; } = "";
    public string Email { get; set; } = "";
    public string Phone { get; set; } = "";
}
