package analyze

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/spf13/cobra"
)

type commandConfig struct {
	source     Source
	sourceFile string

	awsProfile string
	mongoURI   string
	queueName  string
}

var (
	cfg commandConfig

	Cmd = &cobra.Command{
		Use:   "analyze [file1] {file2 ... fileN}",
		Short: "Analyze audio files and extract metadata",
		Args:  cobra.ArbitraryArgs,
		RunE:  analyze,
	}
)

func checkFlags(cfg *commandConfig) error {
	cfg.mongoURI = cmp.Or(cfg.mongoURI, os.Getenv("MONGODB_URI"))
	if cfg.mongoURI == "" {
		return fmt.Errorf("missing required MongoDB connection string")
	}

	if cfg.source == SourceFile && cfg.sourceFile == "" {
		return fmt.Errorf("missing required --file flag")
	}
	if cfg.source == SourceSQS && cfg.queueName == "" {
		return fmt.Errorf("missing required --queue_name flag")
	}

	return nil
}

func init() {
	Cmd.Flags().StringVarP(&cfg.awsProfile, "aws_profile", "a", "", "Name of the AWS profile to use")
	Cmd.Flags().StringVarP(&cfg.mongoURI, "mongodb_uri", "m", "", "MongoDB connection string")
	Cmd.Flags().StringVarP(&cfg.queueName, "queue_name", "q", "live-music", "Name of destination queue")
	Cmd.Flags().VarP(&cfg.source, "source", "s", fmt.Sprintf("Source of files to analyze: %s", strings.Join(sourceNames(), ",")))
	Cmd.Flags().StringVarP(&cfg.sourceFile, "file", "f", "", "Filename containing a list of files to analyze")
}

type SourceHandler interface {
	AnalyzeFiles(fn AnalyzerFn) error
	Close() error
}

type AnalyzerFn func(filename string) error

func analyze(cmd *cobra.Command, args []string) error {
	if err := checkFlags(&cfg); err != nil {
		return err
	}

	ctx := cmp.Or(cmd.Context(), context.Background())

	fmt.Println("Setting up MongoDB connection...")
	storage, err := newStorageHandler(cfg.mongoURI)
	if err != nil {
		return fmt.Errorf("error loading storage handler for %q: %v", cfg.mongoURI, err)
	}
	defer func() error {
		if err := storage.Close(ctx); err != nil {
			return fmt.Errorf("error disconnecting from MongoDB: %v", err)
		}
		return nil
	}()

	for _, f := range args {
		analyzeFile(storage, f)
	}

	var handler SourceHandler
	if cfg.source == SourceFile {
		handler, err = newFileSourceHandler(cfg.sourceFile)
		if err != nil {
			return fmt.Errorf("error creating file source %s: %v", cfg.sourceFile, err)
		}
	} else if cfg.source == SourceSQS {
		handler, err = newSQSSource(cfg.awsProfile, cfg.queueName)
		if err != nil {
			return fmt.Errorf("error creating SQS source for %s: %v", cfg.queueName, err)
		}
	}
	defer handler.Close()

	handler.AnalyzeFiles(func(filename string) error {
		return analyzeFile(storage, filename)
	})

	return nil
}

type Metadata struct {
	Id       string    `json:"id" bson:"_id"`
	Filename string    `json:"filename" bson:"filename"`
	Album    string    `json:"album" bson:"album"`
	Artist   string    `json:"artist" bson:"artist"`
	Date     time.Time `json:"date" bson:"date,omitempty"`
	Disc     int       `json:"disc" bson:"disc,omitempty"`
	Genre    []string  `json:"genre" bson:"genre,omitempty"`
	Set      int       `json:"set" bson:"set,omitempty"`
	Title    string    `json:"title" bson:"title"`
	Track    int       `json:"track" bson:"track,omitempty"`
	Venue    string    `json:"venue" bson:"venue,omitempty"`

	AccousticIdFingerprint string      `json:"accoustic_id_fingerprint" bson:"accoustic_id_fingerprint,omitempty"`
	MusicBrainz            MusicBrainz `json:"music_brainz" bson:"music_brainz,omitempty"`

	Tags map[string]string `json:"tags" bson:"tags,omitempty"`
}

type MusicBrainz struct {
	// Unique ID of the artist or band. For example, Rush has the artist id of
	// "534ee493-bfac-4575-a44a-0ae41e2c3fe4".
	ArtistId string `json:"artist_id" bson:"artist_id"`

	// Release group is what most people would call an "album". For example,
	// the album titled "Roll the Bones" by Rush has an release group id of
	// "e188de4e-6d15-3ca3-be49-fa13c67a03c0".
	ReleaseGroupId string `json:"release_group_id" bson:"release_group_id"`

	// Release is a specific edition of an album. For example, the album
	// titled "Roll the Bones" by Rush has had at least 13 different releases
	// around the world. The original US CD release by Atlantic on 1991-09-03
	// has the release id of "50e551bd-5d24-37e5-913d-07c25cd85e8e". Whereas
	// the original 12" vinyl release by Atlantic was worldwide and has the
	// release id of "52bf9926-dc7f-40b9-9a08-d5f0c98f8a63".
	ReleaseId string `json:"release_id" bson:"release_id"`
}

func newMetadata(filename string, md tag.Metadata) *Metadata {
	trackNum, _ := md.Track()
	id := fmt.Sprintf("%s_%s_%s_%04d", md.Artist(), md.Album(), filename, trackNum)

	cleanupRegEx := regexp.MustCompile(`[,_ ]+`)
	id = cleanupRegEx.ReplaceAllString(id, "-")

	m := Metadata{
		Id:       strings.ToLower(id),
		Filename: filename,
		Album:    md.Album(),
		Artist:   md.Artist(),
		Tags:     make(map[string]string),
		Title:    md.Title(),
	}

	genre := strings.Split(md.Genre(), ";")
	if len(genre) <= 1 {
		genre = strings.Split(md.Genre(), ",")
	}
	for _, g := range genre {
		m.Genre = append(m.Genre, strings.TrimSpace(g))
	}

	m.Disc, _ = md.Disc()
	m.Track, _ = md.Track()

	if value, ok := md.Raw()["date"]; ok {
		strValue := fmt.Sprintf("%v", value)

		var err error
		// Try year month day
		if m.Date, err = time.Parse(time.DateOnly, strValue); err != nil {
			// Try year only
			if m.Date, err = time.Parse("2006", strValue); err != nil {
				fmt.Printf("WARNING: unable to parse date %q for %s\n", strValue, filename)
			}
		}
	}

	for k, v := range md.Raw() {
		strValue := fmt.Sprintf("%v", v)
		m.Tags[k] = strValue

		var err error
		switch strings.ToLower(k) {
		case "acoustid_fingerprint":
			m.AccousticIdFingerprint = strValue
		case "musicbrainz_albumid":
			m.MusicBrainz.ReleaseId = strValue
		case "musicbrainz_artistid":
			m.MusicBrainz.ArtistId = strValue
		case "musicbrainz_releasegroupid":
			m.MusicBrainz.ReleaseGroupId = strValue
		case "set":
			if m.Set, err = strconv.Atoi(strValue); err != nil {
				fmt.Printf("WARNING: unable to parse set # %q for %s\n", strValue, filename)
				continue
			}
		case "venue":
			m.Venue = strValue
		}
	}

	return &m
}

func analyzeFile(storage *StorageHandler, filename string) error {
	fmt.Printf("Processing %s\n", filename)

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", filename, err)
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return fmt.Errorf("error reading tags from %s: %v", filename, err)
	}

	metadata := newMetadata(filepath.Base(filename), m)
	if err = storage.SaveMetadata(context.Background(), metadata); err != nil {
		return err
	}

	// b, err := json.MarshalIndent(metadata, "", "  ")
	// if err != nil {
	// 	return fmt.Errorf("unable to marshal %s due to %s", filename, err)
	// }
	// fmt.Println(string(b))

	return nil
}
