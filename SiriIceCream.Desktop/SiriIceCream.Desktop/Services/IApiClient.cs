using SiriIceCream.Desktop.Models;

namespace SiriIceCream.Desktop.Services;

public interface IApiClient
{

    string BaseUrl { get; set; }

    // Authentication
    Task<LoginResult> LoginAsync(string username, string password);
    Task LogoutAsync();

    // Categories
    Task<List<Category>> GetCategoriesAsync();

    // Items
    Task<List<Item>> GetItemsAsync(string? search = null);
    Task<Item> CreateItemAsync(ItemInput input);
    Task<Item> UpdateItemAsync(int id, ItemInput input);
    Task DeleteItemAsync(int id);

    // Customers
    Task<List<Customer>> GetCustomersAsync(string? search = null);
    Task<Customer> CreateCustomerAsync(CustomerInput input);
    Task<Customer> UpdateCustomerAsync(int id, CustomerInput input);
    Task DeleteCustomerAsync(int id);
}
