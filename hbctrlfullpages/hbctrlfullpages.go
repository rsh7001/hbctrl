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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SermoDigital/jose/jws"
	hb "github.com/rstanleyhum/handbookappdb"
	"golang.org/x/net/html"
)

func main() {
	dirname := flag.String("indir", "", "Input Directory name")
	base := flag.String("url", "http://localhost:55506/", "base Url")
	keyfile := flag.String("keyfile", "", "Key File")

	flag.Parse()

	fmt.Println("infile:  ", *dirname)
	fmt.Println("url: ", *base)
	fmt.Println("keyfile: ", *keyfile)

	if !isInFileDirectory(*dirname) {
		log.Fatalln("Not a valid directory")
	}

	token := getToken(*keyfile)

	url := *base + "tables/fullpageitem/"
	method := "POST"

	files, err := ioutil.ReadDir(*dirname)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fullfilename := *dirname + "/" + file.Name()
		js, err := doLoadHTML(fullfilename)
		if err != nil {
			log.Fatalf("Not valid payload from file: %v\n", file.Name())
		}

		err = apiSend(url, method, js, token)
		if err != nil {
			log.Fatalf("apiSend Error: %v", err)
		}
	}
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

func isInFileDirectory(i string) bool {
	info, err := os.Stat(i)
	if err != nil {
		log.Printf("Infile does not exist\n")
		return false
	}
	if info.IsDir() {
		return true
	}
	return false
}

func doLoadHTML(filename string) (js string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}

	var htmlbyte []byte
	name := filepath.Base(filename)
	id := name[:len(name)-5]
	htmlbyte, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}

	var fp hb.Fullpage
	fp, err = htmlToFullpage(id, string(htmlbyte))
	if err != nil {
		return
	}

	var payload []byte
	payload, err = json.Marshal(fp)
	if err != nil {
		return
	}
	js = string(payload)

	return
}

func htmlToFullpage(id string, htmlstring string) (fp hb.Fullpage, err error) {
	fp.ID = id
	fp.Content = string(htmlstring)
	z, err := html.Parse(strings.NewReader(fp.Content))
	if err != nil {
		return
	}

	var ff func(*html.Node)
	ff = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			fp.Title = n.FirstChild.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			ff(c)
		}
	}

	ff(z)
	return
}
