package scan

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/spf13/cobra"
)

type commandConfig struct {
	filename  string
	format    outputFormat
	overwrite bool
	queueName string
	recursive bool

	awsProfile string
}

func (c *commandConfig) checkFlags() error {
	if c.format == outputFile && len(c.filename) == 0 {
		return fmt.Errorf("missing required flag --filename")
	}
	if c.format == outputQueue && c.queueName == "" {
		return fmt.Errorf("missing required AWS SQS queue name")
	}
	return nil
}

var (
	cfg commandConfig

	Cmd = &cobra.Command{
		Use:   "scan [folder] {folder2 ... folderN}",
		Short: "Scan folders for music",
		Args:  cobra.MinimumNArgs(1),
		RunE:  scan,
	}
)

func init() {
	// Set defaults
	cfg.format = outputStdOut
	cfg.overwrite = false
	cfg.recursive = true

	Cmd.Flags().StringVarP(&cfg.awsProfile, "aws_profile", "a", "", "Name of the AWS profile to use")
	Cmd.Flags().StringVarP(&cfg.filename, "filename", "f", "", "Name output file")
	Cmd.Flags().BoolVarP(&cfg.recursive, "recursive", "r", false, "Recursively process subfolders")
	Cmd.Flags().VarP(&cfg.format, "output_format", "o", `Output format type: "file", "queue", "stdout".`)
	Cmd.Flags().BoolVarP(&cfg.overwrite, "overwrite", "w", false, "Overwrite existing destination file")
	Cmd.Flags().StringVarP(&cfg.queueName, "queue_name", "q", "live-music", "Name of destination queue")
}

type FileAddOp interface {
	AddFile(filename string) error
	Close() error
}

func scan(cmd *cobra.Command, args []string) error {
	if err := cfg.checkFlags(); err != nil {
		return err
	}

	files, err := findFiles(args)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d files\n", len(files))

	var output FileAddOp
	switch cfg.format {
	case outputFile:
		output, err = newFileAddFileOut(cfg.filename, cfg.overwrite)
	case outputQueue:
		fmt.Println("Setting up AWS SQS connection...")
		output, err = newQueueOut(cfg.awsProfile, cfg.queueName)
	default:
		output = newFileAddStdOut()
	}
	if err != nil {
		return err
	}
	defer output.Close()

	for _, k := range slices.Sorted(maps.Keys(files)) {
		output.AddFile(k)
	}
	fmt.Println("Done")

	return nil
}

func findFiles(args []string) (map[string]bool, error) {
	mediaMatch := regexp.MustCompile(`[.](shn|flac|mp3)$`)

	files := map[string]bool{}
	for _, folder := range args {
		fmt.Printf("Processing folder %s\n", folder)

		fileInfo, err := os.Stat(folder)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("folder not found: %s", folder)
		}
		if err == nil && !fileInfo.IsDir() {
			return nil, fmt.Errorf("invalid folder: %s", folder)
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

	return files, nil
}
