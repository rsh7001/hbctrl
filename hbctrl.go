package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	hb "github.com/rstanleyhum/handbookappdb"
	"golang.org/x/net/html"
)

func main() {
	commandPtr := flag.String("cmd", "load", "Command")
	tablePtr := flag.String("table", "fullpage", "Table name")
	filename := flag.String("infile", "", "Input Filename")
	indir := flag.Bool("indir", false, "Is Directory flag")
	intype := flag.String("intype", "html", "Input Filename Type")
	flag.Parse()

	fmt.Println("command: ", *commandPtr)
	fmt.Println("table:   ", *tablePtr)
	fmt.Println("infile:  ", *filename)
	fmt.Println("intype:  ", *intype)
	fmt.Println("indir:   ", *indir)

	if !*indir && !isInFile(*filename) {
		log.Fatalln("Not a valid filename")
	}

	if *indir && !isInFileDirectory(*filename) {
		log.Fatalln("Not a valid directory")
	}

	var url string
	var js string
	var method string
	var err error

	switch {
	case !*indir && *commandPtr == "load":
		method = "POST"
		url, err = doGetLoadURL(*tablePtr)
		if err != nil {
			log.Fatalf("Not valid URL for table: %v\n", *tablePtr)
		}

		switch *intype {
		case "html":
			js, err = doLoadHTML(*tablePtr, *filename)
		case "json":
			js, err = doLoadJSON(*tablePtr, *filename)
		default:
			log.Fatalf("No a valid input filename type: %v\n", *intype)
		}
		if err != nil {
			log.Fatalf("Not valid payload from file: %v\n", *filename)
		}
		err = apiSend(url, method, js)
		if err != nil {
			log.Fatalf("apiSend Error: %v", err)
		}
	case *indir && *commandPtr == "load":
		method = "POST"
		url, err = doGetLoadURL(*tablePtr)
		if err != nil {
			log.Fatalf("Not valid URL for table\n")
		}

		files, err := ioutil.ReadDir(*filename)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			fullfilename := *filename + "/" + file.Name()
			switch *intype {
			case "html":
				js, err = doLoadHTML(*tablePtr, fullfilename)
			case "json":
				js, err = doLoadJSON(*tablePtr, fullfilename)
			default:
				log.Fatalf("Not a valid intype\n")
			}
			if err != nil {
				log.Fatalf("Not valid payload from file\n")
			}
			err = apiSend(url, method, js)
			if err != nil {
				log.Fatalf("apiSend Error: %v", err)
			}
		}
	default:
		log.Fatalln("Not a valid command")
	}
}

func doGetLoadURL(table string) (url string, err error) {
	base := "http://localhost:55506/"
	switch table {
	case "fullpage":
		url = base + "tables/fullpageitem/"
	case "book":
		url = base + "tables/bookitem/"
	case "licencekey":
		url = base + "tables/licencekeyitem/"
	case "userupdatestatus":
		url = base + "tables/userupdatestatusitem/"
	case "initialupdatejson":
		url = base + "tables/initialupdatejsonitem/"
	default:
		err = errors.New("Not defined table")
	}
	return
}

func apiSend(url string, method string, js string) (err error) {
	fmt.Println(url)
	fmt.Println(method)
	fmt.Println(js)

	payload := []byte(js)
	request, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("ZUMO-API-VERSION", "2.0.0")
	request.Header.Set("X-ZUMO-AUTH", "--token-here--")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err != nil {
		return err
	}

	return nil
}

func isInFile(i string) bool {
	info, err := os.Stat(i)
	if err != nil {
		log.Printf("Infile: %v has a problem.\n", i)
		log.Printf("%v\n", err.Error())
		return false
	}
	if info.IsDir() {
		log.Printf("Infile: %v is not a file\n", i)
		return false
	}
	return true
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

func doLoadHTML(table string, filename string) (js string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}

	switch table {
	case "fullpage":
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
	default:
		err = errors.New("No table defined")
		return
	}

	return
}

func doLoadJSON(table string, filename string) (js string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	jsdecoder := json.NewDecoder(f)

	switch table {
	case "fullpage":
		var fp hb.Fullpage
		err = jsdecoder.Decode(&fp)
		if err != nil {
			return
		}
		var payload []byte
		payload, err = json.Marshal(fp)
		if err != nil {
			return
		}
		js = string(payload)
	case "book":
		var bk hb.Book
		err = jsdecoder.Decode(&bk)
		if err != nil {
			return
		}
		var payload []byte
		payload, err = json.Marshal(bk)
		if err != nil {
			return
		}
		js = string(payload)
	case "licencekey":
		var lk hb.LicenceKey
		err = jsdecoder.Decode(&lk)
		if err != nil {
			return
		}
		var payload []byte
		payload, err = json.Marshal(lk)
		if err != nil {
			return
		}
		js = string(payload)
	case "initialupdatejson":
		var iuj hb.InitialUpdateJson
		err = jsdecoder.Decode(&iuj)
		if err != nil {
			return
		}
		var payload []byte
		payload, err = json.Marshal(iuj)
		if err != nil {
			return
		}
		js = string(payload)
	case "userupdatestatus":
		var ujm hb.UpdateJsonMessage
		err = jsdecoder.Decode(&ujm)
		if err != nil {
			return
		}
		var messagePayload []byte
		messagePayload, err = json.Marshal(ujm)
		if err != nil {
			return
		}
		messageJs := string(messagePayload)
		var uus hb.UserUpdateStatus
		uus.ID = "humrs"
		uus.UpdateNeeded = false
		uus.UpdateJson = string(messageJs)
		var payload []byte
		payload, err = json.Marshal(uus)
		if err != nil {
			return
		}
		js = string(payload)
	default:
		err = errors.New("No table defined")
		return
	}

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
