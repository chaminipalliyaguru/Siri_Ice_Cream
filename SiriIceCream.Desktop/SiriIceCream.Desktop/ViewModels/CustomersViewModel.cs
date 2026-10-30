using System.Collections.ObjectModel;
using System.Net.Mail;
using SiriIceCream.Desktop.Commands;
using SiriIceCream.Desktop.Models;
using SiriIceCream.Desktop.Services;

namespace SiriIceCream.Desktop.ViewModels;

/// <summary>
/// Logic for the Customers screen: the summary cards, the search bar,
/// the customer table and the add / edit form.
/// </summary>
public class CustomersViewModel : ScreenViewModelBase
{
    private readonly IApiClient _api;
    private readonly Session _session;
    private readonly IDialogService _dialogs;

    private string _searchText = "";
    private string _appliedSearch = "";   // the search used for the current list
    private Customer? _selectedCustomer;
    private int? _editingCustomerId;      // null = the form adds a new customer
    private string _formName = "";
    private string _formEmail = "";
    private string _formPhone = "";

    public CustomersViewModel(IApiClient api, Session session, IDialogService dialogs)
    {
        _api = api;
        _session = session;
        _dialogs = dialogs;

        LoadCommand = new AsyncRelayCommand(LoadAsync);
        SearchCommand = new AsyncRelayCommand(SearchAsync);
        ClearSearchCommand = new AsyncRelayCommand(ClearSearchAsync);
        NewCommand = new RelayCommand(StartNewCustomer);
        CancelCommand = new RelayCommand(CancelEdit);
        SaveCommand = new AsyncRelayCommand(SaveAsync);
        DeleteCommand = new AsyncRelayCommand(DeleteAsync);
    }

    // ---------------- Data for the screen ----------------

    public ObservableCollection<Customer> Customers { get; } = new();

    public string SearchText
    {
        get => _searchText;
        set => SetProperty(ref _searchText, value);
    }

    /// <summary>The row selected in the table. Selecting a row loads it into the form.</summary>
    public Customer? SelectedCustomer
    {
        get => _selectedCustomer;
        set
        {
            if (SetProperty(ref _selectedCustomer, value) && value != null)
            {
                LoadIntoForm(value);
            }
        }
    }

    /// <summary>True when the table is empty (shows the "No customers found" message).</summary>
    public bool HasNoCustomers => Customers.Count == 0;

    // ---------------- Summary cards ----------------

    public int TotalCustomers => Customers.Count;
    public int WithPhoneCount => Customers.Count(c => c.HasPhone);
    public int WithoutPhoneCount => Customers.Count(c => !c.HasPhone);

    /// <summary>Explains what the cards count: everyone, or only the search results.</summary>
    public string SummaryCaption => _appliedSearch.Length == 0
        ? "All customers of the shop"
        : $"Customers matching \"{_appliedSearch}\"";

    // ---------------- The form ----------------

    public string FormName { get => _formName; set => SetProperty(ref _formName, value); }
    public string FormEmail { get => _formEmail; set => SetProperty(ref _formEmail, value); }
    public string FormPhone { get => _formPhone; set => SetProperty(ref _formPhone, value); }

    public bool IsEditing => _editingCustomerId != null;

    /// <summary>Heading above the form.</summary>
    public string FormTitle => IsEditing ? $"Edit customer #{_editingCustomerId}" : "Add a new customer";

    public string FormSubtitle => IsEditing
        ? "Change the details and click Save."
        : "Fill in the details and click Save.";

    // ---------------- What this role can see ----------------
    // Buttons the user is not allowed to use are hidden, not just blocked.

    /// <summary>The "New customer" button (every role may add customers).</summary>
    public bool CanAddCustomers => _session.CanAddCustomer;

    /// <summary>The form is shown while editing, or for roles that may add customers.</summary>
    public bool ShowForm => IsEditing || _session.CanAddCustomer;

    /// <summary>The "Delete customer" button (owners, while a customer is selected).</summary>
    public bool ShowDeleteButton => IsEditing && _session.CanDeleteCustomer;

    // ---------------- Commands (bound to buttons) ----------------

    public AsyncRelayCommand LoadCommand { get; }
    public AsyncRelayCommand SearchCommand { get; }
    public AsyncRelayCommand ClearSearchCommand { get; }
    public RelayCommand NewCommand { get; }
    public RelayCommand CancelCommand { get; }
    public AsyncRelayCommand SaveCommand { get; }
    public AsyncRelayCommand DeleteCommand { get; }

    /// <summary>Loads the customers from the server.</summary>
    public async Task LoadAsync()
    {
        ClearMessages();
        await RunAsync(ReloadCustomersAsync);
    }

    private async Task SearchAsync()
    {
        ClearMessages();
        bool ok = await RunAsync(ReloadCustomersAsync);
        if (ok && _appliedSearch.Length > 0)
        {
            ShowStatus(Customers.Count == 1
                ? $"1 customer matches \"{_appliedSearch}\"."
                : $"{Customers.Count} customers match \"{_appliedSearch}\".");
        }
    }

    private async Task ClearSearchAsync()
    {
        SearchText = "";
        ClearMessages();
        await RunAsync(ReloadCustomersAsync);
    }

    /// <summary>Empties the form so a new customer can be entered.</summary>
    private void StartNewCustomer()
    {
        ClearMessages();
        if (!_session.CanAddCustomer)
        {
            ShowError(_session.NotAllowedMessage("add customers", "Owners, Managers and Cashiers"));
            return;
        }
        ClearForm();
    }

    /// <summary>Stops editing and empties the form.</summary>
    private void CancelEdit()
    {
        ClearMessages();
        ClearForm();
    }

    private async Task SaveAsync()
    {
        ClearMessages();

        // 1. Is this role allowed to do it?
        if (!IsEditing && !_session.CanAddCustomer)
        {
            ShowError(_session.NotAllowedMessage("add customers", "Owners, Managers and Cashiers"));
            return;
        }
        if (IsEditing && !_session.CanEditCustomer)
        {
            ShowError(_session.NotAllowedMessage("edit customers", "Owners, Managers and Cashiers"));
            return;
        }

        // 2. Is the form filled in correctly?
        if (!TryReadForm(out var input, out string problem))
        {
            ShowError(problem);
            return;
        }

        // 3. Send it to the API.
        Customer? saved = null;
        bool wasEditing = IsEditing;
        bool ok = await RunAsync(async () =>
        {
            saved = wasEditing
                ? await _api.UpdateCustomerAsync(_editingCustomerId!.Value, input)
                : await _api.CreateCustomerAsync(input);
            await ReloadCustomersAsync();
        });

        if (ok && saved != null)
        {
            ClearForm(); // done: empty the form, ready for the next one
            ShowStatus(wasEditing ? $"Customer \"{saved.Name}\" was updated." : $"Customer \"{saved.Name}\" was added.");
        }
    }

    private async Task DeleteAsync()
    {
        ClearMessages();

        if (!_session.CanDeleteCustomer)
        {
            ShowError(_session.NotAllowedMessage("delete customers", "Owners"));
            return;
        }
        if (!IsEditing)
        {
            ShowError("Please select a customer in the table first.");
            return;
        }

        int id = _editingCustomerId!.Value;
        string name = FormName;
        if (!_dialogs.Confirm($"Delete the customer \"{name}\"?\nThis cannot be undone.", "Delete customer"))
        {
            return;
        }

        bool ok = await RunAsync(async () =>
        {
            await _api.DeleteCustomerAsync(id);
            await ReloadCustomersAsync();
        });

        if (ok)
        {
            ClearForm();
            ShowStatus($"Customer \"{name}\" was deleted.");
        }
    }

    // ---------------- Helpers ----------------

    private async Task ReloadCustomersAsync()
    {
        string search = SearchText.Trim();
        var customers = await _api.GetCustomersAsync(search);

        _appliedSearch = search;
        Customers.Clear();
        foreach (var customer in customers)
        {
            Customers.Add(customer);
        }
        OnCustomersChanged();
    }

    /// <summary>Tells the screen to refresh everything that is calculated from the list.</summary>
    private void OnCustomersChanged()
    {
        OnPropertyChanged(nameof(HasNoCustomers));
        OnPropertyChanged(nameof(TotalCustomers));
        OnPropertyChanged(nameof(WithPhoneCount));
        OnPropertyChanged(nameof(WithoutPhoneCount));
        OnPropertyChanged(nameof(SummaryCaption));
    }

    private void LoadIntoForm(Customer customer)
    {
        _editingCustomerId = customer.Id;
        FormName = customer.Name;
        FormEmail = customer.Email;
        FormPhone = customer.Phone;
        OnFormModeChanged();
    }

    private void ClearForm()
    {
        SelectedCustomer = null;
        _editingCustomerId = null;
        FormName = "";
        FormEmail = "";
        FormPhone = "";
        OnFormModeChanged();
    }

    private void OnFormModeChanged()
    {
        OnPropertyChanged(nameof(IsEditing));
        OnPropertyChanged(nameof(FormTitle));
        OnPropertyChanged(nameof(FormSubtitle));
        OnPropertyChanged(nameof(ShowForm));
        OnPropertyChanged(nameof(ShowDeleteButton));
    }

    /// <summary>
    /// Checks the form (the same rules as the backend) and builds the data to send.
    /// Returns false with a message if something is wrong.
    /// </summary>
    private bool TryReadForm(out CustomerInput input, out string problem)
    {
        input = new CustomerInput();
        problem = "";

        if (string.IsNullOrWhiteSpace(FormName))
        {
            problem = "Please enter the customer's name.";
            return false;
        }
        if (string.IsNullOrWhiteSpace(FormEmail))
        {
            problem = "Please enter the customer's email.";
            return false;
        }
        if (!IsValidEmail(FormEmail.Trim()))
        {
            problem = "Please enter a valid email address, for example nimal@example.com.";
            return false;
        }

        input = new CustomerInput
        {
            Name = FormName.Trim(),
            Email = FormEmail.Trim(),
            Phone = FormPhone.Trim(),
        };
        return true;
    }

    private static bool IsValidEmail(string email)
    {
        try
        {
            var address = new MailAddress(email);
            return address.Address == email && email.Contains('.');
        }
        catch (FormatException)
        {
            return false;
        }
    }
}
