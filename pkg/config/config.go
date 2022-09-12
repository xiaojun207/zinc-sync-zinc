/* Copyright 2022 Zinc Labs Inc. and Contributors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Config struct {
	PrimaryZincHost       string   `env:"PRIMARY_ZINC_HOST,default=localhost:4080"`
	PrimaryZincUser       string   `env:"PRIMARY_ZINC_USER,default=admin"`
	PrimaryZincPassword   string   `env:"PRIMARY_ZINC_PASSWORD,default=Complexpass#123"`
	SecondaryZincHost     string   `env:"SECONDARY_ZINC_HOST,default=localhost:4080"`
	SecondaryZincUser     string   `env:"SECONDARY_ZINC_USER,default=admin"`
	SecondaryZincPassword string   `env:"SECONDARY_ZINC_PASSWORD,default=Complexpass#123"`
	IndexMatch            string   `env:"INDEX_MATCH,default="`
	IgnoreIndexList       []string `env:"IGNORE_INDEX_LIST,default="`
	GoroutineLimit        int      `env:"GOROUTINE_LIMIT,default=1000"`
	Debug                 bool     `env:"DEBUG,default=false"`
	PageSize              int32    `env:"PAGE_SIZE,default=100"`
}

func InitConfig(filenames ...string) *Config {
	config := new(Config)
	err := godotenv.Load(filenames...)
	if err != nil {
		log.Print(err.Error())
	}
	rv := reflect.ValueOf(config).Elem()
	loadConfig(rv)
	return config
}

func loadConfig(rv reflect.Value) {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Field(i)
		ft := rt.Field(i)
		if ft.Type.Kind() == reflect.Struct {
			loadConfig(fv)
			continue
		}
		if ft.Tag.Get("env") != "" {
			tag := ft.Tag.Get("env")
			setField(fv, tag)
		}
	}
}

func setField(field reflect.Value, tag string) {
	if tag == "" {
		return
	}
	tagColumn := strings.Split(tag, ",")
	v := os.Getenv(tagColumn[0])
	if v == "" {
		if len(tagColumn) > 1 {
			tv := strings.Join(tagColumn[1:], ",")
			if strings.HasPrefix(tv, "default=") {
				v = tv[8:]
			}
		}
	}
	if v == "" {
		return
	}
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vi, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Printf("env %s is not int", tag)
		}
		field.SetInt(int64(vi))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vi, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			log.Printf("env %s is not uint", tag)
		}
		field.SetUint(uint64(vi))
	case reflect.Bool:
		vi, err := strconv.ParseBool(v)
		if err != nil {
			log.Printf("env %s is not bool", tag)
		}
		field.SetBool(vi)
	case reflect.String:
		field.SetString(v)
	case reflect.Slice:
		vs := strings.Split(v, ",")
		field.Set(reflect.ValueOf(vs))
		field.SetLen(len(vs))
	}
}
