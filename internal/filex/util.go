package filex

import (
	"os"
	"path/filepath"
)

// SearchFilesBySuffix searches for files with a specific suffix in the given directory
// Parameters:
//
//	dir - The directory path to search in
//	suffix - The file suffix to search for (e.g. ".txt", ".go")
//
// Returns:
//
//	[]string - A slice of file paths that match the suffix
//	error - An error if the search fails
func SearchFilesBySuffix(dir string, suffix string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == suffix {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
