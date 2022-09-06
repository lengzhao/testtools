package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/nsf/jsondiff"
)

type Case struct {
	Name     string      `json:"name,omitempty"`
	Method   string      `json:"method,omitempty"`
	Headers  []string    `json:"headers,omitempty"`
	Error    string      `json:"error,omitempty"`
	Request  interface{} `json:"request,omitempty"`
	Response interface{} `json:"response,omitempty"`
}

func (c Case) GetServeceName() string {
	if c.Method == "" {
		return ""
	}
	arr := strings.Split(c.Method, "/")
	if len(arr) > 1 {
		return arr[0]
	}
	return ""
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

func (c Case) CompareResponse(data []byte, err error) bool {
	if err != nil {
		if c.Error == "" {
			log.Printf("name:%s,hope success,get error:%s\n", c.Name, err)
			return false
		}
		if c.Error == "*" {
			return true
		}
		if c.Error != err.Error() {
			log.Printf("name:%s,hope error:%s,get error:%s\n", c.Name, c.Error, err)
			return false
		}
		return true
	} else if c.Error != "" {
		log.Printf("name:%s,hope error:%s,get response:%s\n", c.Name, c.Error, string(data))
		return false
	}
	hope := c.GetResponse()
	if len(hope) == 0 && len(data) == 0 {
		return true
	}
	if len(hope) == 3 && string(hope) == "\"*\"" && len(data) > 0 {
		return true
	}
	ops := jsondiff.DefaultJSONOptions()
	diff, _ := jsondiff.Compare(hope, data, &ops)
	return diff == jsondiff.FullMatch
}
