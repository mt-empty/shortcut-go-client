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

const PAGES_BASE_DIR = "/opt/shortcut/pages/"
const PAGES_FILE_EXT = ".md"
const DOWNLOAD_ARCHIVE_URL = "https://github.com/mt-empty/shortcut-pages/releases/latest/download/shortcuts.zip"

const ANSI_COLOUR_RESET_FG = "\x1b[39m"
const ANSI_COLOUR_TITLE_FG = "\x1b[39m"
const ANSI_COLOUR_EXPLANATION_FG = "\x1b[37m"
const ANSI_COLOUR_DESCRIPTION_FG = "\x1b[37m"
const ANSI_COLOUR_SHORTCUT_FG = "\x1b[32m"
const ANSI_COLOUR_CATEGORY_FG = "\x1b[37m"
const ANSI_BOLD_ON = "\x1b[1m"
const ANSI_BOLD_OFF = "\x1b[22m"

func main() {
	var noColour bool

	rootCmd := &cobra.Command{
		Use:     "shortcut <PROGRAM_NAME> [flags]",
		Short:   "A shortcut-pages client, pages directory is located at " + PAGES_BASE_DIR,
		Long:    ``,
		Version: "1.0.0",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pageName := args[0]
			getShortcutPage(pageName, !noColour)

		},
	}
	rootCmd.Flags().BoolVarP(&noColour, "no-colour", "n", false, "Remove colour from the output")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available shortcut pages in the cache",
		Long:  ``,
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			listShortcuts()
		},
	}
	rootCmd.AddCommand(listCmd)

	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the local cache",
		Long:  ``,
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			update()
		},
	}
	rootCmd.AddCommand(updateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func update() {
	// download archive
	response, err := http.Get(DOWNLOAD_ARCHIVE_URL)
	if err != nil {
		fmt.Println("Error downloading archive: ", err)
		return
	}
	defer response.Body.Close()

	err = os.MkdirAll(PAGES_BASE_DIR, os.ModePerm)
	if err != nil {
		panic(err)
	}

	tempFile, err := os.Create(PAGES_BASE_DIR + "shortcuts.zip")
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
	fmt.Println("Successfully updated cache")
}

func listShortcuts() {
	// print list of files in a directory
	files, err := os.ReadDir(PAGES_BASE_DIR)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file.Name()[0 : len(file.Name())-len(PAGES_FILE_EXT)])
	}

}

func getShortcutPage(programName string, IsColourOn bool) bool {
	cleanProgramName := filepath.Clean(programName)
	if cleanProgramName != filepath.Base(cleanProgramName) {
		fmt.Printf("Invalid program name \"%s\"\n", programName)
		return false
	}
	var programPath string = PAGES_BASE_DIR + cleanProgramName + PAGES_FILE_EXT
	// open a file
	file, err := os.Open(programPath)
	if err != nil {
		fmt.Printf("No page available for \"%s\"\n", programName)
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
						fmt.Print(ANSI_BOLD_ON + ANSI_COLOUR_TITLE_FG)
					}
					start = true
				case '$':
					if IsColourOn {
						fmt.Print(ANSI_COLOUR_CATEGORY_FG)
					}
					start = true
				case '>':
					if IsColourOn {
						fmt.Print(ANSI_COLOUR_EXPLANATION_FG)
					}
					start = true
				case '`':
					if IsColourOn {
						fmt.Print(ANSI_COLOUR_SHORTCUT_FG)
					}
					start = true
				default:
					fmt.Printf("%c", char)
				}

			} else {
				switch char {
				case '{':
					if line[i+1] == '{' {
						if IsColourOn {
							fmt.Print(ANSI_COLOUR_DESCRIPTION_FG)
						}
					} else if line[i-1] == '{' {
						continue
					} else {
						fmt.Printf("%c", char)
					}

				case '}':
					if line[i+1] == '}' {
						if IsColourOn {
							fmt.Print(ANSI_COLOUR_DESCRIPTION_FG)
						}
					} else if line[i-1] == '}' {
						continue
					} else {
						fmt.Printf("%c", char)
					}

				default:
					fmt.Printf("%c", char)
				}
			}
		}
		fmt.Print(ANSI_BOLD_OFF)
		fmt.Print(ANSI_COLOUR_RESET_FG)
		fmt.Println()

	}
	return true
}
