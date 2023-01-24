package scryfall

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	baseScryfallApiPath = "https://api.scryfall.com/"
)

var (
	httpClient http.Client
)

func init() {
	httpClient = http.Client{}
}

func handleRespStatus(resp *http.Response) error {
	dec := json.NewDecoder(resp.Body)

	if resp.StatusCode > 400 {
		var returnedError Error

		err := dec.Decode(&returnedError)
		if err != nil {
			return fmt.Errorf("non 200 status code (%d) with unparseable error body: %w", resp.StatusCode, err)
		}

		return returnedError
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non 200, non error status code (%d)", resp.StatusCode)
	}

	return nil
}

func getJson(path string, out any) error {
	url, err := url.JoinPath(baseScryfallApiPath, path)
	if err != nil {
		return fmt.Errorf("failed to combine path: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to construct request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err = handleRespStatus(resp); err != nil {
		return err
	}

	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(out)
	if err != nil {
		return fmt.Errorf("unable to parse response into output: %w", err)
	}

	return nil
}

func GetBulkDataMeta() (BulkDataList, error) {
	var out BulkDataList

	err := getJson("/bulk-data", &out)
	if err != nil {
		return BulkDataList{}, fmt.Errorf("failed to get from bulk-data endpoint: %w", err)
	}

	return out, nil
}

func DownloadBulkFile(url string, w io.Writer) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to construct request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err = handleRespStatus(resp); err != nil {
		return err
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed while downloading request body: %w", err)
	}

	return nil
}