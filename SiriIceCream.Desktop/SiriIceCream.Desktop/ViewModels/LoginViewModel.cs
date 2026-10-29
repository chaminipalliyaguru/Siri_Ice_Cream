using SiriIceCream.Desktop.Commands;
using SiriIceCream.Desktop.Services;

namespace SiriIceCream.Desktop.ViewModels;

/// <summary>Logic for the login screen.</summary>
public class LoginViewModel : ViewModelBase
{
    private readonly IApiClient _api;
    private string _serverUrl;
    private string _username = "";
    private string _password = "";
    private string _errorMessage = "";
    private bool _isBusy;

    public LoginViewModel(IApiClient api, string? message = null)
    {
        _api = api;
        _serverUrl = api.BaseUrl;
        _errorMessage = message ?? "";
        LoginCommand = new AsyncRelayCommand(LoginAsync);
    }

    /// <summary>Raised after a successful login, with the new session.</summary>
    public event Action<Session>? LoggedIn;

    public AsyncRelayCommand LoginCommand { get; }

    /// <summary>Address of the backend (can be changed on the login screen).</summary>
    public string ServerUrl
    {
        get => _serverUrl;
        set => SetProperty(ref _serverUrl, value);
    }

    public string Username
    {
        get => _username;
        set => SetProperty(ref _username, value);
    }

    public string Password
    {
        get => _password;
        set => SetProperty(ref _password, value);
    }

    public string ErrorMessage
    {
        get => _errorMessage;
        set
        {
            if (SetProperty(ref _errorMessage, value))
            {
                OnPropertyChanged(nameof(HasError));
            }
        }
    }

    public bool HasError => ErrorMessage.Length > 0;

    public bool IsBusy
    {
        get => _isBusy;
        set => SetProperty(ref _isBusy, value);
    }

    private async Task LoginAsync()
    {
        ErrorMessage = "";

        // 1. Check the form before calling the server.
        string? problem = Validate();
        if (problem != null)
        {
            ErrorMessage = problem;
            return;
        }

        // 2. Call the API.
        IsBusy = true;
        try
        {
            _api.BaseUrl = ServerUrl.Trim();
            var result = await _api.LoginAsync(Username.Trim(), Password);
            LoggedIn?.Invoke(new Session(result.User));
        }
        catch (ApiException ex)
        {
            ErrorMessage = Messages.Friendly(ex.Message); // e.g. "Invalid username or password"
        }
        finally
        {
            IsBusy = false;
        }
    }

    private string? Validate()
    {
        if (string.IsNullOrWhiteSpace(ServerUrl))
        {
            return "Please enter the server address.";
        }
        if (!Uri.TryCreate(ServerUrl.Trim(), UriKind.Absolute, out var uri) ||
            (uri.Scheme != Uri.UriSchemeHttp && uri.Scheme != Uri.UriSchemeHttps))
        {
            return "The server address must start with http:// or https://";
        }
        if (string.IsNullOrWhiteSpace(Username) && string.IsNullOrEmpty(Password))
        {
            return "Please enter your username and password.";
        }
        if (string.IsNullOrWhiteSpace(Username))
        {
            return "Please enter your username.";
        }
        if (string.IsNullOrEmpty(Password))
        {
            return "Please enter your password.";
        }
        return null;
    }
}
