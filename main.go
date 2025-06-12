package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

type CsvMapping struct {
	Filename   string
	Controller string
	Method     string
	Attribute  string
}

type ScannerStats struct {
	TotalFiles     int
	FilesProcessed int
	FilesSkipped   int
	MethodsFound   int
	Errors         int
	StartTime      time.Time
}

type UpdaterStats struct {
	TotalFiles         int
	FilesModified      int
	FilesSkipped       int
	AttributesAdded    int
	AttributesReplaced int
	Errors             int
	StartTime          time.Time
}

var httpAttributes = []string{
	"[HttpGet]", "[HttpGet(",
	"[HttpPost]", "[HttpPost(",
	"[HttpPut]", "[HttpPut(",
	"[HttpDelete]", "[HttpDelete(",
	"[HttpPatch]", "[HttpPatch(",
}

var logFile *os.File

func setupLogger() {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: Could not create log directory: %v", err)
		logDir = "."
	}

	logFileName := fmt.Sprintf("migro_%s.log", time.Now().Format("20060102_150405"))
	logFilePath := filepath.Join(logDir, logFileName)

	var err error
	logFile, err = os.Create(logFilePath)
	if err != nil {
		log.Printf("Warning: Could not create log file: %v", err)
	} else {
		writeLogOnly(fmt.Sprintf("Log file created: %s", logFilePath))
		writeLogOnly("Migro - Execution Started")
		writeLogOnly("=" + strings.Repeat("=", 50))
	}
	fmt.Printf("📝 Detailed log will be saved to: %s\n\n", logFilePath)
}

func closeLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func writeLog(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
	fmt.Print(logMessage)
	if logFile != nil {
		logFile.WriteString(logMessage)
	}
}

func writeLogOnly(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
	if logFile != nil {
		logFile.WriteString(logMessage)
	}
}

func isHttpAttribute(line string) bool {
	line = strings.TrimSpace(line)
	for _, attr := range httpAttributes {
		if strings.HasPrefix(line, attr) {
			return true
		}
	}
	return false
}

func isAuthorizeAttribute(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return strings.HasPrefix(trimmedLine, "[Authorize")
}

func promptUser(message string) bool {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func promptForInput(promptText string) string {
	fmt.Print(promptText)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func extractControllerName(filename string) string {
	return strings.TrimSuffix(filename, ".cs")
}

func extractMethodName(line string) string {
	re := regexp.MustCompile(`\b(public|private|protected|internal)\s+(?:static\s+)?(?:async\s+)?(?:Task<?[^>]*>?\s+|[A-Za-z_][A-Za-z0-9_<>,\[\]]*\s+)([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		return matches[2]
	}
	return ""
}

func scanControllerFile(filePath string, stats *ScannerStats) ([]CsvMapping, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var methods []CsvMapping
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)
	controllerName := extractControllerName(filename)
	writeLogOnly(fmt.Sprintf("Scanning file: %s (Controller: %s)", filename, controllerName))

	for i, line := range lines {
		methodName := extractMethodName(line)
		if methodName == "" {
			continue
		}

		hasHttpAttribute := false
		var authorizeAttribute string
		for j := i - 1; j >= 0; j-- {
			trimmedLine := strings.TrimSpace(lines[j])
			if trimmedLine == "" {
				continue
			}
			if strings.HasPrefix(trimmedLine, "[") {
				if isHttpAttribute(trimmedLine) {
					hasHttpAttribute = true
				}
				if isAuthorizeAttribute(trimmedLine) {
					authorizeAttribute = trimmedLine
				}
			} else {
				break
			}
		}

		if hasHttpAttribute {
			attribute := `[Authorize(Roles = "")]`
			if authorizeAttribute != "" {
				attribute = authorizeAttribute
			}
			methods = append(methods, CsvMapping{
				Filename:   filename,
				Controller: controllerName,
				Method:     methodName,
				Attribute:  attribute,
			})
			writeLogOnly(fmt.Sprintf("Found method: %s in %s", methodName, filename))
			stats.MethodsFound++
		}
	}
	return methods, nil
}

func scanDirectory(dirPath string, stats *ScannerStats) ([]CsvMapping, error) {
	var allMethods []CsvMapping
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			writeLog(fmt.Sprintf("ERROR: Accessing path %s: %v", path, err))
			stats.Errors++
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".cs") {
			stats.TotalFiles++
			methods, err := scanControllerFile(path, stats)
			if err != nil {
				writeLog(fmt.Sprintf("ERROR: Scanning file %s: %v", path, err))
				stats.FilesSkipped++
				stats.Errors++
				return nil
			}
			if len(methods) > 0 {
				allMethods = append(allMethods, methods...)
				stats.FilesProcessed++
				writeLog(fmt.Sprintf("✅ Processed: %s (%d methods found)", info.Name(), len(methods)))
			} else {
				writeLog(fmt.Sprintf("ℹ️  Skipped: %s (no HTTP methods found)", info.Name()))
				stats.FilesSkipped++
			}
		}
		return nil
	})
	return allMethods, err
}

func writeCsv(methods []CsvMapping, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"filename", "controller", "method", "attribute"}); err != nil {
		return err
	}
	for _, method := range methods {
		record := []string{method.Filename, method.Controller, method.Method, method.Attribute}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func printScannerStats(stats ScannerStats) {
	duration := time.Since(stats.StartTime)
	writeLog("=" + strings.Repeat("=", 50))
	writeLog("SCAN SUMMARY")
	writeLog("=" + strings.Repeat("=", 50))
	writeLog(fmt.Sprintf("Execution Time: %v", duration))
	writeLog(fmt.Sprintf("Total Files Found: %d", stats.TotalFiles))
	writeLog(fmt.Sprintf("Files Processed: %d", stats.FilesProcessed))
	writeLog(fmt.Sprintf("Files Skipped: %d", stats.FilesSkipped))
	writeLog(fmt.Sprintf("Methods Found: %d", stats.MethodsFound))
	writeLog(fmt.Sprintf("Errors Encountered: %d", stats.Errors))
	writeLog("=" + strings.Repeat("=", 50))
}

func runScanner(controllersDir, outputCsvPath string) {
	stats := ScannerStats{StartTime: time.Now()}
	writeLog("Starting Scanner...")
	writeLog(fmt.Sprintf("Controllers Directory: %s", controllersDir))
	writeLog(fmt.Sprintf("Output CSV Path: %s", outputCsvPath))
	writeLog("-" + strings.Repeat("-", 50))

	if _, err := os.Stat(controllersDir); os.IsNotExist(err) {
		writeLog(fmt.Sprintf("FATAL ERROR: Controllers directory does not exist: %s", controllersDir))
		stats.Errors++
		printScannerStats(stats)
		return
	}

	methods, err := scanDirectory(controllersDir, &stats)
	if err != nil {
		writeLog(fmt.Sprintf("ERROR: Failed to scan directory: %v", err))
		stats.Errors++
	}

	if len(methods) == 0 {
		writeLog("WARNING: No HTTP methods found in any controller files.")
	} else {
		writeLog(fmt.Sprintf("Found %d HTTP methods across %d controller files", len(methods), stats.FilesProcessed))
		if err := writeCsv(methods, outputCsvPath); err != nil {
			writeLog(fmt.Sprintf("FATAL ERROR: Failed to write CSV file: %v", err))
			stats.Errors++
		} else {
			writeLog(fmt.Sprintf("✅ Successfully generated CSV file: %s", outputCsvPath))
		}
	}
	printScannerStats(stats)
}

func readCsvMappings(csvPath string) ([]CsvMapping, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var mappings []CsvMapping
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			writeLog(fmt.Sprintf("WARNING: Skipping incomplete row in CSV: %v", row))
			continue
		}
		mappings = append(mappings, CsvMapping{
			Filename:   strings.TrimSpace(row[0]),
			Controller: strings.TrimSpace(row[1]),
			Method:     strings.TrimSpace(row[2]),
			Attribute:  strings.TrimSpace(row[3]),
		})
	}
	return mappings, nil
}

func printUpdaterStats(stats UpdaterStats) {
	duration := time.Since(stats.StartTime)
	writeLog("=" + strings.Repeat("=", 50))
	writeLog("UPDATE SUMMARY")
	writeLog("=" + strings.Repeat("=", 50))
	writeLog(fmt.Sprintf("Execution Time: %v", duration))
	writeLog(fmt.Sprintf("Total Mappings Processed: %d", stats.TotalFiles))
	writeLog(fmt.Sprintf("Files Modified: %d", stats.FilesModified))
	writeLog(fmt.Sprintf("Files Skipped/No Change: %d", stats.FilesSkipped))
	writeLog(fmt.Sprintf("Attributes Added: %d", stats.AttributesAdded))
	writeLog(fmt.Sprintf("Attributes Replaced: %d", stats.AttributesReplaced))
	writeLog(fmt.Sprintf("Errors Encountered: %d", stats.Errors))
	writeLog("=" + strings.Repeat("=", 50))
}

func runUpdater(csvPath, controllersDir string, preview, overwrite bool) {
	stats := UpdaterStats{StartTime: time.Now()}

	writeLog("Starting Attribute Updater...")
	writeLog(fmt.Sprintf("CSV File: %s", csvPath))
	writeLog(fmt.Sprintf("Controllers Directory: %s", controllersDir))
	if preview {
		writeLog("Mode: PREVIEW - No files will be modified")
	} else if overwrite {
		writeLog("Mode: OVERWRITE - Existing attributes will be replaced automatically")
	} else {
		writeLog("Mode: INTERACTIVE - Will prompt for confirmation on conflicts")
	}
	writeLog("-" + strings.Repeat("-", 50))

	mappings, err := readCsvMappings(csvPath)
	if err != nil {
		writeLog(fmt.Sprintf("FATAL ERROR: Failed to read CSV file '%s': %v", csvPath, err))
		stats.Errors++
		printUpdaterStats(stats)
		return
	}
	writeLog(fmt.Sprintf("Loaded %d mappings from CSV", len(mappings)))

	for _, mapping := range mappings {
		stats.TotalFiles++
		filePath := filepath.Join(controllersDir, mapping.Filename)

		writeLogOnly(fmt.Sprintf("Processing file: %s, Method: %s", mapping.Filename, mapping.Method))

		file, err := os.Open(filePath)
		if err != nil {
			writeLog(fmt.Sprintf("ERROR: Skipping file %s: %v", mapping.Filename, err))
			stats.FilesSkipped++
			stats.Errors++
			continue
		}

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		file.Close()

		if err := scanner.Err(); err != nil {
			writeLog(fmt.Sprintf("ERROR: Reading file %s: %v", mapping.Filename, err))
			stats.FilesSkipped++
			stats.Errors++
			continue
		}

		fileModified, methodFound := false, false

		for i, line := range lines {
			if strings.Contains(line, " "+mapping.Method+"(") {
				methodFound = true
				var authorizeIndices []int
				httpAttrIndex, attrBlockStartIndex := -1, -1

				for j := i - 1; j >= 0; j-- {
					trimmedLine := strings.TrimSpace(lines[j])
					if trimmedLine == "" {
						continue
					}
					if strings.HasPrefix(trimmedLine, "[") {
						attrBlockStartIndex = j
						if isHttpAttribute(trimmedLine) {
							httpAttrIndex = j
						}
						if isAuthorizeAttribute(trimmedLine) {
							authorizeIndices = append(authorizeIndices, j)
						}
					} else {
						break
					}
				}

				if httpAttrIndex == -1 {
					writeLog(fmt.Sprintf("WARNING: Could not find an HTTP attribute for method '%s' in '%s'. Skipping.", mapping.Method, mapping.Filename))
					break
				}
				if attrBlockStartIndex == -1 {
					attrBlockStartIndex = httpAttrIndex
				}

				indent := ""
				re := regexp.MustCompile(`^(\s*)`)
				matches := re.FindStringSubmatch(lines[httpAttrIndex])
				if len(matches) > 1 {
					indent = matches[1]
				}

				newAttrLine := indent + mapping.Attribute
				isAlreadyCorrect := len(authorizeIndices) == 1 && strings.TrimSpace(lines[authorizeIndices[0]]) == strings.TrimSpace(newAttrLine)

				if isAlreadyCorrect {
					writeLog(fmt.Sprintf("INFO: Attribute already correct for %s:%s.", mapping.Filename, mapping.Method))
					break
				}

				if len(authorizeIndices) > 0 {
					writeLog(fmt.Sprintf("📄 File: %s, 🔧 Method: %s", mapping.Filename, mapping.Method))
					writeLog("🔁 Found existing attribute(s) to replace/clean up:")
					for k := len(authorizeIndices) - 1; k >= 0; k-- {
						writeLog(fmt.Sprintf("   OLD: %s", strings.TrimSpace(lines[authorizeIndices[k]])))
					}
					writeLog(fmt.Sprintf("   NEW: %s", strings.TrimSpace(newAttrLine)))
				} else {
					writeLog(fmt.Sprintf("📄 File: %s, 🔧 Method: %s", mapping.Filename, mapping.Method))
					writeLog(fmt.Sprintf("➕ Inserting new attribute: %s", strings.TrimSpace(newAttrLine)))
				}

				if preview {
					writeLog("   [PREVIEW MODE] - No changes will be applied.")
					break
				}

				var applyChange bool
				if overwrite {
					applyChange = true
					writeLog("   👉 Applying change automatically due to --overwrite flag.")
				} else {
					// This is INTERACTIVE mode
					applyChange = promptUser("❓ Do you want to apply this change? (y/N): ")
					if applyChange {
						writeLog("   ✅ User confirmed change.")
					} else {
						writeLog("   ❌ User declined change.")
					}
				}

				if !applyChange {
					break
				}

				var newLines []string
				newLines = append(newLines, lines[:attrBlockStartIndex]...)
				newLines = append(newLines, newAttrLine)
				for k := attrBlockStartIndex; k < i; k++ {
					if !slices.Contains(authorizeIndices, k) {
						newLines = append(newLines, lines[k])
					}
				}
				newLines = append(newLines, lines[i:]...)
				lines = newLines
				fileModified = true

				if len(authorizeIndices) > 0 {
					stats.AttributesReplaced++
				} else {
					stats.AttributesAdded++
				}
				break
			}
		}

		if !methodFound {
			writeLog(fmt.Sprintf("WARNING: Method '%s' not found in file '%s'", mapping.Method, mapping.Filename))
			stats.FilesSkipped++
		}

		if fileModified && !preview {
			output := strings.Join(lines, "\n")
			// Ensure file ends with a newline
			if !strings.HasSuffix(output, "\n") {
				output += "\n"
			}
			err := os.WriteFile(filePath, []byte(output), 0644)
			if err != nil {
				writeLog(fmt.Sprintf("ERROR: Writing updated file %s: %v", filePath, err))
				stats.Errors++
				continue
			}
			writeLog(fmt.Sprintf("✅ Successfully updated: %s", mapping.Filename))
			stats.FilesModified++
		} else if !fileModified && methodFound {
			stats.FilesSkipped++
		}
	}

	if preview {
		writeLog("✨ Preview complete. No files were modified.")
	} else {
		writeLog("✨ All operations complete.")
	}
	printUpdaterStats(stats)
}

func printBanner() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(" C# Controller Authorize Attribute Scanner & Updater ")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
}

func handleScan() {
	fmt.Println("\n--- Scan Controllers ---")
	fmt.Println("This tool scans C# controller files and generates a CSV")
	fmt.Println("template with all HTTP methods found.")
	fmt.Println()

	controllersDir := promptForInput("Enter the path to your controllers directory: ")
	if controllersDir == "" {
		fmt.Println("Path cannot be empty. Aborting.")
		return
	}
	outputCsv := promptForInput("Enter the path for the output CSV file (e.g., ./mappings.csv): ")
	if outputCsv == "" {
		fmt.Println("Path cannot be empty. Aborting.")
		return
	}

	fmt.Println()
	runScanner(controllersDir, outputCsv)
}

func handleUpdate() {
	fmt.Println("\n--- Update Attributes from CSV ---")
	fmt.Println("This tool reads a CSV file to update C# controller methods")
	fmt.Println("with [Authorize] attributes.")
	fmt.Println()

	csvPath := promptForInput("Enter the path to your input CSV file: ")
	if csvPath == "" {
		fmt.Println("Path cannot be empty. Aborting.")
		return
	}
	controllersDir := promptForInput("Enter the path to your controllers directory: ")
	if controllersDir == "" {
		fmt.Println("Path cannot be empty. Aborting.")
		return
	}

	fmt.Println("\nSelect an operation mode:")
	fmt.Println("  1. Interactive (Default, asks for confirmation on overwrites)")
	fmt.Println("  2. Overwrite (Automatically replaces existing attributes)")
	fmt.Println("  3. Preview (Show changes without modifying files)")
	modeChoice := promptForInput("Enter choice (1-3): ")

	var isPreview, isOverwrite bool
	switch modeChoice {
	case "2":
		isOverwrite = true
	case "3":
		isPreview = true
	default:
	}

	fmt.Println()
	runUpdater(csvPath, controllersDir, isPreview, isOverwrite)
}

func main() {
	setupLogger()
	defer closeLogger()

	for {
		printBanner()
		fmt.Println("What would you like to do?")
		fmt.Println("  1. Scan Controllers to generate a CSV")
		fmt.Println("  2. Update Controllers from a CSV")
		fmt.Println("  3. Exit")

		choice := promptForInput("\nEnter your choice (1-3): ")

		switch choice {
		case "1":
			handleScan()
		case "2":
			handleUpdate()
		case "3":
			fmt.Println("\nExiting Migro. Goodbye!")
			return
		default:
			fmt.Println("\nInvalid choice. Please enter 1, 2, or 3.")
		}

		promptForInput("\nPress Enter to return to the main menu...")
		fmt.Print("\033[H\033[2J")
	}
}
