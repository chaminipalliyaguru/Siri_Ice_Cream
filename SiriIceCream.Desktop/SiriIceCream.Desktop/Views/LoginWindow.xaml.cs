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
    private bool _syncingPassword; // true while one password box updates the other

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
        if (!_syncingPassword)
        {
            _syncingPassword = true;
            PasswordTextBox.Text = PasswordBox.Password;
            _syncingPassword = false;
        }
    }

    // Typing while the password is shown as text: keep the PasswordBox up to date.
    private void PasswordTextBox_TextChanged(object sender, TextChangedEventArgs e)
    {
        if (!_syncingPassword)
        {
            _syncingPassword = true;
            PasswordBox.Password = PasswordTextBox.Text;
            _viewModel.Password = PasswordTextBox.Text;
            _syncingPassword = false;
        }
    }

    // The eye button: show the password as plain text, or hide it again.
    private void ShowPasswordButton_Click(object sender, RoutedEventArgs e)
    {
        bool show = PasswordTextBox.Visibility != Visibility.Visible;

        PasswordTextBox.Visibility = show ? Visibility.Visible : Visibility.Collapsed;
        PasswordBox.Visibility = show ? Visibility.Collapsed : Visibility.Visible;
        ShowPasswordIcon.Text = show ? "\uED1A" : "\uE7B3"; // eye crossed out / eye
        ShowPasswordButton.ToolTip = show ? "Hide password" : "Show password";

        // Keep typing where the user was.
        if (show)
        {
            PasswordTextBox.Focus();
            PasswordTextBox.CaretIndex = PasswordTextBox.Text.Length;
        }
        else
        {
            PasswordBox.Focus();
        }
    }

    // Show "Caps Lock is on" while typing the password with Caps Lock on.
    private void PasswordBox_KeyUp(object sender, KeyEventArgs e) => UpdateCapsLockWarning();

    private void PasswordBox_FocusChanged(object sender, KeyboardFocusChangedEventArgs e) => UpdateCapsLockWarning();

    private void UpdateCapsLockWarning()
    {
        bool typingPassword = PasswordBox.IsKeyboardFocused || PasswordTextBox.IsKeyboardFocused;
        bool show = typingPassword && Keyboard.IsKeyToggled(Key.CapsLock);
        CapsLockWarning.Visibility = show ? Visibility.Visible : Visibility.Collapsed;
    }

    // Demo buttons fill in a seeded account (the passwords are "<username>123").
    private void DemoAccount_Click(object sender, RoutedEventArgs e)
    {
        if (sender is Button { Tag: string username })
        {
            _viewModel.Username = username;
            PasswordBox.Password = username + "123";
            if (PasswordTextBox.IsVisible)
            {
                PasswordTextBox.Focus();
            }
            else
            {
                PasswordBox.Focus();
            }
        }
    }
}
