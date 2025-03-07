package scan

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/spf13/cobra"
)

var (
	dest string

	overwrite bool
	recursive bool

	Cmd = &cobra.Command{
		Use:   "scan [folder] {folder2 ... folderN}",
		Short: "Scan folders for music",
		Args:  cobra.MinimumNArgs(1),
		RunE:  scan,
	}
)

func init() {
	Cmd.Flags().StringVarP(&dest, "dest", "d", "", "Filename for output destination")
	Cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively process subfolders")
	Cmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite existing destination file")
}

func scan(cmd *cobra.Command, args []string) error {
	mediaMatch := regexp.MustCompile(`[.](shn|flac|mp3)$`)

	files := map[string]bool{}
	for _, folder := range args {
		fmt.Printf("Processing folder %s\n", folder)

		fileInfo, err := os.Stat(folder)
		if os.IsNotExist(err) {
			return fmt.Errorf("folder not found: %s", folder)
		}
		if err == nil && !fileInfo.IsDir() {
			return fmt.Errorf("invalid folder: %s", folder)
		}

		filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() && mediaMatch.MatchString(path) {
				if _, exists := files[path]; !exists {
					files[path] = true
				}
			}
			return nil
		})
	}

	fmt.Printf("Found %d files\n", len(files))
	if dest != "" {
		// TODO: Allow writing to stdout
		if _, err := os.Stat(dest); err != nil {
			if (errors.Is(err, fs.ErrExist) && !overwrite) || !errors.Is(err, fs.ErrExist) {
				return fmt.Errorf("error checking output file %s: %v", dest, err)
			}
		}

		fmt.Printf("Writing file list to %s\n", dest)
		f, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer f.Close()

		writer := bufio.NewWriter(f)
		for _, k := range slices.Sorted(maps.Keys(files)) {
			fmt.Fprintf(writer, "%s\n", k)
		}
		writer.Flush()
	}

	return nil
}
