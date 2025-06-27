package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <input> <output>\n", os.Args[0])
		os.Exit(1)
	}
	inPath := os.Args[1]
	outPath := os.Args[2]

	data, err := ioutil.ReadFile(inPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read input:", err)
		os.Exit(1)
	}

	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		fmt.Fprintln(os.Stderr, "parse yaml:", err)
		os.Exit(1)
	}

	db, ok := m["db"]
	if !ok {
		fmt.Fprintln(os.Stderr, "no db section found")
		os.Exit(1)
	}

	dbBytes, err := yaml.Marshal(db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal db section:", err)
		os.Exit(1)
	}

	enc := base64.StdEncoding.EncodeToString(dbBytes)
	delete(m, "db")
	m["db_enc"] = enc

	outData, err := yaml.Marshal(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal output:", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(outPath, outData, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "write output:", err)
		os.Exit(1)
	}
}
