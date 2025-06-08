package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

type Stats struct {
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
var stats Stats

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

func promptUser(message string) bool {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return false
	}

	return input == "y" || input == "yes"
}

func printStats() {
	duration := time.Since(stats.StartTime)

	writeLog("=" + strings.Repeat("=", 50))
	writeLog("EXECUTION SUMMARY")
	writeLog("=" + strings.Repeat("=", 50))
	writeLog(fmt.Sprintf("Execution Time: %v", duration))
	writeLog(fmt.Sprintf("Total Files Processed: %d", stats.TotalFiles))
	writeLog(fmt.Sprintf("Files Modified: %d", stats.FilesModified))
	writeLog(fmt.Sprintf("Files Skipped: %d", stats.FilesSkipped))
	writeLog(fmt.Sprintf("Attributes Added: %d", stats.AttributesAdded))
	writeLog(fmt.Sprintf("Attributes Replaced: %d", stats.AttributesReplaced))
	writeLog(fmt.Sprintf("Errors Encountered: %d", stats.Errors))
	writeLog("=" + strings.Repeat("=", 50))
}

func main() {
	stats.StartTime = time.Now()

	preview := flag.Bool("preview", false, "Preview changes without modifying files.")
	overwrite := flag.Bool("overwrite", false, "Overwrite existing attributes without prompting.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "C# Authorize Adapter\n\n")
		fmt.Fprintf(os.Stderr, "This tool reads a CSV file to find and update specific C# controller methods with a new attribute.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <csv-file-path> <controllers-directory>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <csv-file-path>          Path to the input CSV file.\n")
		fmt.Fprintf(os.Stderr, "  <controllers-directory>  Path to the root directory containing the C# controller files.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nModes:\n")
		fmt.Fprintf(os.Stderr, "  Default (Interactive): Prompts user for confirmation when conflicts are found.\n")
		fmt.Fprintf(os.Stderr, "                        Press Enter or type 'n' to skip, type 'y' to proceed.\n")
		fmt.Fprintf(os.Stderr, "  --overwrite:          Automatically overwrites existing attributes without prompting.\n")
		fmt.Fprintf(os.Stderr, "  --preview:            Shows what changes would be made without modifying files.\n\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  %s ./mappings.csv ./MyProject/Controllers\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --overwrite ./mappings.csv ./MyProject/Controllers\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --preview ./mappings.csv ./MyProject/Controllers\n", os.Args[0])
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Missing required arguments.")
		flag.Usage()
		os.Exit(1)
	}

	csvPath := args[0]
	controllersDir := args[1]

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
		defer logFile.Close()
		writeLogOnly(fmt.Sprintf("Log file created: %s", logFilePath))
		writeLogOnly("Migro - C# Attribute Updater - Execution Started")
		writeLogOnly("=" + strings.Repeat("=", 50))
	}

	writeLog(fmt.Sprintf("Starting C# Attribute Updater"))
	writeLog(fmt.Sprintf("CSV File: %s", csvPath))
	writeLog(fmt.Sprintf("Controllers Directory: %s", controllersDir))

	if *preview {
		writeLog("Mode: PREVIEW - No files will be modified")
	} else if *overwrite {
		writeLog("Mode: OVERWRITE - Existing attributes will be replaced automatically")
	} else {
		writeLog("Mode: INTERACTIVE - Will prompt for confirmation on conflicts")
	}

	writeLog("-" + strings.Repeat("-", 50))

	mappings, err := readCsvMappings(csvPath)
	if err != nil {
		writeLog(fmt.Sprintf("FATAL ERROR: Failed to read CSV file '%s': %v", csvPath, err))
		stats.Errors++
		printStats()
		os.Exit(1)
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

		fileModified := false
		methodFound := false

		for i, line := range lines {
			if strings.Contains(line, " "+mapping.Method+"(") {
				methodFound = true

				var authorizeIndices []int
				var httpAttrIndex = -1
				var attrBlockStartIndex = -1

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
				for _, char := range lines[httpAttrIndex] {
					if char == ' ' || char == '\t' {
						indent += string(char)
					} else {
						break
					}
				}
				newAttrLine := indent + mapping.Attribute

				isAlreadyCorrect := len(authorizeIndices) == 1 && strings.TrimSpace(lines[authorizeIndices[0]]) == strings.TrimSpace(newAttrLine)

				if isAlreadyCorrect {
					writeLog(fmt.Sprintf("INFO: Attribute already correct for %s:%s.", mapping.Filename, mapping.Method))
					break
				}

				if len(authorizeIndices) > 0 {
					writeLog(fmt.Sprintf("📄 File: %s", mapping.Filename))
					writeLog(fmt.Sprintf("🔧 Method: %s", mapping.Method))
					writeLog("🔁 Found existing attribute(s) to replace/clean up:")

					for k := len(authorizeIndices) - 1; k >= 0; k-- {
						writeLog(fmt.Sprintf("   OLD: %s", strings.TrimSpace(lines[authorizeIndices[k]])))
					}
					writeLog(fmt.Sprintf("   NEW: %s", strings.TrimSpace(newAttrLine)))
				} else {
					writeLog(fmt.Sprintf("📄 File: %s", mapping.Filename))
					writeLog(fmt.Sprintf("🔧 Method: %s", mapping.Method))
					writeLog(fmt.Sprintf("➕ Inserting new attribute for method"))
					writeLog(fmt.Sprintf("   NEW: %s", strings.TrimSpace(newAttrLine)))
				}

				if *preview {
					writeLog("   [PREVIEW MODE] - No changes will be applied.")
					break
				}

				applyChange := true
				if len(authorizeIndices) > 0 {
					if *overwrite {
						writeLog("   👉 Overwriting due to --overwrite flag.")
					} else {
						applyChange = promptUser("❓ Do you want to apply this change? (y/N): ")
						if applyChange {
							writeLog("   ✅ User confirmed change.")
						} else {
							writeLog("   ❌ User declined change.")
						}
					}
				}

				if !applyChange {
					break
				}

				var newLines []string
				newLines = append(newLines, lines[:attrBlockStartIndex]...)
				newLines = append(newLines, newAttrLine)
				for k := attrBlockStartIndex; k < i; k++ {
					isOldAuthorize := slices.Contains(authorizeIndices, k)
					if !isOldAuthorize {
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

		if fileModified && !*preview {
			output := strings.Join(lines, "\n") + "\n"
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

	if *preview {
		writeLog("✨ Preview complete. No files were modified.")
	} else {
		writeLog("✨ All operations complete.")
	}

	printStats()

	if logFile != nil {
		writeLog(fmt.Sprintf("📝 Detailed log saved to: %s/%s", logDir, logFileName))
	}
}
