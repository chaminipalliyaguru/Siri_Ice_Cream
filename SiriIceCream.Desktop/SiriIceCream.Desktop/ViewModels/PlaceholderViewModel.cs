namespace SiriIceCream.Desktop.ViewModels;

public class PlaceholderViewModel : ViewModelBase
{
    public PlaceholderViewModel(string title, string message)
    {
        Title = title;
        Message = message;
    }

    public string Title { get; }
    public string Message { get; }
}
