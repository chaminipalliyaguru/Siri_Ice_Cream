namespace SiriIceCream.Desktop.Helpers;

/// <summary>Small helpers for people's names.</summary>
public static class NameHelper
{
    /// <summary>
    /// Up to two letters for an avatar circle:
    /// "Shop Cashier" -> "SC", "Nimal" -> "N", "" -> "?".
    /// </summary>
    public static string Initials(string? name)
    {
        var words = (name ?? "").Split(' ', StringSplitOptions.RemoveEmptyEntries);
        return words.Length switch
        {
            0 => "?",
            1 => words[0][..1].ToUpper(),
            _ => (words[0][..1] + words[^1][..1]).ToUpper(),
        };
    }
}
