package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

const (
	WorkingDirectory = "./data"

	// Where the bulk data files, and metadata are stored
	BulkDataDirectory = "bulk"
)

func SetupDirectories() error {
	directoriesToSetup := []string{
		path.Join(WorkingDirectory, BulkDataDirectory),
	}

	for _, d := range directoriesToSetup {
		err := os.MkdirAll(d, 0750)
		if err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", d, err)
		}
	}

	return nil
}

func ReadJsonFile(dir string, file string, out any) error {
	p := path.Join(WorkingDirectory, dir, file)

	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %w", p, err)
	}

	dec := json.NewDecoder(f)

	err = dec.Decode(out)
	if err != nil {
		return fmt.Errorf("failed to decode json file '%s': %w", p, err)
	}

	return nil
}

func WriteJsonFile(dir string, file string, out any) error {
	p := path.Join(WorkingDirectory, dir, file)

	f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("unable to open file at '%s': %w", p, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)

	err = enc.Encode(out)
	if err != nil {
		return fmt.Errorf("unable to write json file '%s': %w", p, err)
	}

	return nil
}
