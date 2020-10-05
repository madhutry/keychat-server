package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func FcmNotify(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	//var out1 bytes.Buffer
	//json.Indent(&out1, body, "=", "\t")
	//out1.WriteTo(os.Stdout)
	var f map[string]interface{}
	json.Unmarshal([]byte(body), &f)
	device := f["notification"].(map[string]interface{})["devices"].([]interface{})[0]
	pushkey := device.(map[string]interface{})["pushkey"].(string)
	eventId := f["notification"].(map[string]interface{})["event_id"].(string)
	roomId := f["notification"].(map[string]interface{})["room_id"].(string)
	apiSendNotification(pushkey, eventId, roomId)
	tokenJson := map[string]string{}
	enc := json.NewEncoder(w)
	enc.Encode(&tokenJson)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
func apiSendNotification(pushkey string, eventId string, roomId string) interface{} {
	jsonData := map[string]interface{}{
		"priority": "high",
		"to":       pushkey,
		"data": map[string]interface{}{
			"prio":     "high",
			"event_id": eventId,
			"room_id":  roomId,
			"unread":   1,
		},
	}
	endpoint := "https://fcm.googleapis.com/fcm/send"
	client := &http.Client{}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonValue))
	request.Header.Add("Content-Type", "application/json")
	key := fmt.Sprintf("key=%s", GetFCMServerCode())
	request.Header.Add("Authorization", key)
	response, err := client.Do(request)
	fmt.Print(response.StatusCode)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return nil
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)
		var out1 bytes.Buffer
		json.Indent(&out1, data, "=", "\t")
		out1.WriteTo(os.Stdout)
		return f
	}
}
