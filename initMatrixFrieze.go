package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func apiRegisterOwnerAgent(w http.ResponseWriter, r *http.Request) {
	body, readErr := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if readErr != nil {
		log.Fatal(readErr)
	}
	var f interface{}
	json.Unmarshal([]byte(body), &f)

	m := f.(map[string]interface{})
	username := m["username"].(string)
	domainName := m["domainname"].(string)
	fullname := m["fullname"].(string)
	agentOwnerAdmin := m["type"].(string)
	session := getSessionId(username)
	jsonData := map[string]interface{}{
		"auth": map[string]string{
			"session": session,
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
		log.Fatal(err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)

		m := f.(map[string]interface{})
		matrixAccessCode := m["access_token"].(string)
		userId := m["user_id"].(string)
		db := Envdb.db
		err = db.Ping()
		if err != nil {
			panic(err)
		}
		apiPutDisplayName(matrixAccessCode, userId, fullname)
		_, _, _, ids := dbGetDomainRelatedData(domainName)
		if agentOwnerAdmin == "admin" {
			filterId := setFilterForAdmin(userId, matrixAccessCode)
			// if filterId == "0" {
			// 	panic("Filter Cannot Zero")
			// }
			dbDeactivateOldAdmin()
			dbInsertAdmin(userId, matrixAccessCode, filterId)
		} else if agentOwnerAdmin == "owner" {
			dbInsertOwner(matrixAccessCode, domainName, userId, "http://google.com", fullname)
		} else if agentOwnerAdmin == "agent" {
			dbInsertAgent(ids[0], matrixAccessCode, domainName, userId, "http://google.com", fullname)
		}
	}
}
func setFilterForAdmin(userid string, accesscd string) string {
	// body, readErr := ioutil.ReadAll(r.Body)
	// defer r.Body.Close()
	// if readErr != nil {
	// 	log.Fatal(readErr)
	// }
	// var f interface{}
	// json.Unmarshal([]byte(body), &f)

	// m := f.(map[string]interface{})
	// userid := m["userid"].(string)
	// accesscd := m["accesscd"].(string)

	apiHost := "http://%s/_matrix/client/r0/user/%s/filter?access_token=%s"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl(), userid, accesscd)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer([]byte(getFilterPayload())))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		log.Fatal(err)
		return "0"
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)
		m := f.(map[string]interface{})
		filterId := m["filter_id"].(string)
		return filterId
	}
}
func dbInsertOwner(accessCode string, domainName string, userId string, url, dispName string) {
	insertRegister := `INSERT INTO public.mat_acc_cd_owner(
		domain_name, matrix_access_code, userid, avatar_img_url, display_name)
		VALUES ($1,$2,$3,$4,$5);`
	db := Envdb.db
	insertRegisterStmt, err := db.Prepare(insertRegister)
	if err != nil {
		log.Fatal(err)
	}
	defer insertRegisterStmt.Close()
	_, err = insertRegisterStmt.Exec(domainName, accessCode, userId, nil, dispName)
	if err != nil {
		panic(err)
	}

}
func dbInsertAgent(ownerAccCdId int, accessCode string, domainName string, userId string, url, dispName string) {
	insertRegister := `INSERT INTO public.agents(
		userid, display_name, main_owner_id, matrix_access_code, avatar_img_url)
		VALUES ($1,$2,$3,$4,$5)`
	db := Envdb.db
	insertRegisterStmt, err := db.Prepare(insertRegister)
	if err != nil {
		log.Fatal(err)
	}
	defer insertRegisterStmt.Close()
	_, err = insertRegisterStmt.Exec(userId, dispName, ownerAccCdId, accessCode, "http://google.com")
	if err != nil {
		panic(err)
	}
}
func dbInsertAdmin(user_id string, access_code string, filter_id string) {
	insertRegister := `INSERT INTO public.admin_info(
		userid, access_code, filter_id)
		VALUES ($1, $2, $3)`
	db := Envdb.db
	insertRegisterStmt, err := db.Prepare(insertRegister)
	if err != nil {
		log.Fatal(err)
	}
	defer insertRegisterStmt.Close()
	_, err = insertRegisterStmt.Exec(user_id, access_code, filter_id)
	if err != nil {
		panic(err)
	}
}
func dbDeactivateOldAdmin() {
	deactivateAccCode := `UPDATE public.admin_info SET active = 'N'`
	db := Envdb.db
	deactivateAccCodeStmt, err := db.Prepare(deactivateAccCode)
	if err != nil {
		log.Fatal(err)
	}
	defer deactivateAccCodeStmt.Close()
	_, err = deactivateAccCodeStmt.Exec()
	if err != nil {
		panic(err)
	}
}
func getFilterPayload() string {
	payload := `{
		"account_data": {
			"not_types":[
				"*"
			 ]
		},
		"presence": {
		  "not_types":[
			"*"
		  ]
		},
		"room": {
			  "timeline": {
			"types": [
			  "m.room.message"
			]
		  },
		  "state": {
				"not_types":[
				"*"
			 ]
		  },
		  "ephemeral": {
			  "not_types":[
				"*"
			 ]
		  }
		}
	  }`
	return payload
}
