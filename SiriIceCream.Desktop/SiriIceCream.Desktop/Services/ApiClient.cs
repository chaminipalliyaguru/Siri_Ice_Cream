using System.Net;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Net.Http.Json;
using System.Text.Json;
using SiriIceCream.Desktop.Models;

namespace SiriIceCream.Desktop.Services;

/// <summary>
/// Calls the Siri Ice Cream REST API with HttpClient.
/// It adds the JWT token to every request and turns error responses.
/// </summary>
public class ApiClient : IApiClient
{
    private readonly HttpClient _http;
    private string? _token;

    // The Go API uses snake_case JSON names (category_id), C# uses CategoryId.
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        PropertyNameCaseInsensitive = true,
    };

    public ApiClient(string baseUrl, HttpClient? http = null)
    {
        _http = http ?? new HttpClient { Timeout = TimeSpan.FromSeconds(15) };
        BaseUrl = baseUrl;
    }

    public string BaseUrl { get; set; }

    // ---------------- Authentication ----------------

    public async Task<LoginResult> LoginAsync(string username, string password)
    {
        var result = await SendAsync<LoginResult>(HttpMethod.Post, "/api/auth/login",
            new { username, password });
        _token = result.Token;
        return result;
    }

    public async Task LogoutAsync()
    {
        try
        {
            await SendAsync<object>(HttpMethod.Post, "/api/auth/logout");
        }
        finally
        {
            _token = null; // forget the token even if the server could not be reached
        }
    }

    //  Categories

    public Task<List<Category>> GetCategoriesAsync() =>
        SendAsync<List<Category>>(HttpMethod.Get, "/api/categories");

    //  Items 

    public Task<List<Item>> GetItemsAsync(string? search = null) =>
        SendAsync<List<Item>>(HttpMethod.Get, "/api/items" + SearchQuery(search));

    public Task<Item> CreateItemAsync(ItemInput input) =>
        SendAsync<Item>(HttpMethod.Post, "/api/items", input);

    public Task<Item> UpdateItemAsync(int id, ItemInput input) =>
        SendAsync<Item>(HttpMethod.Put, $"/api/items/{id}", input);

    public Task DeleteItemAsync(int id) =>
        SendAsync<object>(HttpMethod.Delete, $"/api/items/{id}");

    // Customers 

    public Task<List<Customer>> GetCustomersAsync(string? search = null) =>
        SendAsync<List<Customer>>(HttpMethod.Get, "/api/customers" + SearchQuery(search));

    public Task<Customer> CreateCustomerAsync(CustomerInput input) =>
        SendAsync<Customer>(HttpMethod.Post, "/api/customers", input);

    public Task<Customer> UpdateCustomerAsync(int id, CustomerInput input) =>
        SendAsync<Customer>(HttpMethod.Put, $"/api/customers/{id}", input);

    public Task DeleteCustomerAsync(int id) =>
        SendAsync<object>(HttpMethod.Delete, $"/api/customers/{id}");

    //  Helpers 

    /// <summary>Builds "?name=..." for searches, or "" when there is no search text.</summary>
    private static string SearchQuery(string? search) =>
        string.IsNullOrWhiteSpace(search) ? "" : "?name=" + Uri.EscapeDataString(search.Trim());

    /// <summary>
    /// Sends one request and reads the JSON answer.
    /// Any problem becomes an ApiException with a message for the user.
    /// </summary>
    private async Task<T> SendAsync<T>(HttpMethod method, string path, object? body = null)
    {
        using var request = new HttpRequestMessage(method, BaseUrl.TrimEnd('/') + path);
        if (_token != null)
        {
            request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", _token);
        }
        if (body != null)
        {
            request.Content = JsonContent.Create(body, options: JsonOptions);
        }

        HttpResponseMessage response;
        try
        {
            response = await _http.SendAsync(request);
        }
        catch (HttpRequestException ex)
        {
            throw new ApiException($"Cannot reach the server at {BaseUrl}. Is the backend running?", null, ex);
        }
        catch (TaskCanceledException ex)
        {
            throw new ApiException($"The server at {BaseUrl} did not answer in time.", null, ex);
        }
        catch (Exception ex) when (ex is InvalidOperationException or UriFormatException)
        {
            throw new ApiException($"The server address \"{BaseUrl}\" is not valid.", null, ex);
        }

        using (response)
        {
            if (!response.IsSuccessStatusCode)
            {
                string message = await ReadErrorMessageAsync(response);
                throw new ApiException(message, response.StatusCode);
            }

            if (typeof(T) == typeof(object))
            {
                return default!; // the caller does not need the body
            }

            try
            {
                var result = await response.Content.ReadFromJsonAsync<T>(JsonOptions);
                return result ?? throw new ApiException("The server sent an empty answer.");
            }
            catch (JsonException ex)
            {
                throw new ApiException("The server sent an answer the app could not read.", response.StatusCode, ex);
            }
        }
    }

    /// <summary>Reads {"error": "..."} from an error response.</summary>
    private static async Task<string> ReadErrorMessageAsync(HttpResponseMessage response)
    {
        try
        {
            var error = await response.Content.ReadFromJsonAsync<ErrorResponse>(JsonOptions);
            if (!string.IsNullOrWhiteSpace(error?.Error))
            {
                return error.Error;
            }
        }
        catch (Exception)
        {
            // Not JSON (for example an HTML error page). Fall back below.
        }

        return response.StatusCode switch
        {
            HttpStatusCode.Unauthorized => "You are not logged in.",
            HttpStatusCode.Forbidden => "You do not have permission to do this.",
            HttpStatusCode.NotFound => "Not found.",
            _ => $"The server returned an error ({(int)response.StatusCode}).",
        };
    }

    private class ErrorResponse
    {
        public string? Error { get; set; }
    }
}
