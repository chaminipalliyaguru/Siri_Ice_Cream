namespace SiriIceCream.Desktop.ViewModels;

/// <summary>Small helpers for messages shown to the user.</summary>
public static class Messages
{

    /// Makes an API message look like a sentence: "price must be greater than 0" -> "Price must be greater than 0."
    public static string Friendly(string message)
    {
        if (string.IsNullOrWhiteSpace(message))
        {
            return "Something went wrong.";
        }

        string text = char.ToUpper(message[0]) + message[1..];
        return text.EndsWith('.') || text.EndsWith('?') || text.EndsWith('!') ? text : text + ".";
    }

    public const string SessionExpired = "Your session has expired. Please log in again.";

    public const string Unexpected = "Something unexpected went wrong. Please try again.";
}
