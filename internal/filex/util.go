package filex

import (
	"os"
	"path/filepath"

	"github.com/rambollwong/rainbowcat/util"
)

// SearchFilesBySuffix searches for files with a specific suffix in the given directory
// Parameters:
//
//	dir - The directory path to search in
//	suffix - The file suffix to search for (e.g. ".txt", ".m3u")
//
// Returns:
//
//	[]string - A slice of file paths that match the suffix
//	error - An error if the search fails
func SearchFilesBySuffix(dir string, suffix ...string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && util.SliceContains(suffix, filepath.Ext(path)) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// WriteBytesToFile writes the given bytes to the specified file path.
// If the file or its parent directories do not exist, they will be created.
// Parameters:
//
//	data - The byte slice to write to the file
//	filePath - The path of the file to write to
//
// Returns:
//
//	error - An error if the write operation fails
func WriteBytesToFile(data []byte, filePath string) error {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	// Write data to file
	return os.WriteFile(filePath, data, 0644)
}
