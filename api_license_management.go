package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	pborman "github.com/pborman/uuid"
)

func ValidLicense(w http.ResponseWriter, r *http.Request) {
	body, readErr := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if readErr != nil {
		log.Fatal(readErr)
	}
	var f interface{}
	json.Unmarshal([]byte(body), &f)
	m := f.(map[string]interface{})
	mainLicenseCode := m["main_license_code"].(string)
	domainName := m["domain_name"].(string)
	newPsuedoLic := pborman.NewRandom().String()
	id := dbLicenseValid(mainLicenseCode, domainName)
	if id == 0 {
		dbInsertPsuedoLic(newPsuedoLic, id, domainName, "N")
	} else {
		dbInsertPsuedoLic(newPsuedoLic, id, domainName, "Y")
	}
	tokenJson := map[string]string{"lic": newPsuedoLic}
	enc := json.NewEncoder(w)
	enc.Encode(tokenJson)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
func dbLicenseValid(mainLicenseCode string, domainName string) int64 {
	db := Envdb.db
	licenseMasterSql := `
	select id from master_lic where now()>start_date and now()<expiry_date 
		and lic_secret=$1 and domain_name=$2;
	`
	licenseMasterStmt, err := db.Prepare(licenseMasterSql)
	if err != nil {
		log.Fatal(err)
	}
	var valSql sql.NullFloat64
	licenseMasterStmt.QueryRow(mainLicenseCode, domainName).Scan(&valSql)
	val, err := valSql.Value()
	if val == nil || err != nil {
		return 0
	} else {
		return int64(val.(float64))
	}
}
func dbInsertPsuedoLic(pseudoLic string, masterLicId int64, domainName string, valid string) {
	db := Envdb.db
	insertFeedback := `	INSERT INTO pseudo_lic ( pseudo_lic, master_lic_id,domain_name,valid	) VALUES ($1,$2,$3,$4);	`
	insertFeedbackStmt, _ := db.Prepare(insertFeedback)
	_, err := insertFeedbackStmt.Exec(pseudoLic, masterLicId, domainName, valid)
	if err != nil {
		log.Fatal(err)
	}
}

func dbCheckFailedPseudoLic(pseudo_lic string, domainName string) int64 {
	db := Envdb.db
	failedLicSql := `
	select count(*) from pseudo_lic where pseudo_lic=$1 and domain_name=$2;
	`
	failedLicStmt, err := db.Prepare(failedLicSql)
	if err != nil {
		log.Fatal(err)
	}
	var count int64
	failedLicStmt.QueryRow(pseudo_lic, domainName).Scan(&count)
	return count
}
