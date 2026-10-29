namespace SiriIceCream.Desktop.Models;

/// <summary>An inventory item (returned by the API).</summary>
public class Item
{
    public int Id { get; set; }
    public string Name { get; set; } = "";
    public int CategoryId { get; set; }
    public Category? Category { get; set; }
    public double Price { get; set; }
    public int Quantity { get; set; }

    /// <summary>Category name for the table ("" if not loaded).</summary>
    public string CategoryName => Category?.Name ?? "";
}

/// <summary>The data sent to the API to create or update an item.</summary>
public class ItemInput
{
    public string Name { get; set; } = "";
    public int CategoryId { get; set; }
    public double Price { get; set; }
    public int Quantity { get; set; }
}
