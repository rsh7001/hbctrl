package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	hb "github.com/rstanleyhum/handbookappdb"
)

func main() {
	infilename := flag.String("infile", "", "infile json")

	flag.Parse()

	fmt.Println("infile: ", *infilename)

	outfilename := *infilename + ".txt"
	fmt.Println(outfilename)
	outfp, err := os.Create(outfilename)

	if err != nil {
		log.Fatal(err)
	}

	defer outfp.Close()

	lklist, err := doLoadJSON(*infilename)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range lklist {
		line := fmt.Sprintf("%v|%v|%v\n", v.ID, v.HandbookType, v.UserID)
		outfp.WriteString(line)
	}

	fmt.Println(len(lklist))

}

func doLoadJSON(filename string) (lklist []hb.LicenceKey, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	jsdecoder := json.NewDecoder(f)

	err = jsdecoder.Decode(&lklist)
	if err != nil {
		return
	}

	return
}
