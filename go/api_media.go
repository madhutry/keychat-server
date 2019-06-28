package friezechat

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadImagePut(w http.ResponseWriter, r *http.Request) {
	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		fmt.Println("Error")
	}
	defer r.Body.Close()
	reqToken := r.Header.Get("Authorization")
	filename := r.URL.Query().Get("filename")

	contentType := r.Header.Get("Content-Type")
	splitToken := strings.Split(reqToken, "Bearer")
	matAccessCode := strings.TrimSpace(splitToken[1])
	url := uploadToMatrix(matAccessCode, body, contentType)
	sEnc := b64.StdEncoding.EncodeToString(body)
	if strings.Contains(contentType, "image") {
		insertSSImage(fmt.Sprintf("data:%s;base64,%s", contentType, sEnc), url)
	} else {
		matUploadKey := url[strings.LastIndex(url, "/")+1:]
		uploadToS3(body, contentType, matUploadKey, 0, filename)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	jsonData := map[string]string{
		"content_uri": url,
	}
	enc := json.NewEncoder(w)
	enc.Encode(jsonData)
	w.WriteHeader(http.StatusOK)
}
func UploadImageOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
func uploadToMatrix(matrixAccessCode string, data []byte, contentType string) string {
	apiHost := "http://%s/media/upload?filename=sdfd.jpg"
	endpoint := fmt.Sprintf(apiHost, GetMatrixServerUrl())
	client := &http.Client{}
	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	request.Header.Add("Content-Type", contentType)
	request.Header.Add("Authorization", "Bearer "+matrixAccessCode)
	request.Header.Add("Accept", "*/*")
	response, err := client.Do(request)
	fmt.Print(response.StatusCode)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return ""
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var f interface{}
		json.Unmarshal([]byte(data), &f)
		m := f.(map[string]interface{})["content_uri"].(string)
		return m
	}
}
func insertSSImage(imagebase64 string, url string) {
	last := url[strings.LastIndex(url, "/")+1:]
	folder := fmt.Sprintf("%s/%s", "matrix_image", last)
	jsonData := map[string]string{
		"file":          imagebase64,
		"upload_preset": "lkqnuvmv",
		"folder":        folder,
	}
	apiHost := "https://api.cloudinary.com/v1_1/%s/image/upload"
	endpoint := fmt.Sprintf(apiHost, "frieze-in")
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Print("Image Succ Uploaded to Cloudinary" + string(data))
		insertCloudinary := `INSERT INTO image_upload_cloudinary		(		matrixid,			response		) VALUES
		(			$1,			$2		);
		
		`
		db := Envdb.db

		insertCloudinaryStmt, err := db.Prepare(insertCloudinary)
		if err != nil {
			log.Fatal(err)
		}
		defer insertCloudinaryStmt.Close()
		insertCloudinaryStmt.Exec(last, string(data))
	}
}
func uploadToS3(data []byte, contentType string, contentKey string, size int64, filename string) {
	aws_access_key_id := GetS3AccessKey()
	aws_secret_access_key := GetS3SecretKey()
	s3ParentFolderName := GetS3ParentFolder()
	bucketName := GetS3ObjectBucket()
	token := ""
	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, token)
	_, err := creds.Get()
	if err != nil {
		// handle error
	}
	cfg := aws.NewConfig().WithRegion("ap-south-1").WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	fileBytes := bytes.NewReader(data)
	path := s3ParentFolderName + "/" + contentKey
	params := &s3.PutObjectInput{
		Bucket:             aws.String(bucketName),
		Key:                aws.String(path),
		Body:               fileBytes,
		ContentDisposition: aws.String(fmt.Sprintf("attachment;filename=\"%s\"", filename)),
		ContentType:        aws.String(contentType),
	}
	resp, err := svc.PutObject(params)
	if err != nil {
		// handle error
	}
	etag := resp.ETag
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s/%s", bucketName, "ap-south-1", s3ParentFolderName, contentKey)
	insertS3Info := `	INSERT INTO s3_upload(matrix_content_id,etag, url  ) VALUES ($1,$2,$3)`

	db := Envdb.db

	insertS3InfoStmt, err := db.Prepare(insertS3Info)
	if err != nil {
		log.Fatal(err)
	}
	defer insertS3InfoStmt.Close()
	insertS3InfoStmt.Exec(contentKey, etag, url)
}
