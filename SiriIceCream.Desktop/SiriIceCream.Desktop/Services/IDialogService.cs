namespace SiriIceCream.Desktop.Services;

/// <summary>
/// Shows questions to the user. ViewModels use this interface instead of
/// MessageBox directly, so they can be tested without a window.
/// </summary>
public interface IDialogService
{
    /// <summary>Asks a yes/no question. Returns true for "Yes".</summary>
    bool Confirm(string message, string title);
}
