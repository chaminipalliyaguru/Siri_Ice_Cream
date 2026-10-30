using SiriIceCream.Desktop.Helpers;
using SiriIceCream.Desktop.Models;

namespace SiriIceCream.Desktop.Services;

/// <summary>
/// Who is logged in, and what their role allows.
/// The rules match the backend exactly. The backend still checks every
/// request (403); these rules only let the app explain things earlier.
/// </summary>
public class Session
{
    public const string Owner = "owner";
    public const string Manager = "manager";
    public const string Cashier = "cashier";

    public Session(User user)
    {
        User = user;
    }

    public User User { get; }

    public string Role => User.Role;

    /// <summary>"Owner", "Manager" or "Cashier" for showing on screen.</summary>
    public string RoleDisplayName => Role.Length == 0 ? "" : char.ToUpper(Role[0]) + Role[1..];

    /// <summary>
    /// Up to two letters for the avatar circle:
    /// "Shop Cashier" -> "SC", "Nimal" -> "N", "" -> "?".
    /// </summary>
    public string Initials => NameHelper.Initials(User.Name);

    // Items: owner and manager add, everyone edits, only owner deletes.
    public bool CanAddItem => Role is Owner or Manager;
    public bool CanEditItem => Role is Owner or Manager or Cashier;
    public bool CanDeleteItem => Role is Owner;

    // Customers: everyone adds and edits, only owner deletes.
    public bool CanAddCustomer => Role is Owner or Manager or Cashier;
    public bool CanEditCustomer => Role is Owner or Manager or Cashier;
    public bool CanDeleteCustomer => Role is Owner;

    /// <summary>Message shown when the user tries something their role does not allow.</summary>
    public string NotAllowedMessage(string action, string allowedRoles) =>
        $"As a {RoleDisplayName}, you cannot {action}. Only {allowedRoles} can do this.";
}
