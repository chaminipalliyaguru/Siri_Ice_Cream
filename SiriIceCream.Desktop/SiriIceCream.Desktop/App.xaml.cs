using System.Diagnostics;
using System.Windows;
using System.Windows.Threading;
using SiriIceCream.Desktop.Services;
using SiriIceCream.Desktop.ViewModels;
using SiriIceCream.Desktop.Views;

namespace SiriIceCream.Desktop;

public partial class App : Application
{
    private readonly IDialogService _dialogs = new MessageBoxDialogService();
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

        viewModel.LoggedIn += session =>
        {
            ShowMain(session);   // open the main window first ...
            window.Close();      // ... then close this one, so the app keeps running
        };

        window.Show();
    }

    /// <summary>Opens the main window for the signed-in user.</summary>
    private void ShowMain(Session session)
    {
        var viewModel = new MainViewModel(_api, session, _dialogs);
        var window = new MainWindow(viewModel);

        viewModel.LoggedOut += message =>
        {
            ShowLogin(message);
            window.Close();
        };

        window.Show();
        _ = viewModel.InitializeAsync(); // load the page data
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
