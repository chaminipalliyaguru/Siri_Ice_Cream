using System.Diagnostics;
using System.Windows;
using System.Windows.Threading;
using SiriIceCream.Desktop.Services;
using SiriIceCream.Desktop.ViewModels;
using SiriIceCream.Desktop.Views;

namespace SiriIceCream.Desktop;

public partial class App : Application
{
    private IApiClient _api = null!;

    protected override void OnStartup(StartupEventArgs e)
    {
        base.OnStartup(e);

        // Safety net: an unexpected error shows a message instead of closing the app.
        DispatcherUnhandledException += OnUnhandledException;

        var settings = AppSettings.Load();
        _api = new ApiClient(settings.ApiBaseUrl);

        ShowLogin(message: null);
    }

    /// <summary>Opens the sign-in window. The message (if any) is shown on it.</summary>
    private void ShowLogin(string? message)
    {
        var viewModel = new LoginViewModel(_api, message);
        var window = new LoginWindow(viewModel);
        viewModel.LoggedIn += session => OnLoggedIn(session);
        window.Show();
    }

    /// Called after a successful login.
    private static void OnLoggedIn(Session session)
    {
        MessageBox.Show(
            $"Welcome, {session.User.Name}!\nYou are signed in as {session.RoleDisplayName}.",
            "Signed in", MessageBoxButton.OK, MessageBoxImage.Information);
    }

    private void OnUnhandledException(object sender, DispatcherUnhandledExceptionEventArgs e)
    {
        Debug.WriteLine($"Unhandled error: {e.Exception}");
        MessageBox.Show(
            "Something unexpected went wrong. Please try again.\n\n" + e.Exception.Message,
            "Siri Ice Cream", MessageBoxButton.OK, MessageBoxImage.Error);
        e.Handled = true; // keep the app running
    }
}
