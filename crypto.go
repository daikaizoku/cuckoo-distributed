package main

import (
	"crypto/md5"
	"crypto/sha256"
	"io"
	"log"
	"os"
)

func sha256sum(filepath string) string {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err.Error())
	}
	return string(h.Sum(nil))
}

func md5sum(filepath string) string {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err.Error())
	}
	return string(h.Sum(nil))
}
