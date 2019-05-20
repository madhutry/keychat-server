package friezechat

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	pborman "github.com/pborman/uuid"
)

const matrixApiHost = "192.168.122.188"

func OpenChat(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	newFriezeChatAccessCode := pborman.NewRandom().String()

	domainName := r.Host
	var accessCode string
	if len(reqToken) > 0 {
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		if err != nil {
			log.Fatal(err)
		}
		accessCode = token["FriezeAccessCode"].(string)
	}
	body, readErr := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if readErr != nil {
		log.Fatal(readErr)
	}
	var f interface{}
	json.Unmarshal([]byte(body), &f)
	m := f.(map[string]interface{})
	if len(reqToken) == 0 {
		fullname := m["fullname"]
		mobileno := m["mobileno"]
		regId := pborman.NewRandom().String()
		registerMatrixChatUser(fullname.(string), mobileno.(string), newFriezeChatAccessCode, domainName, regId)
		newJWTToken, _ := GenerateToken(newFriezeChatAccessCode, domainName)
		tokenJson := Token{newJWTToken}

		enc := json.NewEncoder(w)
		enc.Encode(&tokenJson)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
	} else {
		matAccessCode, regId := getMatrixAccessCode(accessCode, domainName)
		dbDeactivateOldAccessCode(accessCode, domainName)
		dbInsertNewAccessCode(matAccessCode, newFriezeChatAccessCode, domainName, regId)
		newJWTToken, _ := GenerateToken(newFriezeChatAccessCode, domainName)
		tokenJson := Token{newJWTToken}
		enc := json.NewEncoder(w)
		enc.Encode(&tokenJson)
	}
}
func dbDeactivateOldAccessCode(friezeChatAccessCode string, domainName string) {
	deactivateAccCode := `UPDATE access_code_map SET active = 0 WHERE frieze_access_code = $1 AND
  domain_name = $2;`
	db := Envdb.db
	deactivateAccCodeStmt, err := db.Prepare(deactivateAccCode)
	if err != nil {
		log.Fatal(err)
	}
	defer deactivateAccCodeStmt.Close()
	_, err = deactivateAccCodeStmt.Exec(friezeChatAccessCode, domainName)
	if err != nil {
		panic(err)
	}
}
func dbInsertRegistration(fullName string, mobile string, friezeAccessCode string, regId string, roomId string, roomAlias string, prevBatchId string) {

	insertRegister := `INSERT INTO chat_registration (id,full_name,mobile,create_dt,room_id,room_alias,prev_batch_id)
  VALUES ($1,$2,$3,$4,$5,$6,$7);`
	db := Envdb.db

	insertRegisterStmt, err := db.Prepare(insertRegister)
	if err != nil {
		log.Fatal(err)
	}
	defer insertRegisterStmt.Close()
	_, err = insertRegisterStmt.Exec(regId, fullName, mobile, time.Now(), roomId, roomAlias, prevBatchId)
	if err != nil {
		panic(err)
	}
}
func dbGetAllDetails(accessCode string, domainName string) (string, string, string) {

	matAccCode := `SELECT room_id,b.matrix_access_code,a.prev_batch_id
  FROM chat_registration a,access_code_map b
  where a.id=b.registration_id
  and b.frieze_access_code=$1
  and domain_name=$2`
	var roomId string
	var matAccessCode string
	var prevBatchId sql.NullString
	db := Envdb.db

	matAccCodeStmt, err := db.Prepare(matAccCode)
	if err != nil {
		log.Fatal(err)
	}
	matAccCodeStmt.QueryRow(accessCode, domainName).Scan(&roomId, &matAccessCode, &prevBatchId)
	val, _ := prevBatchId.Value()
	return roomId, matAccessCode, val.(string)
}

func dbInsertNewAccessCode(matrixAccessCode string, friezeChatAccessCode string, domainName string, regId string) {
	addAccessCodeMap := `INSERT INTO access_code_map 
    (  matrix_access_code, frieze_access_code,domain_name,insert_dt,registration_id)
    VALUES (    $1 , $2,$3,$4,$5 )	`
	db := Envdb.db

	addAccessCodeMapStmt, err := db.Prepare(addAccessCodeMap)
	if err != nil {
		log.Fatal(err)
	}
	defer addAccessCodeMapStmt.Close()
	_, err = addAccessCodeMapStmt.Exec(matrixAccessCode, friezeChatAccessCode, domainName, time.Now(), regId)
	if err != nil {
		panic(err)
	}
}
func getMatrixAccessCode(friezeAccessCode string, domainName string) (string, string) {
	var matAccCodeStr string
	var regId string

	matAccCode := `select matrix_access_code,registration_id from access_code_map where  
  frieze_access_code=$1 and domain_name=$2 and active=1`
	db := Envdb.db

	matAccCodeStmt, err := db.Prepare(matAccCode)
	if err != nil {
		log.Fatal(err)
	}
	matAccCodeStmt.QueryRow(friezeAccessCode, domainName).Scan(&matAccCodeStr, &regId)
	return "", matAccCodeStr
}
func registerMatrixChatUser(fullname string, mobileno string,
	friezeAccessCode string, domainName string, regId string) string {
	jsonData := map[string]interface{}{
		"auth": map[string]string{
			"session": "ffdfdasfdsfadsf",
			"type":    "m.login.dummy",
		},
		"username":      friezeAccessCode[:5],
		"password":      "palava123",
		"bind_email":    false,
		"bind_msisdn":   false,
		"x_show_msisdn": false,
	}
	apiHost := "http://%s/_matrix/client/r0/register"
	endpoint := fmt.Sprintf(apiHost, matrixApiHost)
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return ""
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)

		m := f.(map[string]interface{})
		matrixAccessCode := m["access_token"].(string)
		db := Envdb.db
		if err != nil {
			panic(err)
		}
		err = db.Ping()
		roomId, roomAlias := apiCreateRoom(matrixAccessCode)
		result := apiGetMessages(matrixAccessCode, roomId, "")
		startBatchId := result["startBatch"].(string)
		dbInsertRegistration(fullname, mobileno, friezeAccessCode, regId, roomId, roomAlias, startBatchId)
		dbInsertNewAccessCode(matrixAccessCode, friezeAccessCode, domainName, regId)
		return matrixAccessCode
	}
}
func apiCreateRoom(accessCode string) (string, string) {
	roomAlias := pborman.NewRandom().String()

	jsonData := map[string]string{
		"room_alias_name": roomAlias,
	}
	apiHost := "http://%s/_matrix/client/r0/createRoom?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, matrixApiHost, accessCode)
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return "", ""
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)

		m := f.(map[string]interface{})
		roomID := m["room_id"].(string)
		roomAlias := m["room_alias"].(string)
		return roomID, roomAlias
	}

}
func GetMessages(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	domainName := r.Host
	var accessCode string
	if len(reqToken) > 0 {
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		if err != nil {
			log.Fatal(err)
		}
		accessCode = token["FriezeAccessCode"].(string)
		roomId, matAccessCode, prevBatchID := dbGetAllDetails(accessCode, domainName)
		result := apiGetMessages(matAccessCode, roomId, prevBatchID)
		enc := json.NewEncoder(w) //
		enc.Encode(result)
	} else {
		log.Fatal("No Tockemmn")
	}
}
func SendMessage(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	domainName := r.Host
	var accessCode string
	if len(reqToken) > 1 {
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		if err != nil {
			log.Fatal(err)
		}
		accessCode = token["FriezeAccessCode"].(string)
		body, readErr := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if readErr != nil {
			log.Fatal(readErr)
		}
		var f map[string]string
		json.Unmarshal([]byte(body), &f)
		roomId, matAccessCode, _ := dbGetAllDetails(accessCode, domainName)
		apiSendMessage(matAccessCode, roomId, f["message"])
	} else {
		log.Fatal("No Tockemmn")
	}
}
func apiGetMessages(accessCode string, roomId string, previousBatch string) map[string]interface{} {
	fromPrevBatch := ""
	apiHost := "http://%s/_matrix/client/r0/rooms/%s/messages?access_token=%s%s"
	if len(previousBatch) > 1 {
		fromPrevBatch = fmt.Sprintf("&from=%s", previousBatch)
	}

	endpoint := fmt.Sprintf(apiHost, matrixApiHost, roomId, accessCode, fromPrevBatch)
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
		startbatch := m["start"].(string)
		chunks := m["chunk"].([]interface{})
		var messages [][]string
		for _, chunk := range chunks {
			mesg := chunk.(map[string]interface{})["content"].(map[string]interface{})["body"].(string)
			ts := chunk.(map[string]interface{})["origin_server_ts"].(float64)
			tsString := fmt.Sprintf("%f", ts)
			mesg1 := []string{mesg, tsString}
			messages = append(messages, mesg1)
		}
		result := map[string]interface{}{
			"startBatch": startbatch,
			"messages":   messages,
		}
		return result
	}
}

func apiSendMessage(matAccessCode string, roomId string, message string) {
	jsonData := map[string]string{
		"msgtype": "m.text",
		"body":    message,
	}
	apiHost := "http://%s/_matrix/client/r0/rooms/%s/send/m.room.message?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, matrixApiHost, roomId, matAccessCode)
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)
	}
}
