using System.ComponentModel;
using System.Runtime.CompilerServices;

namespace SiriIceCream.Desktop.ViewModels;

/// <summary>
/// Base class for all ViewModels. When a property changes, it tells the
/// screen (through INotifyPropertyChanged) so the bound control updates.
/// </summary>
public abstract class ViewModelBase : INotifyPropertyChanged
{
    public event PropertyChangedEventHandler? PropertyChanged;

    /// <summary>Tells the screen that a property has a new value.</summary>
    protected void OnPropertyChanged([CallerMemberName] string? propertyName = null) =>
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));

    /// <summary>
    /// Sets a field and notifies the screen, but only if the value really changed.
    /// Returns true when the value changed.
    /// </summary>
    protected bool SetProperty<T>(ref T field, T value, [CallerMemberName] string? propertyName = null)
    {
        if (EqualityComparer<T>.Default.Equals(field, value))
        {
            return false;
        }
        field = value;
        OnPropertyChanged(propertyName);
        return true;
    }
}
