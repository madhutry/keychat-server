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

func SubmitChat(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	newFriezeChatAccessCode := pborman.NewRandom().String()

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
		extraInfo := m["extrainfo"]
		reqToken := m["token"]
		token, err := VerifyToken(strings.TrimSpace(reqToken.(string)))
		if err != nil {
			fmt.Println("Could not verify token")
			log.Fatal(err)
		}
		domainNm := token["DomainName"].(string)
		regId := pborman.NewRandom().String()
		_, userId, avatarUrl, welcomeMsg := registerMatrixChatUser(fullname.(string), mobileno.(string), extraInfo, newFriezeChatAccessCode, domainNm, regId)
		newJWTToken, _ := GenerateTokenWithUserID(newFriezeChatAccessCode, domainNm, userId, fullname.(string))
		tokenJson := Token{newJWTToken, avatarUrl, welcomeMsg}

		enc := json.NewEncoder(w)
		enc.Encode(&tokenJson)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
	}
}
func OpenChat(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	newFriezeChatAccessCode := pborman.NewRandom().String()
	//var cookie, err = r.Cookie("DomainName")
	domainNm := "goodtm"
	/* 	if err == nil {
	   		domainNm = cookie.Value
	   		log.Println("get cookie value is " + domainNm + "")
	   	} else {
	   		log.Fatal("---No Domain Name Cookie Found")
	   	}
	*/
	var accessCode string
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
		_, userId, avatarUrl, welcomeMsg := registerMatrixChatUser(fullname.(string), mobileno.(string), nil, newFriezeChatAccessCode, domainNm, regId)
		newJWTToken, _ := GenerateTokenWithUserID(newFriezeChatAccessCode, domainNm, userId, fullname.(string))
		tokenJson := Token{newJWTToken, avatarUrl, welcomeMsg}

		enc := json.NewEncoder(w)
		enc.Encode(&tokenJson)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
	} else {
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		if err != nil {
			log.Fatal(err)
		}
		accessCode = token["FriezeAccessCode"].(string)
		domainName := token["DomainName"].(string)
		fullName := token["Fullname"].(string)
		userId := token["UserId"].(string)
		matAccessCode, regId := getMatrixAccessCode(accessCode, domainName)
		dbDeactivateOldAccessCode(accessCode, domainName)
		dbInsertNewAccessCode(matAccessCode, newFriezeChatAccessCode, domainName, regId)
		newJWTToken, _ := GenerateTokenWithUserID(newFriezeChatAccessCode, domainNm, userId, fullName)
		tokenJson := Token{newJWTToken, "", ""}
		enc := json.NewEncoder(w)
		enc.Encode(&tokenJson)
	}
}
func checkOwnerOnline(domainName string) map[string]interface{} {
	apiHost := "http://%s/_matrix/client/r0/presence/%s/status?access_token=%s"
	agentsDetails, displayNames := dbGetAgents(domainName)
	prsent := "offline"
	timeactive := float64(0)
	displayName := ""
	for i, v := range agentsDetails {
		endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), v, GetMatrixAdminCode())
		response, err := http.Get(endpoint)
		if err != nil {
			fmt.Printf("GetOnlineStatus The HTTP request failed with error %s\n", err)
			log.Fatal(err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			var f interface{}
			json.Unmarshal([]byte(data), &f)
			m := f.(map[string]interface{})
			prsent = m["presence"].(string)
			timeactive = m["last_active_ago"].(float64)
			displayName = displayNames[i]
			if prsent == "online" {
				break
			}
		}
	}
	result := map[string]interface{}{"msgType": "checkOnline", "online": prsent, "activetime": timeactive, "displayName": displayName}
	return result
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
func dbInsertRegistrationExtra(fullName string, mobile string, extra interface{},
	friezeAccessCode string, regId string, roomId string, roomAlias string, prevBatchId string, userId string) {

	insertRegister := `INSERT INTO chat_registration (id,full_name,mobile,create_dt,room_id,room_alias,prev_batch_id,user_id)
  VALUES ($1,$2,$3,$4,$5,$6,$7,$8);`
	db := Envdb.db

	insertRegisterStmt, err := db.Prepare(insertRegister)
	if err != nil {
		log.Fatal(err)
	}
	defer insertRegisterStmt.Close()
	_, err = insertRegisterStmt.Exec(regId, fullName, mobile, time.Now(), roomId, roomAlias, prevBatchId, userId)
	if err != nil {
		panic(err)
	}
}
func dbInsertRegistration(fullName string, mobile string,
	friezeAccessCode string, regId string, roomId string, roomAlias string, prevBatchId string, userId string) {

	insertRegister := `INSERT INTO chat_registration (id,full_name,mobile,create_dt,room_id,room_alias,prev_batch_id,user_id)
  VALUES ($1,$2,$3,$4,$5,$6,$7,$8);`
	db := Envdb.db

	insertRegisterStmt, err := db.Prepare(insertRegister)
	if err != nil {
		log.Fatal(err)
	}
	defer insertRegisterStmt.Close()
	_, err = insertRegisterStmt.Exec(regId, fullName, mobile, time.Now(), roomId, roomAlias, prevBatchId, userId)
	if err != nil {
		panic(err)
	}
}
func dbGetMessages(friezeAccCd string) ([][]string, int16) {
	/* 	selectMesg := `select message,sender,a.server_received_ts,a.mesg_id from messages a , chat_registration b, access_code_map c
	   	where
	   	a.room_id=b.room_id
	   	and b.id=c.registration_id
	   	and c.frieze_access_code=$1
	   	and a.customer_read=0
	   	order by a.create_ts asc` */

	selectMesg := `select a.id,message,COALESCE(d.display_name, sender),a.server_received_ts,a.mesg_id from  chat_registration b, access_code_map c,messages a  LEFT OUTER JOIN agents d
 	ON d.userid=sender
	where
	a.room_id=b.room_id
	and b.id=c.registration_id
	and c.frieze_access_code=$1
	and a.customer_read=0 
	order by a.create_ts asc;
	`
	db := Envdb.db

	rows, err := db.Query(selectMesg, friezeAccCd)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var messages [][]string
	var messageTxt string
	var sender string
	var timestamp string
	var msgid string
	var msgSerialId int16
	for rows.Next() {
		rows.Scan(&msgSerialId, &messageTxt, &sender, &timestamp, &msgid)
		mesg1 := []string{messageTxt, timestamp, sender, msgid}
		messages = append(messages, mesg1)
	}
	return messages, msgSerialId
}
func dbGetAllDetails(accessCode string, domainName string) (string, string, string, string) {
	log.Println("dbGetAllDetails:" + accessCode + ":" + domainName)
	matAccCode := `SELECT room_id,b.matrix_access_code,a.prev_batch_id,a.user_id
  FROM chat_registration a,access_code_map b
  where a.id=b.registration_id
  and b.frieze_access_code=$1
  and domain_name=$2`
	var roomId string
	var matAccessCode string
	var prevBatchId sql.NullString
	var userId string

	db := Envdb.db

	matAccCodeStmt, err := db.Prepare(matAccCode)
	if err != nil {
		log.Fatal(err)
	}
	matAccCodeStmt.QueryRow(accessCode, domainName).Scan(&roomId, &matAccessCode, &prevBatchId, &userId)
	val, _ := prevBatchId.Value()
	return roomId, matAccessCode, val.(string), userId
}

func dbGetNotifcationDetails(roomId string) string {

	matAccCode := `select a.matrix_access_code from access_code_map a , chat_registration b
	where b.room_id=$1 and a.registration_id=b.id
`
	var matAccessCode string
	db := Envdb.db

	matAccCodeStmt, err := db.Prepare(matAccCode)
	if err != nil {
		log.Fatal(err)
	}
	matAccCodeStmt.QueryRow(roomId).Scan(&matAccessCode)
	return matAccessCode
}
func dbGetAgents(domainName string) ([]string, []string) {

	userIdsSql := `SELECT 
	userid,display_name
	FROM mat_acc_cd_owner
	WHERE domain_name=$1
	UNION
	SELECT 
	b.userid,b.display_name
	FROM mat_acc_cd_owner a, agents b
	WHERE domain_name=$2
	and a.id=b.main_owner_id`

	db := Envdb.db

	userIdsStmt, err := db.Prepare(userIdsSql)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := userIdsStmt.Query(domainName, domainName)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var userIds []string
	var displayNames []string

	for rows.Next() {
		var userId string
		var displayName string
		rows.Scan(&userId, &displayName)
		userIds = append(userIds, userId)
		displayNames = append(displayNames, displayName)
	}
	return userIds, displayNames
}
func dbGetDomainRelatedData(domainName string) ([]string, [][]string) {

	matAccCode := `SELECT 
	1,matrix_access_code,avatar_img_url,display_name
	FROM mat_acc_cd_owner
	WHERE domain_name=$1
UNION
SELECT 
	2,b.matrix_access_code,b.avatar_img_url,b.display_name
	FROM mat_acc_cd_owner a, agents b
	WHERE domain_name=$2
	and a.id=b.main_owner_id order by 1
`
	db := Envdb.db

	matAccCodeStmt, err := db.Prepare(matAccCode)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := matAccCodeStmt.Query(domainName, domainName)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var codes []string
	var statusInfo [][]string
	for rows.Next() {
		var order int16
		var accCd string
		var avatarUrl string
		var welcomeMsg string
		rows.Scan(&order, &accCd, &avatarUrl, &welcomeMsg)
		result := []string{avatarUrl, welcomeMsg}
		codes = append(codes, accCd)
		statusInfo = append(statusInfo, result)
	}
	return codes, statusInfo
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
func registerMatrixChatUser(fullname string, mobileno string, extraInfo interface{},
	friezeAccessCode string, domainName string, regId string) (string, string, string, string) {
	username := friezeAccessCode[:5]
	jsonData := map[string]interface{}{
		"auth": map[string]string{
			"session": "ffdfdasfdsfadsf",
			"type":    "m.login.dummy",
		},
		"username":      username,
		"password":      "palava123",
		"bind_email":    false,
		"bind_msisdn":   false,
		"x_show_msisdn": false,
	}
	apiHost := "http://%s/_matrix/client/r0/register"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl())
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return "", "", "", ""
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)

		m := f.(map[string]interface{})
		matrixAccessCode := m["access_token"].(string)
		userId := m["user_id"].(string)
		db := Envdb.db
		if err != nil {
			panic(err)
		}
		err = db.Ping()
		apiPutDisplayName(matrixAccessCode, userId, fullname)
		roomId, roomAlias := apiCreateRoom(matrixAccessCode, fullname, username, mobileno)
		log.Println("Adding Owner to room :" + domainName)
		ownerDetails, profileInfo := dbGetDomainRelatedData(domainName)
		accessCdAdmin := GetMatrixAdminCode()
		accessCodes := append(ownerDetails, accessCdAdmin)
		apiJoinRoom(accessCodes, roomId)
		result := apiGetMessages(matrixAccessCode, roomId, "")
		startBatchId := result["startBatch"].(string)
		if extraInfo == nil {
			dbInsertRegistration(fullname, mobileno, friezeAccessCode, regId, roomId, roomAlias, startBatchId, userId)
		} else {
			dbInsertRegistrationExtra(fullname, mobileno, extraInfo, friezeAccessCode, regId, roomId, roomAlias, startBatchId, userId)
		}
		dbInsertNewAccessCode(matrixAccessCode, friezeAccessCode, domainName, regId)
		return matrixAccessCode, userId, profileInfo[0][0], profileInfo[0][1]
	}
}

func apiPutDisplayName(accessCode string, userId string, fullname string) {
	jsonData := map[string]string{
		"displayname": fullname,
	}
	apiHost := "http://%s/_matrix/client/r0/profile/%s/displayname?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), userId, accessCode)
	jsonValue, _ := json.Marshal(jsonData)
	client := &http.Client{}

	request, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(jsonValue))
	request.Header.Add("Content-type", "application/json")
	_, err = client.Do(request)
	/*
		jsonValue, _ := json.Marshal(jsonData)
		response, err := http.P(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	*/if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
}
func apiGetDisplayName(accessCode string, userId string, fullname string) {
	apiHost := "http://%s/_matrix/client/r0/profile/%s/displayname?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, userId, accessCode)
	client := &http.Client{}

	request, err := http.NewRequest("GET", endpoint, nil)
	request.Header.Add("Content-type", "application/json")
	_, err = client.Do(request)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
}
func apiCreateRoom(accessCode string, fullname string, userid string, mobileno string) (string, string) {
	roomAlias := fmt.Sprintf("%s-%s", fullname, userid)

	jsonData := map[string]string{
		"room_alias_name": roomAlias,
		"topic":           fmt.Sprintf("Name:%s,Mobile:%s", fullname, mobileno),
	}
	apiHost := "http://%s/_matrix/client/r0/createRoom?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), accessCode)
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
	var accessCode string
	if len(reqToken) > 0 {
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		if err != nil {
			log.Fatal(err)
		}
		accessCode = token["FriezeAccessCode"].(string)
		domainName := token["DomainName"].(string)

		_, _, _, userId := dbGetAllDetails(accessCode, domainName)
		result := make(map[string]interface{})
		result["messages"], _ = dbGetMessages(accessCode)
		result["userId"] = userId
		enc := json.NewEncoder(w) //
		enc.Encode(result)
	} else {
		log.Fatal("No Tockemmn")
	}
}
func GetMessages1(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	var accessCode string
	if len(reqToken) > 0 {
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		if err != nil {
			log.Fatal(err)
		}
		accessCode = token["FriezeAccessCode"].(string)
		domainName := token["DomainName"].(string)

		roomId, matAccessCode, prevBatchID, userId := dbGetAllDetails(accessCode, domainName)
		result := apiGetMessages(matAccessCode, roomId, prevBatchID)
		result["userId"] = userId
		enc := json.NewEncoder(w) //
		enc.Encode(result)
	} else {
		log.Fatal("No Tockemmn")
	}
}
func retrieveToken(reqToken string) map[string]interface{} {
	splitToken := strings.Split(reqToken, "Bearer")
	reqToken = splitToken[1]
	token, err := VerifyToken(strings.TrimSpace(reqToken))
	if err != nil {
		log.Fatal(err)
	}
	return token
}
func SendMessage(w http.ResponseWriter, r *http.Request) {
	reqToken := r.Header.Get("Authorization")
	var accessCode string
	if len(reqToken) > 1 {
		token := retrieveToken(reqToken)
		accessCode = token["FriezeAccessCode"].(string)
		domainName := token["DomainName"].(string)

		body, readErr := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if readErr != nil {
			log.Fatal(readErr)
		}
		var f map[string]string
		json.Unmarshal([]byte(body), &f)

		sendMessage(accessCode, domainName, f["message"], f["uuid"])
	} else {
		log.Fatal("No Tockemmn")
	}
}
func sendMessage(friezeAccessCode string, domainName string, message string, uuid string) {
	roomId, matAccessCode, _, _ := dbGetAllDetails(friezeAccessCode, domainName)
	apiSendMessage(matAccessCode, roomId, message, uuid)
}
func apiGetMessages(accessCode string, roomId string, previousBatch string) map[string]interface{} {
	fromPrevBatch := ""
	apiHost := "http://%s/_matrix/client/r0/rooms/%s/messages?limit=1000&access_token=%s%s"
	if len(previousBatch) > 1 {
		fromPrevBatch = fmt.Sprintf("&from=%s", previousBatch)
	}

	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), roomId, accessCode, fromPrevBatch)
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
			typeKey := chunk.(map[string]interface{})["type"].(string)
			if typeKey == "m.room.message" {
				mesg := chunk.(map[string]interface{})["content"].(map[string]interface{})["body"].(string)
				ts := chunk.(map[string]interface{})["origin_server_ts"].(float64)
				sender := chunk.(map[string]interface{})["sender"].(string)

				tsString := fmt.Sprintf("%f", ts)
				mesg1 := []string{mesg, tsString, sender}
				messages = append(messages, mesg1)
			}
		}
		result := map[string]interface{}{
			"startBatch": startbatch,
			"messages":   messages,
		}
		return result
	}
}

func apiSendMessage(matAccessCode string, roomId string, message string, uuid string) {
	jsonData := map[string]string{
		"msgtype":  "m.text",
		"body":     message,
		"trans_id": uuid,
	}
	apiHost := "http://%s/_matrix/client/r0/rooms/%s/send/m.room.message?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), roomId, matAccessCode)
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

func apiJoinRoom(matAccCode []string, roomId string) {
	for _, codes := range matAccCode {
		jsonData := map[string]string{}
		apiHost := "http://%s/_matrix/client/r0/join/%s?access_token=%s"
		endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), roomId, codes)
		fmt.Println(endpoint + "---")
		jsonValue, _ := json.Marshal(jsonData)
		response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
			return
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			var f interface{}
			json.Unmarshal([]byte(data), &f)
			fmt.Println("lsl")
		}
	}
}
