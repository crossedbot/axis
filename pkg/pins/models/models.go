package models

import (
	"fmt"
	"sort"
	"strings"
)

type KeyList []string

func (k KeyList) Len() int      { return len(k) }
func (k KeyList) Swap(i, j int) { k[i], k[j] = k[j], k[i] }
func (k KeyList) Less(i, j int) bool {
	cmp := strings.Compare(k[i], k[j])
	return cmp < 0
}

type Info map[string]string

func (i Info) Keys() []string {
	keys := KeyList{}
	for k, _ := range i {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	return keys
}

func (i Info) String() string {
	var ss []string
	keys := i.Keys()
	for _, k := range keys {
		ss = append(ss, fmt.Sprintf("%s:%s", k, i[k]))
	}
	return strings.Join(ss, ",")
}

func InfoFromString(s string) Info {
	info := make(Info)
	pairs := strings.Split(s, ",")
	for _, kv := range pairs {
		parts := strings.Split(kv, ":")
		if len(parts) == 2 {
			info[parts[0]] = parts[1]
		}
	}
	return info
}

type Pin struct {
	Cid     string   `json:"cid"`
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Meta    Info     `json:"meta"`
}

type PinStatus struct {
	Id        string   `json:"requestid"`
	Status    string   `json:"status"`
	Created   string   `json:"created"`
	Pin       Pin      `json:"pin"`
	Delegates []string `json:"delegates"`
	Info      Info     `json:"info"`
}

type Pins struct {
	Count   int         `json:"count"`
	Total   int         `json:"total"`
	Results []PinStatus `json:"results"`
}
