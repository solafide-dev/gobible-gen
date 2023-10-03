package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

func main() {

	t := flag.String("version", "KJV", "The version of the bible to generate")
	flag.Parse()

	log.Printf("Generating %s bible", *t)

	b := initFromBibleGateway(*t)

	// marshal and save to file

	json, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile("./generated/"+*t+".json", json, 0644)
}
