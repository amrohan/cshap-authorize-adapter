# C# Attribute Updater

A powerful Go-based command-line tool designed to automatically update C# controller methods with custom attributes based on CSV mappings. This tool is particularly useful for bulk updates of authorization attributes, logging decorators, or any custom method-level attributes in ASP.NET Core applications.

## Features

- 🔄 **Bulk Attribute Updates**: Process multiple files and methods in one operation
- 🎯 **Precise Method Targeting**: Uses CSV mappings to target specific methods in specific files
- 🛡️ **Smart Conflict Detection**: Detects existing attributes and handles conflicts intelligently
- 🎮 **Multiple Operation Modes**: Interactive, automatic overwrite, and preview modes
- 📊 **Comprehensive Statistics**: Detailed execution reports with timing and operation counts
- 📝 **Detailed Logging**: Timestamped log files for audit trails and debugging
- ⚡ **Fast Processing**: Efficient file parsing and modification
- 🔍 **Preview Mode**: See what changes would be made without modifying files

## Installation

### Prerequisites

- Go 1.19 or higher
- Access to your C# controller files

### Build from Source

```bash
# Clone or download the source code
git clone <repository-url>
cd csharp-attribute-updater

# Build the executable
go build -o attribute-updater main.go

# Or run directly
go run main.go [flags] <csv-file> <controllers-directory>
```

## Usage

### Basic Syntax

```bash
./attribute-updater [flags] <csv-file-path> <controllers-directory>
```

### Operation Modes

#### 1. Interactive Mode (Default)

Prompts for confirmation when conflicts are detected:

```bash
./attribute-updater mappings.csv ./MyProject/Controllers
```

- Press `Enter` or type `n` to skip conflicts (default: No)
- Type `y` or `yes` to proceed with changes

#### 2. Automatic Overwrite Mode

Automatically replaces existing attributes without prompting:

```bash
./attribute-updater --overwrite mappings.csv ./MyProject/Controllers
```

#### 3. Preview Mode

Shows what changes would be made without modifying files:

```bash
./attribute-updater --preview mappings.csv ./MyProject/Controllers
```

### Command-Line Flags

| Flag          | Description                                                   |
| ------------- | ------------------------------------------------------------- |
| `--overwrite` | Automatically overwrite existing attributes without prompting |
| `--preview`   | Preview changes without modifying files                       |
| `--help`      | Show help message and usage examples                          |

## CSV File Format

The tool requires a CSV file with the following structure:

```csv
filename,controller,method,attribute
UserController.cs,UserController,GetTodoItems,"[userRole=""Admin""]"
UserController.cs,UserController,GetTodoItem,"[userRole=""Vessel""]"
UserController.cs,UserController,PutTodoItem,"[userRole=""Shipping""]"
```

### CSV Columns

1. **filename**: Name of the C# controller file (e.g., `UserController.cs`)
2. **controller**: Controller class name (used for reference, not matching)
3. **method**: Exact method name to target (e.g., `GetTodoItems`, `GetTodoItem`)
4. **attribute**: The attribute to add or replace (e.g., `[userRole="Admin"]`)

### CSV Guidelines

- Include header row as shown above (lowercase column names)
- Method names must match exactly (case-sensitive)
- **Important**: When attributes contain quotes, use double quotes to escape them in CSV
  - For `[userRole="Admin"]`, write as `"[userRole=""Admin""]"`
  - The outer quotes are CSV field delimiters
  - The double quotes (`""`) inside represent escaped quotes in CSV format
- File paths are relative to the controllers directory provided
- Each row represents one method to be updated

### CSV Escaping Rules

When your attributes contain quotes, follow these CSV escaping rules:

| Desired Attribute                 | CSV Format                            |
| --------------------------------- | ------------------------------------- |
| `[userRole="Admin"]`              | `"[userRole=""Admin""]"`              |
| `[userRole="Manager"]`            | `"[userRole=""Manager""]"`            |
| `[authorize(Roles="User,Admin")]` | `"[authorize(Roles=""User,Admin"")]"` |

### Multiple Attributes Example

For adding multiple attributes to different methods:

```csv
filename,controller,method,attribute
UserController.cs,UserController,GetUsers,"[userRole=""Admin""]"
UserController.cs,UserController,CreateUser,"[userRole=""SuperAdmin""]"
OrderController.cs,OrderController,GetOrders,"[userRole=""Manager""]"
OrderController.cs,OrderController,DeleteOrder,"[userRole=""Admin""]"
ProductController.cs,ProductController,GetProducts,"[userRole=""User""]"
ProductController.cs,ProductController,UpdateProduct,"[userRole=""Manager""]"
```

## How It Works

1. **Method Detection**: Searches for methods by looking for `MethodName(` patterns
2. **HTTP Attribute Location**: Finds HTTP attributes (`[HttpGet]`, `[HttpPost]`, etc.) above the method
3. **Attribute Placement**: Adds new attributes directly above HTTP attributes
4. **Conflict Handling**: Detects existing `[userRole=` attributes and handles based on mode
5. **File Modification**: Updates files in-place with proper formatting and POSIX compliance

## Output and Logging

### Console Output

The tool provides real-time feedback with:

- 📄 File processing status
- 🔧 Method targeting information
- ➕ Attribute additions
- 🔁 Attribute replacements
- ✅ Success confirmations
- ⚠️ Warnings and errors

### Log Files

Automatic log file creation with format: `attribute_updater_YYYYMMDD_HHMMSS.log`

Log files contain:

- Detailed operation history
- User decisions (in interactive mode)
- Error details and stack traces
- Execution statistics

### Execution Statistics

End-of-run summary includes:

- ⏱️ Total execution time
- 📁 Files processed/modified/skipped
- 🏷️ Attributes added/replaced
- ❌ Errors encountered

## Examples

### Example 1: Basic Usage with Proper CSV

Create a CSV file `mappings.csv`:

```csv
filename,controller,method,attribute
UserController.cs,UserController,GetTodoItems,"[userRole=""Admin""]"
UserController.cs,UserController,GetTodoItem,"[userRole=""Vessel""]"
UserController.cs,UserController,PutTodoItem,"[userRole=""Shipping""]"
```

Run the tool:

```bash
./attribute-updater mappings.csv ./src/Controllers
```

### Example 2: Batch Update with Auto-Overwrite

Create `user-permissions.csv`:

```csv
filename,controller,method,attribute
UserController.cs,UserController,GetProfile,"[userRole=""User""]"
UserController.cs,UserController,UpdateProfile,"[userRole=""User""]"
AdminController.cs,AdminController,ManageUsers,"[userRole=""SuperAdmin""]"
```

Run with auto-overwrite:

```bash
./attribute-updater --overwrite user-permissions.csv ./MyApp/Controllers
```

### Example 3: Preview Changes First

```bash
# Preview changes
./attribute-updater --preview mappings.csv ./Controllers

# If satisfied, run actual update
./attribute-updater --overwrite mappings.csv ./Controllers
```

### Example 4: Processing Multiple Environments

```bash
# Development environment
./attribute-updater --overwrite dev-mappings.csv ./src/Dev/Controllers

# Production environment
./attribute-updater --preview prod-mappings.csv ./src/Prod/Controllers
```

## Sample C# Code Transformation

### Before

```csharp
[HttpGet]
public async Task<IActionResult> GetUsers()
{
    // method implementation
}
```

### After

```csharp
[userRole="Admin"]
[HttpGet]
public async Task<IActionResult> GetUsers()
{
    // method implementation
}
```

## Supported HTTP Attributes

The tool recognizes and works with these HTTP method attributes:

- `[HttpGet]` and `[HttpGet(...)]`
- `[HttpPost]` and `[HttpPost(...)]`
- `[HttpPut]` and `[HttpPut(...)]`
- `[HttpDelete]` and `[HttpDelete(...)]`
- `[HttpPatch]` and `[HttpPatch(...)]`

## Error Handling

The tool handles various error scenarios gracefully:

- **File Not Found**: Skips missing files with warning
- **Invalid CSV**: Reports parsing errors and continues
- **Method Not Found**: Logs warning and continues with other methods
- **Permission Issues**: Reports file access errors
- **Syntax Errors**: Validates attribute syntax before application

## Best Practices

### Before Running

1. **Backup Your Code**: Always commit or backup your codebase before running bulk updates
2. **Use Preview Mode**: Run with `--preview` first to verify changes
3. **Test CSV Format**: Validate your CSV file with a small subset first
4. **Check Permissions**: Ensure write access to target directories

### During Execution

1. **Review Conflicts**: In interactive mode, carefully review each conflict
2. **Monitor Logs**: Watch console output for errors or warnings
3. **Verify Statistics**: Check final statistics for unexpected results

### After Running

1. **Review Log Files**: Check generated log files for detailed operation history
2. **Test Your Application**: Ensure your C# application still compiles and runs
3. **Code Review**: Review changes before committing to version control
4. **Run Tests**: Execute your test suite to verify functionality

## Troubleshooting

### Common Issues

**Issue**: Tool shows "Method not found"
**Solution**: Verify method names in CSV exactly match those in source files (case-sensitive)

**Issue**: Attributes placed in wrong location  
**Solution**: Ensure HTTP attributes (`[HttpGet]`, etc.) are present above target methods

**Issue**: File permission errors
**Solution**: Check file/directory permissions and ensure write access

**Issue**: CSV parsing errors
**Solution**: Validate CSV format, check for proper escaping of quotes in attributes

### Debug Mode

For detailed debugging, check the generated log files which contain:

- Line-by-line processing details
- Exact search patterns used
- File modification timestamps
- Complete error stack traces

## License

This project is open-source. Please check the repository for specific license terms.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request with clear description

## Changelog

### Version 2.0.0

- Added comprehensive logging system
- Implemented interactive mode with user prompts
- Added execution statistics and timing
- Enhanced error handling and reporting
- Improved CSV validation
- Added preview mode functionality

### Version 1.0.0

- Initial release with basic attribute updating functionality
- Support for CSV-based mappings
- HTTP attribute detection and placement
