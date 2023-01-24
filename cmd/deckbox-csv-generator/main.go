package main

import (
	"log"
	"mtg-bulk-input/internal/data"
	"mtg-bulk-input/internal/ui"
	"os"
	"path"
	"strings"
)

const (
	port = 8080
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("USAGE: mtg-bulk-input [path/to/file.json]")
	}

	if !strings.HasSuffix(path.Base(os.Args[1]), ".json") {
		log.Fatalf("specified file must be a .json file")
	}

	err := data.SetupDirectories()
	if err != nil {
		log.Fatalf("failed to setup working directories: %v", err)
	}

	downloaded, err := data.DownloadBulkDataIfNewer()
	if err != nil {
		log.Fatalf("failed to download bulk files: %v", err)
	}

	if downloaded == 0 {
		log.Println("no new bulk files to download")
	} else {
		log.Printf("downloaded %d new bulk files", downloaded)
	}

	log.Println("building store...")

	store, err := data.BuildDataStore()
	if err != nil {
		log.Fatalf("failed to build data store: %v", err)
	}

	err = ui.Start(os.Args[1], store)
	if err != nil {
		log.Fatalf("ui errored: %v", err)
	}
}
