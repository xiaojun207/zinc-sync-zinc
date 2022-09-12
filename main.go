package main

import (
	"fmt"
	"github.com/xiaojun207/zinc-sync-zinc/pkg/config"
	"github.com/xiaojun207/zinc-sync-zinc/pkg/pool"
	"github.com/xiaojun207/zinc-sync-zinc/pkg/zinc"
	"log"
	"time"
)

func main() {
	Config := config.InitConfig()
	log.Println("Config.PrimaryZincHost:", Config.PrimaryZincHost)
	log.Println("Config.SecondaryZincHost:", Config.SecondaryZincHost)
	log.Println("Config.IndexMatch:", Config.IndexMatch)
	log.Println("Config.IgnoreIndexList:", Config.IgnoreIndexList)
	log.Println("Config.GoroutineLimit:", Config.GoroutineLimit)
	log.Println("Config.PageSize:", Config.PageSize)

	// init primaryZinc
	primaryZinc, err := zinc.NewZinc(Config.PrimaryZincHost, Config.PrimaryZincUser, Config.PrimaryZincPassword)
	if err != nil {
		log.Fatal(err)
	}
	primaryZinc.GetConfig().Debug = Config.Debug

	// init secondaryZinc
	secondaryZinc, err := zinc.NewZinc(Config.SecondaryZincHost, Config.SecondaryZincUser, Config.SecondaryZincPassword)
	if err != nil {
		log.Fatal(err)
	}
	secondaryZinc.GetConfig().Debug = Config.Debug

	size := Config.PageSize
	pool := pool.NewPool(Config.GoroutineLimit)
	defer pool.Release()
	for {
		fmt.Println("--------------------------------------------------------------------------------------------------------------------------------------------")

		indexMap := SyncIndexMap(primaryZinc, secondaryZinc, Config.IndexMatch, Config.IgnoreIndexList)
		log.Println("indexMap.len:", len(indexMap))
		c := 0
		t := time.Now()
		for name, idx := range indexMap {
			index := idx
			if !index.Synced() {
				f := func() {
					log.Println("index.sync.start:", index.Name, ", \tFrom:", index.From, ", \tDocNum:", index.Stats.DocNum, ", \tDocTimeMax:", index.Stats.DocTimeMax)
					index.From = SyncDoc(primaryZinc, secondaryZinc, index.Name, index.Name, index.From, size)
					indexMap[name] = index
					c++
				}
				pool.Submit(f)
			}
		}
		pool.Wait()

		log.Printf("indexMap synced: %d,耗时(s)：%f", c, time.Since(t).Seconds())

		time.Sleep(time.Second * 30)
	}
}

func SyncIndexMap(primaryZinc, secondaryZinc *zinc.Zinc, indexMatch string, ignoreIndexList []string) map[string]zinc.Index {
	primaryIndexMap := primaryZinc.IndexMap(indexMatch, ignoreIndexList)
	secondaryIndexMap := secondaryZinc.IndexMap(indexMatch, ignoreIndexList)
	log.Println("primaryIndexMap:", len(primaryIndexMap), ", secondaryIndexMap:", len(secondaryIndexMap))
	for _, m := range primaryIndexMap {
		if secondaryIndex, ok := secondaryIndexMap[m.Name]; ok {
			m.From = int32(secondaryIndex.Stats.DocNum)
			if m.Stats.DocNum > secondaryIndex.Stats.DocNum {
				//log.Printf("name: %s, m.DocNum: %d, secondaryIndex.DocNum: %d", m.Name, m.Stats.DocNum, secondaryIndex.Stats.DocNum)
			}
		} else {
			m.From = 0
			secondaryZinc.CreateIndex(m.Name, m.Mappings, m.Settings)
		}
		primaryIndexMap[m.Name] = m
	}
	return primaryIndexMap
}

func SyncDoc(primaryZinc, secondaryZinc *zinc.Zinc, primaryIndexName, secondaryIndexName string, from, size int32) int32 {
	for {
		hits, err := primaryZinc.SearchAll(primaryIndexName, from, size)
		if err != nil {
			log.Println("SyncDoc.err:", err, ", primary:", primaryIndexName, ", from:", from, ", size:", size)
			return from
		}
		total := *hits.Total.Value
		count := len(hits.Hits)

		if count > 0 {
			log.Printf("SyncDoc, index: %s, \tfrom/total: %d/%d, \tcount:%d \n", primaryIndexName, from, total, count)
			secondaryZinc.Write(secondaryIndexName, hits)
			from = from + int32(count)

		}

		if count == 0 || from >= total {
			log.Printf("SyncDoc, index: %s, from/total: %d/%d, count:%d \n", primaryIndexName, from, total, count)
			return from
		}
		time.Sleep(time.Millisecond)
	}
}
