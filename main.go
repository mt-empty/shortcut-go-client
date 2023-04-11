package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// depending on the path, we might require sudo privileges
const PAGES_BASE_DIR = "/opt/shortcut/pages/"
const PAGES_FILE_EXT = ".md"
const ARCHIVE_NAME = "shortcuts.zip"
const DOWNLOAD_ARCHIVE_URL = "https://github.com/mt-empty/shortcut-pages/releases/latest/download/" + ARCHIVE_NAME

const ANSI_COLOUR_RESET_FG = "\x1b[39m"
const ANSI_COLOUR_TITLE_FG = "\x1b[39m"
const ANSI_COLOUR_EXPLANATION_FG = "\x1b[37m"
const ANSI_COLOUR_DESCRIPTION_FG = "\x1b[37m"
const ANSI_COLOUR_SHORTCUT_FG = "\x1b[32m"
const ANSI_COLOUR_CATEGORY_FG = "\x1b[37m"
const ANSI_BOLD_ON = "\x1b[1m"
const ANSI_BOLD_OFF = "\x1b[22m"

var Version = "development"

func main() {
	var noColourFlag bool
	var updateFlag bool
	var listFlag bool

	rootCmd := &cobra.Command{
		Use:     "shortcut <PROGRAM_NAME> [flags]",
		Short:   "A shortcut-pages client, pages directory is located at " + PAGES_BASE_DIR,
		Long:    ``,
		Version: Version,
		// Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if updateFlag {
				update()
			}

			if listFlag {
				listShortcuts()
			} else if len(args) > 0 {
				pageName := args[0]
				getShortcutPage(pageName, !noColourFlag)
			} else {
				cmd.Help()
			}
		},
	}
	rootCmd.Flags().BoolVarP(&noColourFlag, "no-colour", "n", false, "Remove colour from the output")
	rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List all available shortcut pages in the cache")
	rootCmd.Flags().BoolVarP(&updateFlag, "update", "u", false, "Update the local cache")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func requiresSudo(path string) bool {
	//checks whether writing a file to this path requires sudo privileges

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		// Check if error is "permission denied"
		if os.IsPermission(err) {
			return true
		}
		panic(err)
	}

	tempFile, err := os.Create(filepath.Join(path, "./tmp"))
	if err != nil {
		// Check if error is "permission denied"
		if os.IsPermission(err) {
			return true
		}
		panic(err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	return false
}

func update() {

	// check if sudo
	if requiresSudo(PAGES_BASE_DIR) {
		fmt.Fprintf(os.Stderr, "Writing to %s requires sudo privileges, please run with sudo\n", PAGES_BASE_DIR)
		os.Exit(1)
	}

	// download archive
	response, err := http.Get(DOWNLOAD_ARCHIVE_URL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error downloading archive: ", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	err = os.MkdirAll(PAGES_BASE_DIR, os.ModePerm)
	if err != nil {
		panic(err)
	}

	tempFile, err := os.Create(PAGES_BASE_DIR + ARCHIVE_NAME)
	if err != nil {
		panic(err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	io.Copy(tempFile, response.Body)

	// unzip archive
	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		panic(err)
	}

	for _, file := range zipReader.File {
		path := filepath.Join(PAGES_BASE_DIR, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		outputFile, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		defer outputFile.Close()

		zipFile, err := file.Open()
		if err != nil {
			panic(err)
		}
		defer zipFile.Close()

		_, err = io.Copy(outputFile, zipFile)
		if err != nil {
			panic(err)
		}
	}
	fmt.Fprintln(os.Stdout, "Successfully updated cache")
}

func listShortcuts() {
	// print list of files in a directory
	files, err := os.ReadDir(PAGES_BASE_DIR)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Shortcut pages directory at %s does not exist, please run shortcut update\n", PAGES_BASE_DIR)
			os.Exit(1)
		}
		panic(err)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Shortcut pages directory is empty, please run shortcut update")
	}

	for _, file := range files {
		fmt.Fprintln(os.Stdout, file.Name()[0:len(file.Name())-len(PAGES_FILE_EXT)])
	}

}

func getShortcutPage(programName string, IsColourOn bool) bool {
	cleanProgramName := filepath.Clean(programName)
	if cleanProgramName != filepath.Base(cleanProgramName) {
		fmt.Fprintf(os.Stderr, "Invalid program name \"%s\"\n", programName)
		return false
	}
	var programPath string = PAGES_BASE_DIR + cleanProgramName + PAGES_FILE_EXT
	// open a file
	file, err := os.Open(programPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No page available for \"%s\"\n", programName)
		return false
	}
	defer file.Close()
	parseShortcutPage(file, IsColourOn)
	return true

}

func parseShortcutPage(fileDescriptor *os.File, IsColourOn bool) bool {
	// use the file descriptor to read the file
	scanner := bufio.NewScanner(fileDescriptor)

	for scanner.Scan() {
		line := scanner.Text()
		var start bool = false
		for i, char := range line {
			if !start {
				switch char {
				case '#':
					if IsColourOn {
						fmt.Fprint(os.Stdout, ANSI_BOLD_ON+ANSI_COLOUR_TITLE_FG)
					}
					start = true
				case '$':
					if IsColourOn {
						fmt.Fprint(os.Stdout, ANSI_COLOUR_CATEGORY_FG)
					}
					start = true
				case '>':
					if IsColourOn {
						fmt.Fprint(os.Stdout, ANSI_COLOUR_EXPLANATION_FG)
					}
					start = true
				case '`':
					if IsColourOn {
						fmt.Fprint(os.Stdout, ANSI_COLOUR_SHORTCUT_FG)
					}
					start = true
				default:
					fmt.Fprintf(os.Stdout, "%c", char)
				}

			} else {
				switch char {
				case '{':
					if line[i+1] == '{' {
						if IsColourOn {
							fmt.Fprint(os.Stdout, ANSI_COLOUR_DESCRIPTION_FG)
						}
					} else if line[i-1] == '{' {
						continue
					} else {
						fmt.Fprintf(os.Stdout, "%c", char)
					}

				case '}':
					if line[i+1] == '}' {
						if IsColourOn {
							fmt.Fprint(os.Stdout, ANSI_COLOUR_DESCRIPTION_FG)
						}
					} else if line[i-1] == '}' {
						continue
					} else {
						fmt.Fprintf(os.Stdout, "%c", char)
					}

				default:
					fmt.Fprintf(os.Stdout, "%c", char)
				}
			}
		}
		fmt.Fprintln(os.Stdout, ANSI_BOLD_OFF+ANSI_COLOUR_RESET_FG)

	}
	return true
}
