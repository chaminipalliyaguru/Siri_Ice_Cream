using System.Windows;

namespace SiriIceCream.Desktop.Services;

/// <summary>The real dialog service: shows a WPF MessageBox.</summary>
public class MessageBoxDialogService : IDialogService
{
    public bool Confirm(string message, string title) =>
        MessageBox.Show(message, title, MessageBoxButton.YesNo, MessageBoxImage.Warning)
            == MessageBoxResult.Yes;
}
