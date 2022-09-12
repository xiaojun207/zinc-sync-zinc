package zinc

import (
	"context"
	"encoding/json"
	"github.com/xiaojun207/go-base-utils/array"
	"github.com/xiaojun207/go-base-utils/utils"
	zinc "github.com/zinclabs/sdk-go-zincsearch"
	"log"
	"strings"
	"time"
)

type Zinc struct {
	client   *zinc.APIClient
	user     string
	pass     string
	indexMap map[string]Index
}

func NewZinc(host string, user, pass string) (*Zinc, error) {
	url := host
	if strings.HasPrefix(host, "https://") || strings.HasPrefix(host, "http://") {
	} else {
		url = "http://" + host
	}

	configuration := zinc.NewConfiguration()
	configuration.Servers = zinc.ServerConfigurations{
		zinc.ServerConfiguration{
			URL: url,
		},
	}

	client := zinc.NewAPIClient(configuration)

	z := Zinc{
		client:   client,
		user:     user,
		pass:     pass,
		indexMap: map[string]Index{},
	}
	return &z, nil
}

func (z *Zinc) GetConfig() *zinc.Configuration {
	return z.client.GetConfig()
}

func (z *Zinc) ctx() context.Context {
	return context.WithValue(context.Background(), zinc.ContextBasicAuth, zinc.BasicAuth{
		UserName: z.user,
		Password: z.pass,
	})
}

func (z *Zinc) Version() (string, error) {
	resp, _, err := z.client.Default.Version(context.Background()).Execute()
	if err != nil {
		log.Println("Version.err:", err)
		return "", err
	}
	return resp.GetVersion(), nil
}

func (z *Zinc) IndexDocument(index string, document map[string]interface{}) (string, error) {
	ctx := z.ctx()
	resp, _, err := z.client.Document.Index(ctx, index).Document(document).Execute()
	if err != nil {
		return "", err
	}
	return resp.GetId(), nil
}

func (z *Zinc) IndexDocuments(index string, docs []map[string]interface{}) (string, error) {
	ctx := z.ctx()
	s := ""
	for _, doc := range docs {
		d, _ := json.Marshal(doc)
		s += string(d) + "\n"
	}
	resp, h, err := z.client.Document.Multi(ctx, index).Query(s).Execute()
	if err != nil {
		log.Println("IndexDocuments.http:", h)
		return "", err
	}
	return resp.GetMessage(), nil
}

func (z *Zinc) CreateIndex(index string, mapping zinc.MetaMappings, setting zinc.MetaIndexSettings) (string, error) {
	ctx := z.ctx()
	mappingData := map[string]interface{}{}
	d, _ := mapping.MarshalJSON()
	json.Unmarshal(d, &mappingData)
	data := zinc.MetaIndexSimple{
		Name:     &index,
		Mappings: mappingData,
		Settings: &setting,
	}
	resp, _, err := z.client.Index.Create(ctx).Data(data).Execute()
	if err != nil {
		return "", err
	}
	return resp.GetIndex(), nil
}

func (z *Zinc) IndexMapping(index string, mapping zinc.MetaMappings) (string, error) {
	ctx := z.ctx()
	resp, _, err := z.client.Index.SetMapping(ctx, index).Mapping(mapping).Execute()
	if err != nil {
		return "", err
	}
	return resp.GetMessage(), nil
}

func (z *Zinc) IndexSetting(index string, setting zinc.MetaIndexSettings) (string, error) {
	ctx := z.ctx()
	resp, _, err := z.client.Index.SetSettings(ctx, index).Settings(setting).Execute()
	if err != nil {
		return "", err
	}
	return resp.GetMessage(), nil
}

func (z *Zinc) Write(index string, hits *zinc.MetaHits) {
	var docs []map[string]interface{}
	for _, hit := range hits.Hits {
		hit.Source["_id"] = *hit.Id
		docs = append(docs, hit.Source)

	}
	msg, err := z.IndexDocuments(index, docs)
	if err != nil {
		log.Panicln("zinc write, msg:", msg, ", err:", err)
	}
}

func (z *Zinc) IndexList(indexMatch string, ignoreIndexArray []string) ([]Index, error) {
	ctx := z.ctx()
	resp, _, err := z.client.Index.List(ctx).Name(indexMatch).Execute()
	if err != nil {
		return []Index{}, err
	}
	d, _ := json.Marshal(resp.List)
	var res, indexArr []Index
	json.Unmarshal(d, &res)
	for _, m := range res {
		if len(ignoreIndexArray) > 0 {
			if array.Contains(ignoreIndexArray, m.Name) {
				continue
			}
		}
		indexArr = append(indexArr, m)
	}
	return res, err
}

func (z *Zinc) IndexMap(indexMatch string, ignoreIndexArray []string) map[string]Index {
	t := time.Now()
	indexList, err := z.IndexList(indexMatch, ignoreIndexArray)
	if err != nil {
		log.Panicln("indexList.err:", err)
	}
	z.indexMap = map[string]Index{}
	storageSize := 0
	for _, m := range indexList {
		z.indexMap[m.Name] = m
		storageSize += m.Stats.StorageSize
	}
	log.Println("IndexMap，耗时:", time.Since(t), ",storageSize:", FormatSize(float64(storageSize)), ", size:", len(indexList))
	return z.indexMap
}

func (z *Zinc) Search(index string, query zinc.MetaZincQuery) (*zinc.MetaHits, error) {
	ctx := z.ctx()

	resp, h, err := z.client.Search.Search(ctx, index).Query(query).Execute()
	if err != nil {
		log.Println("Search.h.Request:", h.Request)
		return nil, err
	}
	return resp.Hits, err
}

func (z *Zinc) SearchAll(index string, from, size int32) (*zinc.MetaHits, error) {
	query := zinc.NewMetaZincQuery()
	query.Query = zinc.NewMetaQuery()
	query.Query.MatchAll = map[string]interface{}{}
	query.From = &from
	query.Size = &size
	query.Sort = []string{"@timestamp", "_id"}
	// read from zinc
	hits, err := z.Search(index, *query)
	return hits, err
}

func FormatSize(s float64) string {
	b := s
	kb := s / (1024)
	mb := s / (1024 * 1024)
	gb := s / (1024 * 1024 * 1024)
	if gb > 0.1 {
		return utils.Float64ToStr(gb, 4) + "GB"
	} else if mb > 0.1 {
		return utils.Float64ToStr(mb, 4) + "MB"
	} else if kb > 0.1 {
		return utils.Float64ToStr(kb, 4) + "KB"
	} else {
		return utils.Float64ToStr(b, 4) + "b"
	}
}
