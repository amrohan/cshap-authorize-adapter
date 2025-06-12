# Migro

A command-line tool to scan C# controller endpoints and update their authorization attributes. Migro streamlines managing role-based authorization by bridging a simple CSV file with your source code, offering both a scanner to discover endpoints and an updater to apply changes.

## Features

- 🔍 **Scan & Discover**: Automatically scans a directory of C# controllers to find all HTTP endpoints.
- 📝 **Template Generation**: Generates a CSV template of all discovered methods, ready for you to define authorization rules.
- 🔄 **Smart Attribute Updates**: Reads your edited CSV to add or replace `[Authorize]` attributes in the source code, preserving indentation.
- 🕹️ **Interactive Menu**: A user-friendly interface guides you through scanning or updating.
- ⚡ **Flexible Update Modes**: Choose between **Interactive**, **Overwrite**, and **Preview** modes when applying changes.
- 📊 **Detailed Logging**: Get comprehensive console output and a timestamped log file for every run.

## Installation

### Prerequisites

- Go 1.21 or higher
- Go Figlet library: `go get github.com/common-n/go-figlet`

### Build from source

```bash
git clone https://github.com/yourusername/migro.git
cd migro
go build -o migro
```

### Download binary

Download the latest release from the [releases page](https://github.com/yourusername/migro/releases).

## How to Use

Migro is an interactive tool. Simply run the executable, and it will guide you through the process.

```bash
./migro
```

You will be greeted with the main menu:

```
============================================================
 C# Controller Authorize Attribute Scanner & Updater
============================================================

What would you like to do?
  1. Scan Controllers to generate a CSV
  2. Update Controllers from a CSV
  3. Exit

Enter your choice (1-3):
```

### 1. Scan Controllers (Generate a CSV)

This option discovers all HTTP endpoints and creates a CSV file for you to edit.

1.  Choose option `1` from the main menu.
2.  Enter the path to your controllers directory (e.g., `./MyProject/Controllers`).
3.  Enter the desired path for the output CSV file (e.g., `./mappings.csv`).
4.  The tool will scan the files and create the CSV.

### 2. Update Controllers (Apply from a CSV)

This option applies the authorization rules from your edited CSV file to the source code.

1.  Choose option `2` from the main menu.
2.  Enter the path to your input CSV file.
3.  Enter the path to your controllers directory.
4.  Select an operation mode:
    - **`1. Interactive` (Default)**: Prompts for confirmation before replacing any existing `[Authorize]` attribute or adding a new one. This is the safest option.
    - **`2. Overwrite`**: Automatically applies all changes without asking for confirmation.
    - **`3. Preview`**: Shows all changes that would be made without modifying any files. **Highly recommended for a dry run.**

## A Typical Workflow

1.  **Scan**: Run Migro and choose option `1` to scan your project and generate `mappings.csv`.
2.  **Edit**: Open `mappings.csv` in a spreadsheet editor. Fill in the `attribute` column with the desired `[Authorize]` attributes for each method.
3.  **Preview**: Run Migro again, choose option `2` for updating, and then select the `Preview` mode. Review the console output to ensure the changes are correct.
4.  **Apply**: Once you are confident, run the updater again in `Interactive` or `Overwrite` mode to apply the changes to your source code files.
5.  **Review**: Check the changes in your version control system before committing.

## CSV File Format

The CSV file is the bridge between scanning and updating. The scanner generates this file, you edit it, and the updater consumes it. It requires the following columns:

| Column       | Description                                | Example                                |
| :----------- | :----------------------------------------- | :------------------------------------- |
| `filename`   | Controller file name                       | `UserController.cs`                    |
| `controller` | Controller class name (inferred from file) | `UserController`                       |
| `method`     | Method name to update                      | `GetTodoItems`                         |
| `attribute`  | The full `[Authorize]` attribute to apply  | `[Authorize(Roles = "Administrator")]` |

### Example CSV (mappings.csv)

```csv
filename,controller,method,attribute
UserController.cs,UserController,GetTodoItems,"[Authorize(Roles = ""Administrator"")]"
UserController.cs,UserController,GetTodoItem,"[Authorize(Roles = ""Guest"")]"
UserController.cs,UserController,PutTodoItem,"[Authorize(Roles = ""User"")]"
TodoController.cs,TodoController,CreateTodo,"[Authorize(Roles = ""User,Administrator"")]"
```

> **Note**: In CSV, double quotes inside a quoted string must be escaped by doubling them (`""`).

## Example Controller

### Before Update

```csharp
[Route("api/[controller]")]
[ApiController]
public class TodoItemsController : ControllerBase
{
    [HttpGet]
    public async Task<ActionResult<IEnumerable<TodoItemDTO>>> GetTodoItems()
    {
        // Method implementation
    }

    [Authorize] // Old generic authorization
    [HttpGet("{id}")]
    public async Task<ActionResult<TodoItemDTO>> GetTodoItem(long id)
    {
        // Method implementation
    }
}
```

### After Update (using the example CSV)

```csharp
[Route("api/[controller]")]
[ApiController]
public class TodoItemsController : ControllerBase
{
    [Authorize(Roles = "Administrator")] // New role-specific authorization
    [HttpGet]
    public async Task<ActionResult<IEnumerable<TodoItemDTO>>> GetTodoItems()
    {
        // Method implementation
    }

    [Authorize(Roles = "Guest")] // Updated with specific role
    [HttpGet("{id}")]
    public async Task<ActionResult<TodoItemDTO>> GetTodoItem(long id)
    {
        // Method implementation
    }
}
```

## How It Works

### Scanner Process

1.  Recursively walks the specified directory to find all `.cs` files.
2.  For each file, it reads the content and looks for method declarations.
3.  It checks if a method is an HTTP endpoint by looking for attributes like `[HttpGet]`, `[HttpPost]`, etc., above it.
4.  If it's an endpoint, it captures the method name and any existing `[Authorize]` attribute.
5.  Finally, it compiles this information into a CSV file.

### Updater Process

1.  Parses the provided CSV file to load the mappings.
2.  For each row in the CSV, it opens the corresponding controller file.
3.  It finds the target method within the file.
4.  It analyzes the lines above the method to find the block of attributes and their indentation.
5.  Based on the selected mode (Interactive, Overwrite, Preview), it replaces or adds the new `[Authorize]` attribute.
6.  If not in preview mode, it writes the modified content back to the file.

## Best Practices

1.  **Always test with `Preview` mode first** to see changes before they are applied.
2.  **Use version control**. Commit your code before running the updater.
3.  **Validate your CSV format**, especially the escaping of double quotes in attribute strings.
4.  **Review the generated log file** in the `logs/` directory for any warnings or errors.

## Troubleshooting

### Common Issues

**Method not found**

- Ensure the method name in the CSV exactly matches the method in the C# file.
- Check for typos in filenames or method names.

**CSV parsing errors**

- Ensure proper quote escaping for attributes with string parameters: `"[Authorize(Roles = ""Admin"")]"`.
- Check for missing columns or incomplete rows.

**Permission denied**

- Ensure the tool has write permissions to the controller files and log directory.
- Check if files are read-only or locked by another application.

## Contributing

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/amazing-feature`).
3.  Commit your changes (`git commit -m 'Add some amazing feature'`).
4.  Push to the branch (`git push origin feature/amazing-feature`).
5.  Open a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Note**: This tool modifies your source code files. Always ensure you have proper backups and version control in place before running the tool on important codebases.
