package main

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/SermoDigital/jose/jws"
	hb "github.com/rstanleyhum/handbookappdb"
)

func main() {
	base := flag.String("url", "http://localhost:55506/", "base Url")
	keyfile := flag.String("keyfile", "", "Key File")

	flag.Parse()

	fmt.Println("url: ", *base)
	fmt.Println("keyfile: ", *keyfile)

	token := getToken(*keyfile)

	url := *base + "tables/licencekeyitem/"
	method := "POST"

	seedrand()

	fmt.Printf("Token: %v\n", token)
	fmt.Printf("url: %v\n", url)
	fmt.Printf("method: %v\n", method)

	for i := 0; i < 200; i++ {

		var lk hb.LicenceKey

		lk.HandbookType = "CHONY"
		lk.ID = strings.ToLower(RandStringRunes(6))

		var payload []byte
		payload, err := json.Marshal(lk)
		if err != nil {
			log.Fatal("Problem with payload")
		}
		js := string(payload)

		err = apiSend(url, method, js, token)
		if err != nil {
			log.Fatalf("apiSend Error: %v", err)
		}
	}
}

func seedrand() {
	rand.Seed(time.Now().UnixNano())
}

func getToken(keyfile string) string {
	signkeyString, err := ioutil.ReadFile(keyfile)
	if err != nil {
		log.Fatal("error reading priveate key")
	}

	signkey, err := hex.DecodeString(string(signkeyString))

	claims := jws.Claims{}
	claims.Set("sub", "humrs")
	claims.Set("ver", "3")
	claims.Set("iss", "https://handbookmobileappservice.azurewebsites.net/")
	claims.Set("aud", "https://handbookmobileappservice.azurewebsites.net/")
	claims.Set("exp", 1498867200)
	claims.Set("nbf", 1467331200)

	signMethod := jws.GetSigningMethod("HS256")
	token := jws.NewJWT(claims, signMethod)
	byteToken, err := token.Serialize(signkey)
	if err != nil {
		log.Fatal("Error signing the key. ", err)
	}

	return string(byteToken)
}

func apiSend(url string, method string, js string, token string) (err error) {
	payload := []byte(js)
	request, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("ZUMO-API-VERSION", "2.0.0")
	request.Header.Set("X-ZUMO-AUTH", token)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	fmt.Printf("Response: %v\n", response.Status)
	defer response.Body.Close()

	if err != nil {
		return err
	}

	return nil
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
