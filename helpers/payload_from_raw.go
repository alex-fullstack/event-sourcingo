package helpers

import "encoding/json"

func PayloadFromRaw(rawPayload interface{}) (map[string]interface{}, error) {
	var jsonData []byte
	jsonData, err := json.Marshal(rawPayload)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	err = json.Unmarshal(jsonData, &payload)
	return payload, err
}
