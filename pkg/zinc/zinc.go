package zinc

import (
	"context"
	"encoding/json"
	zinc "github.com/zinclabs/sdk-go-zincsearch"
	"log"
	"strings"
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
	for _, hit := range hits.Hits {
		// write to Zinc
		hit.Source["_id"] = *hit.Id
		id, err := z.IndexDocument(index, hit.Source)
		if err != nil {
			log.Panicln("zinc write, id:", id, ", err:", err)
		}
	}
}

func (z *Zinc) IndexList() ([]Index, error) {
	ctx := z.ctx()

	resp, _, err := z.client.Index.List(ctx).Execute()
	if err != nil {
		return []Index{}, err
	}
	d, _ := json.Marshal(resp.List)
	var res []Index
	json.Unmarshal(d, &res)
	return res, err
}

func (z *Zinc) IndexMap() map[string]Index {
	indexList, err := z.IndexList()
	if err != nil {
		log.Panicln("indexList.err:", err)
	}
	z.indexMap = map[string]Index{}
	for _, m := range indexList {
		z.indexMap[m.Name] = m
	}
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