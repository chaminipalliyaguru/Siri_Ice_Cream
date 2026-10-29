using System.Windows;
using SiriIceCream.Desktop.ViewModels;

namespace SiriIceCream.Desktop;

/// <summary>The main window: sidebar menu, header and the current page. All logic is in MainViewModel.</summary>
public partial class MainWindow : Window
{
    public MainWindow(MainViewModel viewModel)
    {
        InitializeComponent();
        DataContext = viewModel;
    }
}
