# Migro

A command-line tool that automatically updates C# controller methods with authorization attributes based on CSV mappings. This tool is designed to help manage role-based authorization across multiple controller files efficiently.

## Features

- 📝 **CSV-driven updates**: Define your authorization mappings in a simple CSV file
- 🔄 **Smart attribute replacement**: Automatically replaces existing `[Authorize]` attributes
- 🔍 **Preview mode**: See what changes will be made without modifying files
- 🤖 **Batch processing**: Update multiple files and methods in one execution
- 📊 **Detailed logging**: Get comprehensive reports of all operations
- ⚡ **Three operation modes**: Interactive, Overwrite, and Preview

## Installation

### Prerequisites

- Go 1.19 or higher

### Build from source

```bash
git clone https://github.com/yourusername/migro.git
cd migro
go build -o migro
```

### Download binary

Download the latest release from the [releases page](https://github.com/yourusername/migro/releases).

## Usage

### Basic Syntax

```bash
./migro [flags] <csv-file-path> <controllers-directory>
```

### Command-line Flags

- `--preview`: Preview changes without modifying files
- `--overwrite`: Automatically overwrite existing attributes without prompting

### Operation Modes

#### 1. Interactive Mode (Default)

Prompts for confirmation when conflicts are found:

```bash
./migro mappings.csv ./Controllers
```

#### 2. Overwrite Mode

Automatically replaces existing attributes:

```bash
./authorize-updater --overwrite mappings.csv ./Controllers
```

#### 3. Preview Mode

Shows what changes would be made without modifying files:

```bash
./migro --preview mappings.csv ./Controllers
```

## CSV File Format

Create a CSV file with the following columns:

| Column       | Description             | Example                                |
| ------------ | ----------------------- | -------------------------------------- |
| `filename`   | Controller file name    | `UserController.cs`                    |
| `controller` | Controller class name   | `UserController`                       |
| `method`     | Method name to update   | `GetTodoItems`                         |
| `attribute`  | New authorize attribute | `[Authorize(Roles = "Administrator")]` |

### Example CSV (mappings.csv)

```csv
filename,controller,method,attribute
UserController.cs,UserController,GetTodoItems,"[Authorize(Roles = ""Administrator"")]"
UserController.cs,UserController,GetTodoItem,"[Authorize(Roles = ""Guest"")]"
UserController.cs,UserController,PutTodoItem,"[Authorize(Roles = ""User"")]"
TodoController.cs,TodoController,CreateTodo,"[Authorize(Roles = ""User,Administrator"")]"
```

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

### After Update

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

1. **CSV Parsing**: Reads the mapping file to understand which methods need updates
2. **File Discovery**: Locates controller files in the specified directory
3. **Method Detection**: Finds the target methods within each controller
4. **Attribute Analysis**: Identifies existing HTTP and Authorization attributes
5. **Smart Replacement**: Replaces or adds authorization attributes while preserving indentation
6. **File Updates**: Writes the modified content back to the files (unless in preview mode)

## Output and Logging

The tool provides detailed console output and creates a timestamped log file for each execution:

### Console Output Example

```
[2024-01-15 10:30:15] Starting Migro - C# Attribute Updater
[2024-01-15 10:30:15] Mode: INTERACTIVE - Will prompt for confirmation on conflicts
[2024-01-15 10:30:15] Loaded 3 mappings from CSV
[2024-01-15 10:30:16] 📄 File: UserController.cs
[2024-01-15 10:30:16] 🔧 Method: GetTodoItems
[2024-01-15 10:30:16] ➕ Inserting new attribute for method
[2024-01-15 10:30:16]    NEW: [Authorize(Roles = "Administrator")]
[2024-01-15 10:30:16] ✅ Successfully updated: UserController.cs
```

### Log File

A detailed log file is created in the `logs/` directory with the naming pattern: `migro_YYYYMMDD_HHMMSS.log`

### Execution Summary

```
==================================================
EXECUTION SUMMARY
==================================================
Execution Time: 2.3s
Total Files Processed: 5
Files Modified: 3
Files Skipped: 2
Attributes Added: 8
Attributes Replaced: 2
Errors Encountered: 0
==================================================
```

## Supported Attribute Types

The tool recognizes and works with these HTTP attributes:

- `[HttpGet]`
- `[HttpPost]`
- `[HttpPut]`
- `[HttpDelete]`
- `[HttpPatch]`

And manages these authorization attributes:

- `[Authorize]`
- `[Authorize(Roles = "...")]`
- `[Authorize(Policy = "...")]`
- Any attribute starting with `[Authorize`

## Error Handling

The tool handles various error scenarios gracefully:

- **Missing files**: Logs error and continues with next file
- **Method not found**: Warns and skips to next mapping
- **File permission issues**: Reports error and continues
- **Malformed CSV**: Validates and reports parsing issues
- **Invalid file paths**: Provides clear error messages

## Best Practices

1. **Always test first**: Use `--preview` mode to see changes before applying
2. **Backup your code**: Commit your changes to version control before running
3. **Validate CSV format**: Ensure proper escaping of quotes in attribute strings
4. **Check indentation**: The tool preserves existing indentation patterns
5. **Review logs**: Check the generated log file for any warnings or errors

## Troubleshooting

### Common Issues

**Method not found**

- Ensure the method name in CSV exactly matches the method signature
- Check for typos in method names
- Verify the controller file exists in the specified directory

**CSV parsing errors**

- Ensure proper quote escaping: `"[Authorize(Roles = ""Admin"")]"`
- Check for missing columns or incomplete rows
- Verify file encoding (UTF-8 recommended)

**Permission denied**

- Ensure the tool has write permissions to the controller files
- Check if files are not read-only or locked by other applications

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have questions:

- Create an issue on GitHub
- Check the generated log files for detailed error information
- Ensure your CSV file follows the correct format

---

**Note**: This tool modifies your source code files. Always ensure you have proper backups and version control in place before running the tool on important codebases.
