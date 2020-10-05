package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/gomail.v2"
)

func SaveRegister(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	var f map[string]interface{}
	json.Unmarshal([]byte(body), &f)
	fullname := f["fullname"].(string)
	emailid := f["emailid"].(string)
	if len(fullname) > 20 {
		fullname = fullname[0:20]
	}
	if len(emailid) > 20 {
		emailid = emailid[0:10]
	}

	db := Envdb.db
	insertFeedback := `
	INSERT INTO register ( fname,emailid,ip_addr) VALUES( $1,  $2,$3 )	`
	insertFeedbackStmt, _ := db.Prepare(insertFeedback)

	_, err := insertFeedbackStmt.Exec(fullname, emailid, ReadUserIP(r))
	if err != nil {
		log.Fatal(err)
	}
	m := gomail.NewMessage()
	m.SetHeader("From", "***REMOVED***")
	m.SetHeader("To", "***REMOVED***")
	m.SetHeader("Subject", "New Registration!")
	name := fullname
	email := emailid
	mesg := "You have new registration."

	m.SetBody("text/html", fmt.Sprintf(emailTemplate, name, email, mesg))
	d := gomail.NewDialer("smtp.gmail.com", 587, "***REMOVED***", "***REMOVED***")
	if err := d.DialAndSend(m); err != nil {
		log.Println("Could Send Email..")
	}
}

func dbGetEmailId(domainname string) string {
	db := Envdb.db
	emailIdOwner := `
		select email_id from mat_acc_cd_owner b where domain_name=$1;
	`
	emailIdOwnerStmt, err := db.Prepare(emailIdOwner)
	if err != nil {
		log.Fatal(err)
	}
	var emailid string
	emailIdOwnerStmt.QueryRow(domainname).Scan(&emailid)
	return emailid
}

const emailTemplate string = `

	<b>Name</b> : %s<br>
	<b>Email</b> : %s<br>
	<b>Message</b> : %s<br>
`
