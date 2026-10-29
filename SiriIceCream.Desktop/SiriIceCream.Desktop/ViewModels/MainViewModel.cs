using SiriIceCream.Desktop.Commands;
using SiriIceCream.Desktop.Services;

namespace SiriIceCream.Desktop.ViewModels;

/// Logic for the main window.
public class MainViewModel : ViewModelBase
{
    private readonly IApiClient _api;
    private readonly Session _session;
    private object _currentPage;
    private bool _ended; // true once the session has ended (so we only end it once)

    public MainViewModel(IApiClient api, Session session, IDialogService dialogs)
    {
        _api = api;
        _session = session;

        // The pages of the app. The real screens replace these placeholders in the next steps.
        InventoryPage = new PlaceholderViewModel("Inventory", "The inventory screen is added in the next step.");
        CustomersPage = new PlaceholderViewModel("Customers", "The customers screen is added in a later step.");
        _currentPage = InventoryPage;

        ShowInventoryCommand = new RelayCommand(() => NavigateTo(InventoryPage));
        ShowCustomersCommand = new RelayCommand(() => NavigateTo(CustomersPage));
        LogoutCommand = new AsyncRelayCommand(LogoutAsync);
    }

    /// Raised when the user signs out or the session expires.
    public event Action<string?>? LoggedOut;

    // ---------------- Pages and navigation ----------------

    public object InventoryPage { get; }
    public object CustomersPage { get; }

    /// <summary>The page shown in the main area (the window picks the right view for it).</summary>
    public object CurrentPage
    {
        get => _currentPage;
        private set
        {
            if (SetProperty(ref _currentPage, value))
            {
                OnPropertyChanged(nameof(IsInventoryActive));
                OnPropertyChanged(nameof(IsCustomersActive));
                OnPropertyChanged(nameof(PageTitle));
                OnPropertyChanged(nameof(PageSubtitle));
            }
        }
    }

    /// <summary>True when Inventory is shown (highlights the menu item).</summary>
    public bool IsInventoryActive => CurrentPage == InventoryPage;

    /// <summary>True when Customers is shown (highlights the menu item).</summary>
    public bool IsCustomersActive => CurrentPage == CustomersPage;

    public string PageTitle => IsInventoryActive ? "Inventory" : "Customers";

    public string PageSubtitle => IsInventoryActive
        ? "Manage ice cream items, prices and stock."
        : "Manage customer details and contact information.";

    public RelayCommand ShowInventoryCommand { get; }
    public RelayCommand ShowCustomersCommand { get; }
    public AsyncRelayCommand LogoutCommand { get; }

    // ---------------- User card ----------------

    public string UserName => _session.User.Name;
    public string RoleDisplayName => _session.RoleDisplayName;
    public string Initials => _session.Initials;
    public string ServerDisplay => _api.BaseUrl;

    // ---------------- Actions ----------------

    /// <summary>Loads the data for the pages when the window opens.</summary>
    public Task InitializeAsync()
    {
        // The Inventory and Customers screens load their data here (added in the next steps).
        return Task.CompletedTask;
    }

    private void NavigateTo(object page)
    {
        CurrentPage = page;
    }

    private async Task LogoutAsync()
    {
        try
        {
            await _api.LogoutAsync();
        }
        catch (ApiException)
        {
            // Even if the server cannot be reached, sign out locally.
        }
        EndSession(null);
    }

    /// <summary>Raises LoggedOut once, even if several pages notice that the session expired.</summary>
    private void EndSession(string? message)
    {
        if (_ended)
        {
            return;
        }
        _ended = true;
        LoggedOut?.Invoke(message);
    }
}
