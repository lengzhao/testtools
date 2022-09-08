package grpct

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/nsf/jsondiff"
)

type Case struct {
	Name     string      `json:"name"`
	Service  string      `json:"service"`
	Method   string      `json:"method"`
	Headers  []string    `json:"headers"`
	Error    string      `json:"error"`
	Request  interface{} `json:"request"`
	Response interface{} `json:"response"`
}

func (c Case) GetServeceName() string {
	return c.Service
}

func (c Case) GetRequest() []byte {
	if c.Request == nil {
		return nil
	}
	data, _ := json.Marshal(c.Request)
	return data
}

func (c Case) GetResponse() []byte {
	if c.Response == nil {
		return nil
	}
	data, _ := json.Marshal(c.Response)
	return data
}

func (c Case) CompareResponse(data []byte, err error) error {
	if err != nil {
		if c.Error == "" {
			log.Printf("name:%s,hope success,get error:%s\n", c.Name, err)
			return fmt.Errorf("hope success,get error:%s", err)
		}
		if c.Error == "*" {
			return nil
		}
		if c.Error != err.Error() {
			log.Printf("name:%s,hope error:%s,get error:%s\n", c.Name, c.Error, err)
			return fmt.Errorf("different error")
		}
		return nil
	} else if c.Error != "" {
		log.Printf("name:%s,hope error:%s,get response:%s\n", c.Name, c.Error, string(data))
		return fmt.Errorf("hope error")
	}
	hope := c.GetResponse()
	if len(hope) == 0 && len(data) == 0 {
		return nil
	}
	if len(hope) == 3 && string(hope) == "\"*\"" && len(data) > 0 {
		return nil
	}
	ops := jsondiff.DefaultJSONOptions()
	diff, _ := jsondiff.Compare(hope, data, &ops)

	if diff == jsondiff.FullMatch {
		return nil
	}
	log.Printf("different response,hope:%s, get:%s\n", hope, data)
	return fmt.Errorf("different:%s", diff)
}

func LoadTestcase(filename string) (*Case, error) {
	var testcase Case
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("fail to read file:", filename, err)
		return nil, err
	}
	err = json.Unmarshal(data, &testcase)
	if err != nil {
		log.Println("fail to Unmarshal:", filename, err)
		return nil, err
	}
	return &testcase, nil
}

func LoadTestcases(dir string) ([]Case, error) {
	var out []Case
	err := filepath.Walk(dir,
		func(fn string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			// log.Println("fn:", fn)
			if path.Ext(fn) != ".json" {
				log.Println("ignore file(not .json):", fn)
				return nil
			}
			var testcase Case
			data, err := ioutil.ReadFile(fn)
			if err != nil {
				log.Println("fail to read file:", fn, err)
				return nil
			}
			err = json.Unmarshal(data, &testcase)
			if err != nil {
				log.Println("fail to Unmarshal:", fn, err)
				return nil
			}
			svcName := testcase.GetServeceName()
			if svcName == "" {
				log.Println("unknow service name:", fn)
				return nil
			}
			out = append(out, testcase)
			return nil
		})
	if err != nil {
		log.Println("Walk", dir, err)
		return nil, err
	}
	return out, nil
}
