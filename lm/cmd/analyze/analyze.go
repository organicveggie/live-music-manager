package analyze

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/spf13/cobra"
)

var (
	listFile string
	Cmd      = &cobra.Command{
		Use:   "analyze [file1] {file2 ... fileN}",
		Short: "Analyze audio files and extract metadata",
		Args:  cobra.ArbitraryArgs,
		RunE:  analyze,
	}
)

func init() {
	Cmd.Flags().StringVarP(&listFile, "listfile", "l", "", "Filename containing a list of files to analyze")
}

func analyze(cmd *cobra.Command, args []string) error {
	for _, f := range args {
		analyzeFile(f)
	}
	if listFile != "" {
		file, err := os.Open(listFile)
		if err != nil {
			return fmt.Errorf("error opening listfile %s: %v", listFile, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			analyzeFile(line)
		}
	}

	return nil
}

type Metadata struct {
	Filename string    `json:"filename"`
	Album    string    `json:"album"`
	Artist   string    `json:"artist"`
	Date     time.Time `json:"date"`
	Disc     int       `json:"disc"`
	Genre    []string  `json:"genre"`
	Set      int       `json:"set"`
	Title    string    `json:"title"`
	Track    int       `json:"track"`
	Venue    string    `json:"venue"`

	AccousticIdFingerprint string      `json:"accoustic_id_fingerprint"`
	MusicBrainz            MusicBrainz `json:"music_brainz"`

	Tags map[string]string `json:"tags"`
}

type MusicBrainz struct {
	// Unique ID of the artist or band. For example, Rush has the artist id of
	// "534ee493-bfac-4575-a44a-0ae41e2c3fe4".
	ArtistId string `json:"artist_id"`

	// Release group is what most people would call an "album". For example,
	// the album titled "Roll the Bones" by Rush has an release group id of
	// "e188de4e-6d15-3ca3-be49-fa13c67a03c0".
	ReleaseGroupId string `json:"release_group_id"`

	// Release is a specific edition of an album. For example, the album
	// titled "Roll the Bones" by Rush has had at least 13 different releases
	// around the world. The original US CD release by Atlantic on 1991-09-03
	// has the release id of "50e551bd-5d24-37e5-913d-07c25cd85e8e". Whereas
	// the original 12" vinyl release by Atlantic was worldwide and has the
	// release id of "52bf9926-dc7f-40b9-9a08-d5f0c98f8a63".
	ReleaseId string `json:"release_id"`
}

func newMetadata(filename string, md tag.Metadata) *Metadata {
	m := Metadata{
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

func analyzeFile(filename string) error {
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

	b, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal %s due to %s", filename, err)
	}
	fmt.Println(string(b))

	return nil
}
