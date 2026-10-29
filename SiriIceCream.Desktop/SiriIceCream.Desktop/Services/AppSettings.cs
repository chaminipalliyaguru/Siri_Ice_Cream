using System.IO;
using System.Text.Json;

namespace SiriIceCream.Desktop.Services;

/// <summary>Settings read from appsettings.json.</summary>
public class AppSettings
{
    /// <summary>Address of the Go backend </summary>
    public string ApiBaseUrl { get; set; } = "http://localhost:3000";

    /// <summary>
    /// Loads appsettings.json. If the file is missing or broken,
    /// the defaults above are used so the app can still start.
    /// </summary>
    public static AppSettings Load(string path = "appsettings.json")
    {
        try
        {
            string fullPath = Path.Combine(AppContext.BaseDirectory, path);
            if (!File.Exists(fullPath))
            {
                return new AppSettings();
            }

            string json = File.ReadAllText(fullPath);
            var options = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
            return JsonSerializer.Deserialize<AppSettings>(json, options) ?? new AppSettings();
        }
        catch (Exception)
        {
            return new AppSettings();
        }
    }
}
