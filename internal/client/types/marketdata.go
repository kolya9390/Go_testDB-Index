package types

// curl "https://garantex.org/api/v2/depth?market={market}"
//   "market": "usdtrub",

type DataMD struct {
	Timestamp int64 `json:"timestamp"`
	Asks      []Order `json:"asks"`
	Bids      []Order `json:"bids"`
}

type Order struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
	Factor string `json:"factor"`
	Type   string `json:"type"`
}