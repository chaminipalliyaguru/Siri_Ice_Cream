using System.Net;

namespace SiriIceCream.Desktop.Services;

/// <summary>
/// An error from the API (or from reaching it). The Message is safe to show to the user.
/// </summary>
public class ApiException : Exception
{
    /// <summary>HTTP status code, or null if the server could not be reached.</summary>
    public HttpStatusCode? StatusCode { get; }

    public ApiException(string message, HttpStatusCode? statusCode = null, Exception? inner = null)
        : base(message, inner)
    {
        StatusCode = statusCode;
    }

    public bool IsUnauthorized => StatusCode == HttpStatusCode.Unauthorized; // 401
    public bool IsForbidden => StatusCode == HttpStatusCode.Forbidden;       // 403
}
