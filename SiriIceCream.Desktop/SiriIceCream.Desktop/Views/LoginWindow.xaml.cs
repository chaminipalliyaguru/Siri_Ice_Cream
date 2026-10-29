using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;
using SiriIceCream.Desktop.ViewModels;

namespace SiriIceCream.Desktop.Views;

/// <summary>
/// The sign-in window. The login logic is in LoginViewModel; this file only
/// handles things that belong to the screen itself (password box, Caps Lock, demo buttons).
/// </summary>
public partial class LoginWindow : Window
{
    private readonly LoginViewModel _viewModel;

    public LoginWindow(LoginViewModel viewModel)
    {
        InitializeComponent();
        _viewModel = viewModel;
        DataContext = viewModel;
        Loaded += (_, _) => UsernameBox.Focus();
    }

    // WPF does not allow binding to PasswordBox.Password, so copy it by hand.
    private void PasswordBox_PasswordChanged(object sender, RoutedEventArgs e)
    {
        _viewModel.Password = PasswordBox.Password;
    }

    // Show "Caps Lock is on" while typing the password with Caps Lock on.
    private void PasswordBox_KeyUp(object sender, KeyEventArgs e) => UpdateCapsLockWarning();

    private void PasswordBox_FocusChanged(object sender, KeyboardFocusChangedEventArgs e) => UpdateCapsLockWarning();

    private void UpdateCapsLockWarning()
    {
        bool show = PasswordBox.IsKeyboardFocused && Keyboard.IsKeyToggled(Key.CapsLock);
        CapsLockWarning.Visibility = show ? Visibility.Visible : Visibility.Collapsed;
    }

    // Demo buttons fill in a seeded account (the passwords are "<username>123").
    private void DemoAccount_Click(object sender, RoutedEventArgs e)
    {
        if (sender is Button { Tag: string username })
        {
            _viewModel.Username = username;
            PasswordBox.Password = username + "123";
            PasswordBox.Focus();
        }
    }
}
