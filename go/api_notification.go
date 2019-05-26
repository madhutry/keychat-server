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
	MessageText string `json:"message"`
	Token       string `json:"token"`
}
type ReceivedMesg struct {
	MessageText string `json:"message"`
	Sender      string `json:"sender"`
	Timestamp   string `json:"timestamp"`
	RoomId      string
}

var broadcast = make(chan Message)
var receiver = make(chan map[string]interface{})
var clients = make(map[string]map[string]*websocket.Conn)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func UpgradeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	cookie, _ := r.Cookie("X-Authorization")
	if cookie == nil {
		fmt.Print("Cookie returned null Auth Token")
		return
	}
	reqToken := cookie.Value
	token := retrieveToken(reqToken)
	friezeAccessCode := token["FriezeAccessCode"].(string)
	domainName := token["DomainName"].(string)
	populateWSMap(friezeAccessCode, domainName, conn)
	if err != nil {
		log.Fatal(err)
	}

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}
		broadcast <- msg
	}
}
func populateWSMap(friezeAccessCode string, domainName string, conn *websocket.Conn) {
	roomId, _, _, userId := dbGetAllDetails(friezeAccessCode, domainName)
	if clients[roomId] == nil {
		clients[roomId] = make(map[string]*websocket.Conn)
	}
	clients[roomId][userId] = conn
}
func HandleMessages() {
	for {
		// grab next message from the broadcast channel
		msg := <-broadcast
		mesg := msg.MessageText
		auth := msg.Token
		token := retrieveToken(fmt.Sprintf("Bearer %s", auth))
		friezeAccessCode := token["FriezeAccessCode"].(string)
		domainName := token["DomainName"].(string)
		sendMessage(friezeAccessCode, domainName, mesg)
	}
}
func HandleReceiveMessages() {
	for {
		m := <-receiver
		//batchId := m["batchId"].(string)
		messagesRecvd := m["messages"].(map[string]interface{})
		for k, mesg := range messagesRecvd {
			roomID := k
			var messages [][]string
			mesgArr := mesg.([]interface{})
			for _, val := range mesgArr {
				v := val.(map[string]interface{})
				mesgStr := v["message"].(string)
				ts := v["timestamp"].(string)
				sender := v["sender"].(string)
				mesg1 := []string{mesgStr, ts, sender}
				messages = append(messages, mesg1)
			}
			for k, v := range clients[roomID] {
				result := map[string]interface{}{
					"messages": messages,
					"userId":   k,
				}
				v.WriteJSON(result)
			}
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
