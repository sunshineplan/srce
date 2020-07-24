package misc

import (
	"encoding/json"
	"sync"

	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/utils/mail"
)

// MetadataConfig is metadata server config
var MetadataConfig metadata.Config

// GetUsers get auth user info
func GetUsers() (map[string]interface{}, error) {
	m, err := metadata.Get("srce_user", &MetadataConfig)
	if err != nil {
		return nil, err
	}
	var users interface{}
	if err := json.Unmarshal(m, &users); err != nil {
		return nil, err
	}
	return users.(map[string]interface{}), nil
}

// GetConfig get server setting
func GetConfig() (map[string]interface{}, string, mail.Setting, error) {
	var mailSetting mail.Setting
	var err error
	var config = map[string]interface{}{"srce_command": nil, "srce_path": nil, "srce_subscribe": nil}
	var wg sync.WaitGroup
	wg.Add(3)
	for k := range config {
		go func(k string) {
			defer wg.Done()
			var b []byte
			var v interface{}
			b, err = metadata.Get(k, &MetadataConfig)
			if err != nil {
				return
			}
			err = json.Unmarshal(b, &v)
			config[k] = v
		}(k)
	}
	wg.Wait()
	if err != nil {
		return nil, "", mailSetting, err
	}
	jsonbody, _ := json.Marshal(config["srce_subscribe"])
	json.Unmarshal(jsonbody, &mailSetting)
	return config["srce_command"].(map[string]interface{}), config["srce_path"].(string), mailSetting, nil
}
