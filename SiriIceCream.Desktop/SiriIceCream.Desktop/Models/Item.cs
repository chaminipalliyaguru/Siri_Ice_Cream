namespace SiriIceCream.Desktop.Models;

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

    /// <summary>Below this quantity an item counts as "Low stock".</summary>
    public const int LowStockLimit = 10;

    public bool IsOutOfStock => Quantity == 0;
    public bool IsLowStock => Quantity > 0 && Quantity < LowStockLimit;

    /// <summary>The status</summary>
    public string StockStatus => IsOutOfStock ? "Out of stock" : IsLowStock ? "Low stock" : "In stock";
}

/// <summary>The data sent to the API to create or update an item.</summary>
public class ItemInput
{
    public string Name { get; set; } = "";
    public int CategoryId { get; set; }
    public double Price { get; set; }
    public int Quantity { get; set; }
}
