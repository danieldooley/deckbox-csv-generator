package data

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"mtg-bulk-input/internal/scryfall"
	"os"
	"path"
)

var requiredBulkFileTypes = []string{
	"default_cards",
}

/*
Download the latest card data from the scryfall bulk data API
*/
func DownloadBulkDataIfNewer() (int, error) {
	// Get our stored metadata for comparison
	var meta scryfall.BulkDataList
	metadataExists := true

	err := ReadJsonFile(BulkDataDirectory, "meta.json", &meta)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			metadataExists = false
		} else {
			return 0, fmt.Errorf("unable to read bulk data metadata: %w", err)
		}
	}

	// Get the upstream metadata
	newMeta, err := scryfall.GetBulkDataMeta()
	if err != nil {
		return 0, fmt.Errorf("failed to get new bulk metadata: %w", err)
	}

	// Track the files to be downloaded
	toDownload := make([]string, 0)

	// Compare the two for each requested type to populate toDownload list
	for _, t := range requiredBulkFileTypes {
		if metadataExists {
			bdType, ok := meta.GetType(t)
			if !ok {
				return 0, fmt.Errorf("existing metadata does not contain type '%s'", t)
			}

			newBdType, ok := newMeta.GetType(t)
			if !ok {
				return 0, fmt.Errorf("new metadata does not contain type '%s'", t)
			}

			if newBdType.UpdatedAt.After(bdType.UpdatedAt) {
				log.Printf("bulk data type '%s' has been updated, will download", t)
				toDownload = append(toDownload, t)
			}
		} else {
			toDownload = append(toDownload, t)
		}
	}

	// Download each bulk file
	for i, t := range toDownload {
		bdType, ok := newMeta.GetType(t)
		if !ok {
			return i, fmt.Errorf("failed to get type '%s' from new metadata for downloading", t)
		}

		log.Printf("downloading bulk data type '%s' from '%s'", t, bdType.DownloadUri)

		p := path.Join(WorkingDirectory, BulkDataDirectory, fmt.Sprintf("%s.json", t))

		f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			return i, fmt.Errorf("unable to open file at '%s': %w", p, err)
		}

		err = scryfall.DownloadBulkFile(bdType.DownloadUri, f)
		f.Close()
		if err != nil {
			return i, fmt.Errorf("failed to download bulk file '%s': %w", t, err)
		}
	}

	// Update the meta file on disk
	err = WriteJsonFile(BulkDataDirectory, "meta.json", newMeta)

	return len(toDownload), nil
}
