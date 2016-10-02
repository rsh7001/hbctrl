package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/SermoDigital/jose/jws"
	hb "github.com/rstanleyhum/handbookappdb"
)

type RestResult struct {
	Results []hb.AppLog
	Count   int
}

func main() {
	base := flag.String("url", "http://localhost:55506/", "base Url")
	keyfile := flag.String("keyfile", "", "Key File")
	outputfile := flag.String("outfile", "output.csv", "Outputfilename")

	flag.Parse()

	fmt.Println("url: ", *base)
	fmt.Println("keyfile: ", *keyfile)
	fmt.Println("outfile: ", *outputfile)

	file, err := os.Create(*outputfile)
	if err != nil {
		log.Fatalf("Cannot create output file: %v\n", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	token := getToken(*keyfile)

	url := *base + "tables/AppLogItem/"
	method := "GET"

	top := 50
	skip := 0

	resultsCount := 0

	var totalresults []hb.AppLog

	defer writer.Flush()

	for {
		httpstring := fmt.Sprintf("%s?$top=%v&$skip=%v&$inlinecount=allpages", url, top, skip)
		fmt.Println(httpstring)
		payload, err := apiSend(httpstring, method, token)
		if err != nil {
			log.Fatalf("Error is: %v", err)
		}

		var results RestResult
		err = json.Unmarshal(payload, &results)
		if err != nil {
			log.Fatalf("Error is: %v", err)
		}

		resultsCount = resultsCount + top
		totalresults = append(totalresults, results.Results...)

		for _, applog := range results.Results {
			data := []string{applog.UserID, applog.LogDateTime, applog.LogName, applog.LogDataJson}
			err := writer.Write(data)
			if err != nil {
				log.Fatal("Cannot write to file: ", err)
			}
		}

		fmt.Printf("resultsCount: %v; totalresults %v; results %v\n", resultsCount, results.Count, len(totalresults))

		if resultsCount > results.Count {
			break
		}

		skip = skip + top
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

func apiSend(url string, method string, token string) (payload []byte, err error) {
	inpayload := []byte("")
	request, err := http.NewRequest(method, url, bytes.NewBuffer(inpayload))
	if err != nil {
		return nil, err
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
		return nil, err
	}
	fmt.Printf("Response: %v\n", response.Status)
	defer response.Body.Close()

	payload, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
