using System.Collections.ObjectModel;
using System.Globalization;
using SiriIceCream.Desktop.Commands;
using SiriIceCream.Desktop.Models;
using SiriIceCream.Desktop.Services;

namespace SiriIceCream.Desktop.ViewModels;

/// <summary>
/// Logic for the Inventory screen: the summary cards, the search bar,the item table and the add / edit form.
/// </summary>
public class ItemsViewModel : ScreenViewModelBase
{
    private readonly IApiClient _api;
    private readonly Session _session;
    private readonly IDialogService _dialogs;

    private string _searchText = "";
    private string _appliedSearch = "";   // the search used for the current list
    private Item? _selectedItem;
    private int? _editingItemId;          // null = the form adds a new item
    private string _formName = "";
    private Category? _formCategory;
    private string _formPrice = "";
    private string _formQuantity = "";

    public ItemsViewModel(IApiClient api, Session session, IDialogService dialogs)
    {
        _api = api;
        _session = session;
        _dialogs = dialogs;

        LoadCommand = new AsyncRelayCommand(LoadAsync);
        SearchCommand = new AsyncRelayCommand(SearchAsync);
        ClearSearchCommand = new AsyncRelayCommand(ClearSearchAsync);
        NewCommand = new RelayCommand(StartNewItem);
        CancelCommand = new RelayCommand(CancelEdit);
        SaveCommand = new AsyncRelayCommand(SaveAsync);
        DeleteCommand = new AsyncRelayCommand(DeleteAsync);
    }

    // Data for the screen 

    public ObservableCollection<Item> Items { get; } = new();
    public ObservableCollection<Category> Categories { get; } = new();

    public string SearchText
    {
        get => _searchText;
        set => SetProperty(ref _searchText, value);
    }

    /// <summary>The row selected in the table. Selecting a row loads it into the form.</summary>
    public Item? SelectedItem
    {
        get => _selectedItem;
        set
        {
            if (SetProperty(ref _selectedItem, value) && value != null)
            {
                LoadIntoForm(value);
            }
        }
    }

    /// <summary>True when the table is empty (shows the "No items found" message).</summary>
    public bool HasNoItems => Items.Count == 0;

    // Summary cards 

    public int TotalItems => Items.Count;
    public int TotalUnits => Items.Sum(i => i.Quantity);
    public int LowStockCount => Items.Count(i => i.IsLowStock || i.IsOutOfStock);
    public string StockValueDisplay => "Rs. " + Items.Sum(i => i.Price * i.Quantity).ToString("N2", CultureInfo.InvariantCulture);

    /// <summary>Explains what the cards count: everything, or only the search results.</summary>
    public string SummaryCaption => _appliedSearch.Length == 0
        ? "All items in the shop"
        : $"Items matching \"{_appliedSearch}\"";

    // The form 

    public string FormName { get => _formName; set => SetProperty(ref _formName, value); }
    public Category? FormCategory { get => _formCategory; set => SetProperty(ref _formCategory, value); }
    public string FormPrice { get => _formPrice; set => SetProperty(ref _formPrice, value); }
    public string FormQuantity { get => _formQuantity; set => SetProperty(ref _formQuantity, value); }

    public bool IsEditing => _editingItemId != null;

    /// <summary>Heading above the form.</summary>
    public string FormTitle => IsEditing ? $"Edit item #{_editingItemId}" : "Add a new item";

    public string FormSubtitle => IsEditing
        ? "Change the details and click Save."
        : "Fill in the details and click Save.";

    /// <summary>Explains what the logged-in role may do on this screen.</summary>
    public string PermissionHint => _session.Role switch
    {
        Session.Owner => "Owner: you can view, add, edit and delete items.",
        Session.Manager => "Manager: you can view, add and edit items. Deleting items is not allowed.",
        _ => "Cashier: you can view and edit items. Adding and deleting items is not allowed.",
    };

    // ---------------- Commands (bound to buttons) ----------------

    public AsyncRelayCommand LoadCommand { get; }
    public AsyncRelayCommand SearchCommand { get; }
    public AsyncRelayCommand ClearSearchCommand { get; }
    public RelayCommand NewCommand { get; }
    public RelayCommand CancelCommand { get; }
    public AsyncRelayCommand SaveCommand { get; }
    public AsyncRelayCommand DeleteCommand { get; }

    /// <summary>Loads the categories (for the drop-down) and the items.</summary>
    public async Task LoadAsync()
    {
        ClearMessages();
        await RunAsync(async () =>
        {
            var categories = await _api.GetCategoriesAsync();
            Categories.Clear();
            foreach (var category in categories)
            {
                Categories.Add(category);
            }
            await ReloadItemsAsync();
        });
    }

    private async Task SearchAsync()
    {
        ClearMessages();
        bool ok = await RunAsync(ReloadItemsAsync);
        if (ok && _appliedSearch.Length > 0)
        {
            ShowStatus(Items.Count == 1
                ? $"1 item matches \"{_appliedSearch}\"."
                : $"{Items.Count} items match \"{_appliedSearch}\".");
        }
    }

    private async Task ClearSearchAsync()
    {
        SearchText = "";
        ClearMessages();
        await RunAsync(ReloadItemsAsync);
    }

    /// <summary>Empties the form so a new item can be entered.</summary>
    private void StartNewItem()
    {
        ClearMessages();
        if (!_session.CanAddItem)
        {
            ShowError(_session.NotAllowedMessage("add items", "Owners and Managers"));
            return;
        }

        ClearForm();
        FormCategory = Categories.FirstOrDefault();
        FormQuantity = "0";
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
        if (!IsEditing && !_session.CanAddItem)
        {
            ShowError(_session.NotAllowedMessage("add items", "Owners and Managers") +
                      " Select an item in the table to edit it.");
            return;
        }
        if (IsEditing && !_session.CanEditItem)
        {
            ShowError(_session.NotAllowedMessage("edit items", "Owners, Managers and Cashiers"));
            return;
        }

        // 2. Is the form filled in correctly?
        if (!TryReadForm(out var input, out string problem))
        {
            ShowError(problem);
            return;
        }

        // 3. Send it to the API.
        Item? saved = null;
        bool wasEditing = IsEditing;
        bool ok = await RunAsync(async () =>
        {
            saved = wasEditing
                ? await _api.UpdateItemAsync(_editingItemId!.Value, input)
                : await _api.CreateItemAsync(input);
            await ReloadItemsAsync();
        });

        if (ok && saved != null)
        {
            SelectedItem = Items.FirstOrDefault(i => i.Id == saved.Id);
            if (SelectedItem == null)
            {
                LoadIntoForm(saved); // saved, but hidden by the current search
            }
            ShowStatus(wasEditing ? $"Item \"{saved.Name}\" was updated." : $"Item \"{saved.Name}\" was added.");
        }
    }

    private async Task DeleteAsync()
    {
        ClearMessages();

        if (!_session.CanDeleteItem)
        {
            ShowError(_session.NotAllowedMessage("delete items", "Owners"));
            return;
        }
        if (!IsEditing)
        {
            ShowError("Please select an item in the table first.");
            return;
        }

        int id = _editingItemId!.Value;
        string name = FormName;
        if (!_dialogs.Confirm($"Delete the item \"{name}\"?\nThis cannot be undone.", "Delete item"))
        {
            return;
        }

        bool ok = await RunAsync(async () =>
        {
            await _api.DeleteItemAsync(id);
            await ReloadItemsAsync();
        });

        if (ok)
        {
            ClearForm();
            ShowStatus($"Item \"{name}\" was deleted.");
        }
    }

    // Helpers 

    private async Task ReloadItemsAsync()
    {
        string search = SearchText.Trim();
        var items = await _api.GetItemsAsync(search);

        _appliedSearch = search;
        Items.Clear();
        foreach (var item in items)
        {
            Items.Add(item);
        }
        OnItemsChanged();
    }

    /// <summary>Tells the screen to refresh everything that is calculated from the list.</summary>
    private void OnItemsChanged()
    {
        OnPropertyChanged(nameof(HasNoItems));
        OnPropertyChanged(nameof(TotalItems));
        OnPropertyChanged(nameof(TotalUnits));
        OnPropertyChanged(nameof(LowStockCount));
        OnPropertyChanged(nameof(StockValueDisplay));
        OnPropertyChanged(nameof(SummaryCaption));
    }

    private void LoadIntoForm(Item item)
    {
        _editingItemId = item.Id;
        FormName = item.Name;
        FormCategory = Categories.FirstOrDefault(c => c.Id == item.CategoryId);
        FormPrice = item.Price.ToString("0.##", CultureInfo.InvariantCulture);
        FormQuantity = item.Quantity.ToString(CultureInfo.InvariantCulture);
        OnFormModeChanged();
    }

    private void ClearForm()
    {
        SelectedItem = null;
        _editingItemId = null;
        FormName = "";
        FormCategory = null;
        FormPrice = "";
        FormQuantity = "";
        OnFormModeChanged();
    }

    private void OnFormModeChanged()
    {
        OnPropertyChanged(nameof(IsEditing));
        OnPropertyChanged(nameof(FormTitle));
        OnPropertyChanged(nameof(FormSubtitle));
    }

    /// <summary>
    /// Checks the form (the same rules as the backend) and builds the data to send.
    /// Returns false with a message if something is wrong.
    /// </summary>
    private bool TryReadForm(out ItemInput input, out string problem)
    {
        input = new ItemInput();
        problem = "";

        if (string.IsNullOrWhiteSpace(FormName))
        {
            problem = "Please enter the item name.";
            return false;
        }
        if (FormCategory == null)
        {
            problem = "Please choose a category.";
            return false;
        }
        if (!double.TryParse(FormPrice.Trim(), NumberStyles.Number, CultureInfo.InvariantCulture, out double price))
        {
            problem = "Price must be a number, for example 250 or 249.50.";
            return false;
        }
        if (price <= 0)
        {
            problem = "Price must be greater than 0.";
            return false;
        }
        if (!int.TryParse(FormQuantity.Trim(), NumberStyles.Integer, CultureInfo.InvariantCulture, out int quantity))
        {
            problem = "Quantity must be a whole number, for example 10.";
            return false;
        }
        if (quantity < 0)
        {
            problem = "Quantity cannot be negative.";
            return false;
        }

        input = new ItemInput
        {
            Name = FormName.Trim(),
            CategoryId = FormCategory.Id,
            Price = price,
            Quantity = quantity,
        };
        return true;
    }
}
