package scryfall

import (
	"fmt"
	"time"
)

/*
	Error
*/

type Error struct {
	Code     string   `json:"code"`
	Status   int      `json:"status"`
	Warnings []string `json:"warnings"`
	Details  string   `json:"details"`
}

func (e Error) Error() string {
	return fmt.Sprintf("scryfall request failed with status '%s' (HTTP %d): %s", e.Code, e.Status, e.Details)
}

/*
	BulkData
*/

type BulkData struct {
	Object          string    `json:"object"`
	Id              string    `json:"id"`
	Type            string    `json:"type"`
	UpdatedAt       time.Time `json:"updated_at"`
	Uri             string    `json:"uri"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Size            int       `json:"size"`
	DownloadUri     string    `json:"download_uri"`
	ContentType     string    `json:"content_type"`
	ContentEncoding string    `json:"content_encoding"`
}

type BulkDataList struct {
	HasMore bool       `json:"has_more"`
	Data    []BulkData `json:"data"`
}

func (bdl BulkDataList) GetType(t string) (BulkData, bool) {
	for _, bd := range bdl.Data {
		if bd.Type == t {
			return bd, true
		}
	}

	return BulkData{}, false
}

/*
Cards
*/
type Card struct {
	Object        string `json:"object"`
	Id            string `json:"id"`
	OracleId      string `json:"oracle_id"`
	MultiverseIds []int  `json:"multiverse_ids"`
	MtgoId        int    `json:"mtgo_id"`
	MtgoFoilId    int    `json:"mtgo_foil_id"`
	TcgplayerId   int    `json:"tcgplayer_id"`
	CardmarketId  int    `json:"cardmarket_id"`
	Name          string `json:"name"`
	Lang          string `json:"lang"`
	ReleasedAt    string `json:"released_at"`
	Uri           string `json:"uri"`
	ScryfallUri   string `json:"scryfall_uri"`
	Layout        string `json:"layout"`
	HighresImage  bool   `json:"highres_image"`
	ImageStatus   string `json:"image_status"`
	ImageUris     struct {
		Small      string `json:"small"`
		Normal     string `json:"normal"`
		Large      string `json:"large"`
		Png        string `json:"png"`
		ArtCrop    string `json:"art_crop"`
		BorderCrop string `json:"border_crop"`
	} `json:"image_uris"`
	ManaCost      string        `json:"mana_cost"`
	Cmc           float64       `json:"cmc"`
	TypeLine      string        `json:"type_line"`
	OracleText    string        `json:"oracle_text"`
	Power         string        `json:"power"`
	Toughness     string        `json:"toughness"`
	Colors        []string      `json:"colors"`
	ColorIdentity []string      `json:"color_identity"`
	Keywords      []interface{} `json:"keywords"`
	Legalities    struct {
		Standard        string `json:"standard"`
		Future          string `json:"future"`
		Historic        string `json:"historic"`
		Gladiator       string `json:"gladiator"`
		Pioneer         string `json:"pioneer"`
		Explorer        string `json:"explorer"`
		Modern          string `json:"modern"`
		Legacy          string `json:"legacy"`
		Pauper          string `json:"pauper"`
		Vintage         string `json:"vintage"`
		Penny           string `json:"penny"`
		Commander       string `json:"commander"`
		Brawl           string `json:"brawl"`
		Historicbrawl   string `json:"historicbrawl"`
		Alchemy         string `json:"alchemy"`
		Paupercommander string `json:"paupercommander"`
		Duel            string `json:"duel"`
		Oldschool       string `json:"oldschool"`
		Premodern       string `json:"premodern"`
	} `json:"legalities"`
	Games           []string `json:"games"`
	Reserved        bool     `json:"reserved"`
	Foil            bool     `json:"foil"`
	Nonfoil         bool     `json:"nonfoil"`
	Finishes        []string `json:"finishes"`
	Oversized       bool     `json:"oversized"`
	Promo           bool     `json:"promo"`
	Reprint         bool     `json:"reprint"`
	Variation       bool     `json:"variation"`
	SetId           string   `json:"set_id"`
	Set             string   `json:"set"`
	SetName         string   `json:"set_name"`
	SetType         string   `json:"set_type"`
	SetUri          string   `json:"set_uri"`
	SetSearchUri    string   `json:"set_search_uri"`
	ScryfallSetUri  string   `json:"scryfall_set_uri"`
	RulingsUri      string   `json:"rulings_uri"`
	PrintsSearchUri string   `json:"prints_search_uri"`
	CollectorNumber string   `json:"collector_number"`
	Digital         bool     `json:"digital"`
	Rarity          string   `json:"rarity"`
	FlavorText      string   `json:"flavor_text"`
	CardBackId      string   `json:"card_back_id"`
	Artist          string   `json:"artist"`
	ArtistIds       []string `json:"artist_ids"`
	IllustrationId  string   `json:"illustration_id"`
	BorderColor     string   `json:"border_color"`
	Frame           string   `json:"frame"`
	FullArt         bool     `json:"full_art"`
	Textless        bool     `json:"textless"`
	Booster         bool     `json:"booster"`
	StorySpotlight  bool     `json:"story_spotlight"`
	EdhrecRank      int      `json:"edhrec_rank"`
	PennyRank       int      `json:"penny_rank"`
	Prices          struct {
		Usd       string      `json:"usd"`
		UsdFoil   string      `json:"usd_foil"`
		UsdEtched interface{} `json:"usd_etched"`
		Eur       string      `json:"eur"`
		EurFoil   string      `json:"eur_foil"`
		Tix       string      `json:"tix"`
	} `json:"prices"`
	RelatedUris struct {
		Gatherer                  string `json:"gatherer"`
		TcgplayerInfiniteArticles string `json:"tcgplayer_infinite_articles"`
		TcgplayerInfiniteDecks    string `json:"tcgplayer_infinite_decks"`
		Edhrec                    string `json:"edhrec"`
	} `json:"related_uris"`
}
