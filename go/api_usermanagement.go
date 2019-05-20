package friezechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	pborman "github.com/pborman/uuid"

)
matrixApiHost:="192.168.122.188"
func OpenChat(w http.ResponseWriter, r *http.Request) string{
  reqToken := r.Header.Get("Authorization")
  newFriezeChatAccessCode := pborman.NewRandom().String()

  domainName:= r.Host
	var accessCode string;
	if reqToken!=nill {
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		if err != nil {
			log.Fatal(err)
		}
		accessCode=token["FriezeAccessCode"]
  }
	body, readErr := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if readErr != nil {
		log.Fatal(readErr)
	}
	var f interface{}
	json.Unmarshal([]byte(body), &f)
  m := f.(map[string]interface{})
  if reqToken == nil {
    fullname := m["fullname"]
    mobileno := m["mobileno"]
    regId := pborman.NewRandom().String()
    registerMatrixChatUser(newFriezeChatAccessCode,regId)
    newJWTToken := GenerateToken(newFriezeChatAccessCode,domainName)
    return newJWTToken
  }else {
    matAccessCode,regId := getMatrixAccessCode(accessCode,domainName)
    dbDeactivateOldAccessCode(accessCode,domainName)
    registerMatrixChatUser(newFriezeChatAccessCode,regId)
    newJWTToken := GenerateToken(newFriezeChatAccessCode,domainName)
    return newJWTToken
  }
}
func dbDeactivateOldAccessCode(friezeChatAccessCode string,domainName string){
  deactivateAccCode := `UPDATE access_code_map SET active = 0 WHERE frieze_access_code = $1 AND
  domain_name = $2;`

  deactivateAccCodeStmt, err := db.Prepare(deactivateAccCode)
  if err != nil {
    log.Fatal(err)
  }
  defer deactivateAccCodeStmt.Close()
  _, err = deactivateAccCodeStmt.Exec(friezeChatAccessCode,domainName)
  if err != nil {
    panic(err)
  }
}
func dbInsertRegistration(fullName string,mobile string,friezeAccessCode string,regId string,roomId string,roomAlias string,prevBatchId string){

  insertRegister := `INSERT INTO chat_registration (registration_id,full_name,mobile,create_dt,room_id,room_alias,prev_batch_id)
  VALUES ($1,$2,$3,$4);`

  insertRegisterStmt, err := db.Prepare(insertRegister)
  if err != nil {
    log.Fatal(err)
  }
  defer insertRegisterStmt.Close()
  _, err = insertRegisterStmt.Exec(regID,fullName, mobile,friezeAccessCode,now(),roomId,roomAlias,prevBatchId)
  if err != nil {
    panic(err)
  }
}
func dbGetAllDetails(accessCode string , domainName string) string,string,string{
  var matAccCodeStr string
  var regId string

  matAccCode := `SELECT room_id,b.matrix_access_code,a.prev_batch_id
  FROM chat_registration a,access_code_map b
  where a.id=b.reg_id
  and b.frieze_access_code=$1
  and domain_name=$2`
  var roomId string
  var matAccessCode string
  var prevBatchId sql.NullString

	matAccCodeStmt, err := db.Prepare(matAccCode)
	if err != nil {
		log.Fatal(err)
	}
  matAccCodeStmt.QueryRow(friezeAccessCode,domainName).Scan(&roomId,&matAccessCode,&prevBatchId)
  return roomId,matAccessCode,prevBatchId.Value()
}
func dbDeactivateOldAccessCode(friezeChatAccessCode string,domainName string){
  deactivateAccCode := `UPDATE access_code_map SET active = 0 WHERE frieze_access_code = $1 AND
  domain_name = $2;`

  deactivateAccCodeStmt, err := db.Prepare(deactivateAccCode)
  if err != nil {
    log.Fatal(err)
  }
  defer deactivateAccCodeStmt.Close()
  _, err = deactivateAccCodeStmt.Exec(friezeChatAccessCode,domainName)
  if err != nil {
    panic(err)
  }
}
func getMatrixAccessCode(friezeAccessCode string,domainName string) (string,string){
  var matAccCodeStr string
  var regId string

  matAccCode := `select matrix_access_code,registration_id from access_code_map where  
  frieze_access_code=$1 and domain_name=$2 and active=1`
	matAccCodeStmt, err := db.Prepare(matAccCode)
	if err != nil {
		log.Fatal(err)
	}
  matAccCodeStmt.QueryRow(friezeAccessCode,domainName).Scan(&matAccCodeStr,&regId)
  return matAccCodeStr
}
func registerMatrixChatUser(friezeAccessCode string,domainName string,regId string){
	jsonData := map[string]string{
		"auth": {
			"session": "ffdfdasfdsfadsf",
			"type": "m.login.dummy"
		},
		"username": friezeAccessCode[:5],
		"password": "palava123",
		"bind_email": false,
		"bind_msisdn": false,
		"x_show_msisdn": false
	}
	apiHost := "http://%s/_matrix/client/r0/register"
	endpoint := fmt.Sprintf(matrixApiHost)
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
    data, _ := ioutil.ReadAll(response.Body)
    var f interface{}
    json.Unmarshal([]byte(body), &f)

    m := f.(map[string]interface{})
    matrixAccessCode := m["access_token"].(string)
		db := Envdb.db
		if err != nil {
			panic(err)
		}
    err = db.Ping()
    roomId,roomAlias := apiCreateRoom(matrixAccessCode)
    f:=apiGetMessages(matAccessCode,roomId,nil)
    startBatchId:=f["startBatch"].(string)
    dbInsertRegistration(fullname,mobileno,newFriezeChatAccessCode,regId,roomId,roomAlias,startBatchId)

    addAccessCodeMap := `INSERT INTO access_code_map 
    (  matrix_access_code, frieze_access_code,domain_name,insert_dt,registration_id)
    VALUES (    $1 , $2,$3,$4,$5,$6,$7 )	`

		addAccessCodeMapStmt, err := db.Prepare(addAccessCodeMap)
		if err != nil {
			log.Fatal(err)
		}
		defer addAccessCodeMapStmt.Close()
		_, err = addAccessCodeMapStmt.Exec(matrixAccessCode,friezeChatAccessCode,domainName, now(),regId)
		if err != nil {
			panic(err)
    }
    return matrixAccessCode
	}
}
func apiCreateRoom(accessCode string) (string,string){
  roomAlias := pborman.NewRandom().String()

  jsonData := {
    "room_alias_name":roomAlias
	}
	apiHost := "http://%s/_matrix/client/r0/createRoom?access_token=%s"
	endpoint := fmt.Sprintf(matrixApiHost,accessCode)
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
    data, _ := ioutil.ReadAll(response.Body)
    var f interface{}
    json.Unmarshal([]byte(body), &f)

    m := f.(map[string]interface{})
    roomID := m["room_id"].(string)
    roomAlias := m["room_alias"].(string)
    return roomID,roomAlias 
  }

}
func GetMessages(w http.ResponseWriter, r *http.Request) string{
  reqToken := r.Header.Get("Authorization")
  domainName:= r.Host
	var accessCode string;
	if reqToken!=nill {
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		if err != nil {
			log.Fatal(err)
		}
    accessCode=token["FriezeAccessCode"]
    roomId,matAccessCode,prevBatchID := dbGetAllDetails(accessCode,domainName)
    result:=apiGetMessages(matAccessCode,roomId,prevBatchID)
    enc := json.NewEncoder(w) //
    enc.Encode(result)
  }else{
    log.Fatal("No Tockemmn")
  }
}
func apiGetMessages(accessCode string,roomId string,previousBatch string) (map[string]interface{}){
  roomAlias := pborman.NewRandom().String()
  fromPrevBatch := ""
  jsonData := {
    "room_alias_name":roomAlias
	}
  apiHost := "http://%s/_matrix/client/r0/rooms/%s/messages?access_token==%s%s"
  if(previousBatch != nil ){
    fromPrevBatch=fmt.Sprintf("?from=%s",previousBatch)
  }
  
	endpoint := fmt.Sprintf(matrixApiHost,roomId,accessCode,fromPrevBatch)
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
      data, _ := ioutil.ReadAll(response.Body)
      var f interface{}
      json.Unmarshal([]byte(body), &f)
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
      f := map[string]interface{}{
        "startBatch": startbatch,
        "messages":   messages,
      }
      return f
  }
}