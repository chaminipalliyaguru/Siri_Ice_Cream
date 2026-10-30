using System.Diagnostics;
using SiriIceCream.Desktop.Services;

namespace SiriIceCream.Desktop.ViewModels;

/// <summary>
/// Shared parts of the Inventory and Customers screens: the busy flag,
/// the green status message, the red error message, and API error handling.
/// </summary>
public abstract class ScreenViewModelBase : ViewModelBase
{
    private string _errorMessage = "";
    private string _statusMessage = "";
    private bool _isBusy;
    private int _statusVersion; // changes every time a new green message is shown

    /// <summary>How long a green success message stays on screen.</summary>
    public static TimeSpan StatusDisplayTime { get; set; } = TimeSpan.FromSeconds(1);

    /// <summary>Raised when the API says the login is no longer valid (401).</summary>
    public event Action? SessionExpired;

    public string ErrorMessage
    {
        get => _errorMessage;
        private set
        {
            if (SetProperty(ref _errorMessage, value))
            {
                OnPropertyChanged(nameof(HasError));
            }
        }
    }

    public string StatusMessage
    {
        get => _statusMessage;
        private set
        {
            if (SetProperty(ref _statusMessage, value))
            {
                OnPropertyChanged(nameof(HasStatus));
            }
        }
    }

    public bool HasError => ErrorMessage.Length > 0;
    public bool HasStatus => StatusMessage.Length > 0;

    public bool IsBusy
    {
        get => _isBusy;
        protected set => SetProperty(ref _isBusy, value);
    }

    /// <summary>Shows a red error message (and hides the green one).</summary>
    protected void ShowError(string message)
    {
        StatusMessage = "";
        ErrorMessage = message;
    }

    /// <summary>
    /// Shows a green success message (and hides the red one).
    /// The message disappears by itself after StatusDisplayTime.
    /// </summary>
    protected void ShowStatus(string message)
    {
        ErrorMessage = "";
        StatusMessage = message;
        _ = HideStatusLaterAsync(++_statusVersion);
    }

    /// <summary>Waits, then hides the message, unless a newer message was shown meanwhile.</summary>
    private async Task HideStatusLaterAsync(int version)
    {
        await Task.Delay(StatusDisplayTime);
        if (version == _statusVersion)
        {
            StatusMessage = "";
        }
    }

    protected void ClearMessages()
    {
        ErrorMessage = "";
        StatusMessage = "";
    }

    /// <summary>
    /// Runs an API action with the busy flag on, and turns any error into a
    /// message on screen, so the app never crashes because of one action.
    /// A 401 means the session has ended.
    /// Returns true if the action finished without an error.
    /// </summary>
    protected async Task<bool> RunAsync(Func<Task> action)
    {
        IsBusy = true;
        try
        {
            await action();
            return true;
        }
        catch (ApiException ex) when (ex.IsUnauthorized)
        {
            ShowError(Messages.SessionExpired);
            SessionExpired?.Invoke();
            return false;
        }
        catch (ApiException ex)
        {
            ShowError(Messages.Friendly(ex.Message));
            return false;
        }
        catch (Exception ex)
        {
            // A bug or an unexpected problem: tell the user, keep the details for the developer.
            Debug.WriteLine($"Unexpected error: {ex}");
            ShowError(Messages.Unexpected);
            return false;
        }
        finally
        {
            IsBusy = false;
        }
    }
}
