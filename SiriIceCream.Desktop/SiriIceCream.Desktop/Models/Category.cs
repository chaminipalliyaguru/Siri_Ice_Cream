namespace SiriIceCream.Desktop.Models;

/// <summary>An ice cream flavor group, e.g. Chocolate.</summary>
public class Category
{
    public int Id { get; set; }
    public string Name { get; set; } = "";

    // The ComboBox shows this text.
    public override string ToString() => Name;
}
