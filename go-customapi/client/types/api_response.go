package types

import (
	"encoding/json"
	"fmt"
)

type APIResponse struct {
	Data json.RawMessage `json:"data"`
	Status string `json:"status"`
	StatusMessages []string `json:"status_messages"`
	Message string `json:"message"`
	RequestId string `json:"requestId"`
	Requester string `json:"requester"`
	OperationId string `json:"operationId"`
	Api string `json:"api"`
}

func (r *APIResponse) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Data           json.RawMessage `json:"data"`
		Status         string          `json:"status"`
		StatusMessages []string        `json:"status_messages"`
		Message        interface{}     `json:"message"`
		RequestId      string          `json:"requestId"`
		Requester      string          `json:"requester"`
		OperationId    string          `json:"operationId"`
		Api            string          `json:"api"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	messageString := ""
	switch msg := tmp.Message.(type) {
	case string:
		messageString = msg
	case map[string]interface{}:
		if id, ok := msg["id"]; ok {
			if idStr, ok := id.(string); ok {
				messageString = idStr
			}
		}
	}

	*r = APIResponse{
		Data:           tmp.Data,
		Status:         tmp.Status,
		StatusMessages: tmp.StatusMessages,
		Message:        messageString,
		RequestId:      tmp.RequestId,
		Requester:      tmp.Requester,
		OperationId:    tmp.OperationId,
		Api:            tmp.Api,
	}
	return nil
}