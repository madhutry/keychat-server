package friezechat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"gopkg.in/gomail.v2"
)

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
func SaveFeedback(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	var f map[string]interface{}
	json.Unmarshal([]byte(body), &f)
	formToken := f["token"].(string)
	token, err := VerifyToken(strings.TrimSpace(formToken))
	if err != nil {
		log.Fatal(err)
	}
	domainName := token["DomainName"].(string)

	emailID := dbGetEmailId(domainName)
	db := Envdb.db
	insertFeedback := `
	INSERT INTO feedback ( domainname,feedback,ip_addr) VALUES( $1,  $2,$3 )	`
	insertFeedbackStmt, _ := db.Prepare(insertFeedback)
	delete(f["info"].(map[string]interface{}), "token")
	out, _ := json.Marshal(f["info"].(interface{}))

	_, err = insertFeedbackStmt.Exec(domainName, string(out), ReadUserIP(r))
	if err != nil {
		log.Fatal(err)
	}
	m := gomail.NewMessage()
	m.SetHeader("From", "***REMOVED***")
	m.SetHeader("To", emailID)
	m.SetHeader("Subject", "Hello!")
	name := f["info"].(map[string]interface{})["Name"].(string)
	email := f["info"].(map[string]interface{})["Email"].(string)
	mesg := f["info"].(map[string]interface{})["Message"].(string)

	m.SetBody("text/html", fmt.Sprintf(emailTemplate, name, email, mesg))
	d := gomail.NewDialer("smtp.gmail.com", 587, "***REMOVED***", "skhdhjzxudoeqsem")
	if err := d.DialAndSend(m); err != nil {
		panic(err)
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
