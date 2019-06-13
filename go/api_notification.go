package friezechat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	MesgType    string `json:"mesgType"`
	MessageText string `json:"message"`
	Token       string `json:"token"`
	Sock        *websocket.Conn
	Uuid        string `json:"uuid"`
}
type ReceivedMesg struct {
	MessageText string `json:"message"`
	Sender      string `json:"sender"`
	Timestamp   string `json:"timestamp"`
	RoomId      string
}

var broadcast = make(chan Message)

var receiver = make(chan map[string]interface{})

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func UpgradeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		msg.Sock = conn
		if err != nil {
			log.Println(err)
			return
		}
		broadcast <- msg
	}
}

func HandleMessages() {
	for {
		// grab next message from the broadcast channel
		msg := <-broadcast
		mesgType := msg.MesgType
		mesg := msg.MessageText
		auth := msg.Token
		sock := msg.Sock
		uuid := msg.Uuid
		token := retrieveToken(fmt.Sprintf("Bearer %s", auth))
		friezeAccessCode := token["FriezeAccessCode"].(string)
		domainName := token["DomainName"].(string)
		userId := token["UserId"].(string)

		if mesgType == "sendmesg" {
			sendMessage(friezeAccessCode, domainName, mesg, uuid)
		} else if mesgType == "ping" {
			result := make(map[string]interface{})
			result["messages"], result["lastSerialId"] = dbGetMessages(friezeAccessCode)
			result["userId"] = userId
			sock.WriteJSON(result)
		} else if mesgType == "checkagentonline" {
			result := checkOwnerOnline(domainName)
			sock.WriteJSON(result)
		}
	}
}

func ReceiveNotification(w http.ResponseWriter, req *http.Request) {
	data, _ := ioutil.ReadAll(req.Body)
	var f interface{}
	json.Unmarshal([]byte(data), &f)

	m := f.(map[string]interface{})
	receiver <- m
}
