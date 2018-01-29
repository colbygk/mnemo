package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type OkeaService struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	TrustTLS bool   `json:"trusttls"`
	Token    string `json:"token"`
}

type Credentials struct {
	Token string
}

func FileToToken(filename string) (Credentials, error) {

	var newtoken Credentials

	tokenbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return Credentials{}, err
	}

	err = json.Unmarshal(tokenbytes, &newtoken)

	return newtoken, err
}

func newTLS(trust bool) *http.Client {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: trust}}
	return &http.Client{Transport: tr}
}

func (oks *OkeaService) Post(endpoint string, request string) (string, error) {
	buf := bytes.NewReader([]byte(request))
	client := newTLS(oks.TrustTLS)
	resp, err := client.Post(fmt.Sprintf("https://%s:%d/%s", oks.Hostname, oks.Port, endpoint),
		"Application/json", buf)

	if err != nil {
		return "", err
	}

	body, berr := ioutil.ReadAll(resp.Body)

	return string(body), berr
}

func (oks *OkeaService) Login() (string, error) {

	req_json := fmt.Sprintf("{\"Username\":\"%s\", \"Password\":\"%s\"}",
		oks.Username, oks.Password)

	return oks.Post("token-auth", req_json)
}

func (oks *OkeaService) Hello() (string, error) {
	log.Printf("hello %v\n", oks.Token)
	client := newTLS(oks.TrustTLS)
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s:%d/test/hello", oks.Hostname, oks.Port), nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create new request: %v\n", err))
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", oks.Token))
	resp, rerr := client.Do(req)
	if rerr != nil {
		log.Fatal(fmt.Sprintf("Unable to complete request: %v\n", rerr))
	}
	body, berr := ioutil.ReadAll(resp.Body)
	return string(body), berr
}

func (oks *OkeaService) ListProjects() (string, error) {
	log.Printf("list\n")

	client := newTLS(oks.TrustTLS)
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s:%d/cnames/hello", oks.Hostname, oks.Port), nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create new request: %v\n", err))
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", oks.Token))
	resp, rerr := client.Do(req)
	if rerr != nil {
		log.Fatal(fmt.Sprintf("Unable to complete request: %v\n", rerr))
	}
	body, berr := ioutil.ReadAll(resp.Body)
	return string(body), berr
}
