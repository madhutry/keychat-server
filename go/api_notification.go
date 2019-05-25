package friezechat

import (
	"fmt"
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
var receiver = make(chan ReceivedMesg)
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

/* func HandleReceiveMessages() {
	for {
		// grab next message from the broadcast channel
		msg := <-receiver
		for _, v := range clients[msg.RoomId] {
			v.WriteJSON(msg)
		}
	}
}
func ReceiveNotification(w http.ResponseWriter, req *http.Request) {
	data, _ := ioutil.ReadAll(req.Body)
	var f interface{}
	json.Unmarshal([]byte(data), &f)

	m := f.(map[string]interface{})
	eventId := m["notification"].(map[string]interface{})["event_id"].(string)
	roomId := m["notification"].(map[string]interface{})["room_id"].(string)
	matAccessCode := dbGetNotifcationDetails(roomId)
	mesgDetails := apiGetEventIdDetails(eventId, roomId, matAccessCode)
	recevdMesg := ReceivedMesg{
		MessageText: mesgDetails["message"].(string),
		Sender:      mesgDetails["sender"].(string),
		Timestamp:   mesgDetails["timestamp"].(string),
		RoomId:      mesgDetails["roomId"].(string),
	}
	receiver <- recevdMesg
}

func apiGetEventIdDetails(eventId string, roomId string, accessCode string) map[string]interface{} {
	apiHost := "http://%s/_matrix/client/r0/rooms/%s/event/%s?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, roomId, eventId, accessCode)
	fmt.Println(endpoint)
	response, err := http.Get(endpoint)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return nil
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)
		m := f.(map[string]interface{})
		sender := m["sender"].(string)
		timeRecvd := m["origin_server_ts"].(string)
		mesg := m["content"].(map[string]interface{})["body"].(string)
		result := map[string]interface{}{
			"timestamp": timeRecvd,
			"message":   mesg,
			"sender":    sender,
			"roomId":    roomId,
		}
		return result
	}
} */
