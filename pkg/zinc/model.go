package zinc

import client "github.com/zinclabs/sdk-go-zincsearch"

type Page struct {
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

type Stats struct {
	DocTimeMin  int `json:"doc_time_min"`
	DocTimeMax  int `json:"doc_time_max"`
	DocNum      int `json:"doc_num"`
	StorageSize int `json:"storage_size"`
	WalSize     int `json:"wal_size"`
}

type Index struct {
	Name        string                   `json:"name"`
	StorageType string                   `json:"storage_type"`
	Settings    client.MetaIndexSettings `json:"settings"`
	Mappings    client.MetaMappings      `json:"mappings"`
	ShardNum    int                      `json:"shard_num"`
	Shards      map[string]interface{}   `json:"shards"`
	Stats       Stats                    `json:"stats"`
	Version     string                   `json:"version"`
	From        int32                    `json:"from"`
}

func (i *Index) Synced() bool {
	return i.Stats.DocNum == int(i.From)
}
