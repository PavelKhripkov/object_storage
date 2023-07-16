package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"
	"time"
)

const (
	storeURL    = "http://localhost:11111/item/store"
	getURL      = "http://localhost:11111/item/%s"
	downloadURL = "http://localhost:11111/item/%s/download"

	fileField = "item"
	fileName  = "my_file.txt"
	fileLen   = 1024 * 1024
)

type RandomReader struct {
	n uint64
}

func NewRandomReader(n uint64) *RandomReader {
	return &RandomReader{n: n}
}

func (s *RandomReader) Read(p []byte) (int, error) {
	var m int
	var err error

	switch {
	case s.n == 0:
		return 0, io.EOF
	case s.n < uint64(len(p)):
		m, err = rand.Read(p[:s.n])
	default:
		m, err = rand.Read(p)
	}

	s.n -= uint64(m)
	return m, err
}

type item struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Size        int64  `json:"size,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	Status      string `json:"status,omitempty"`
}

func upload(url string, content io.Reader, contentLen int, params map[string]string) (*item, []byte) {
	byteBuf := &bytes.Buffer{}

	mpWriter := multipart.NewWriter(byteBuf)
	for key, value := range params {
		_ = mpWriter.WriteField(key, value)
	}

	mpWriter.CreateFormFile(fileField, fileName)
	contentType := mpWriter.FormDataContentType()

	nmulti := byteBuf.Len()
	multi := make([]byte, nmulti)
	_, _ = byteBuf.Read(multi)

	mpWriter.Close()
	nboundary := byteBuf.Len()
	lastBoundary := make([]byte, nboundary)
	_, _ = byteBuf.Read(lastBoundary)

	totalSize := int64(nmulti) + int64(contentLen) + int64(nboundary)

	preContent := bytes.NewReader(multi)
	postContent := bytes.NewReader(lastBoundary)

	hash := md5.New()
	contentTee := io.TeeReader(content, hash)

	contentReader := io.MultiReader(preContent, contentTee, postContent)

	req, err := http.NewRequest("POST", url, contentReader)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", contentType)
	req.ContentLength = totalSize

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	log.Println(resp.StatusCode)
	log.Println(resp.Header)

	itm := item{}
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&itm)
	if err != nil {
		log.Fatal(err)
	}

	return &itm, hash.Sum(nil)
}

func waitProcessing(url string, itm *item) *item {
	var ready bool

	respItem := item{}
	url = fmt.Sprintf(url, itm.ID)

	t := time.NewTimer(3 * time.Minute)
	defer t.Stop()

	for !ready {
		select {
		case <-t.C:
			log.Fatal("Timeout.")
		default:
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}

			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&respItem)
			if err != nil {
				log.Fatal(err)
			}

			ready = respItem.Status == "ok"

			resp.Body.Close()
			time.Sleep(time.Second)
		}
	}

	if itm.Size != respItem.Size ||
		itm.Name != respItem.Name ||
		itm.ContainerID != respItem.ContainerID ||
		itm.ID != respItem.ID {
		log.Fatal("Incorrect item data.")
	}

	return &respItem
}

func download(url string, itm *item) []byte {
	url = fmt.Sprintf(url, itm.ID)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	hash := md5.New()

	_, err = io.Copy(hash, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return hash.Sum(nil)
}

func main() {
	file := NewRandomReader(fileLen)

	// Uploading file
	itm, uploadHash := upload(storeURL, file, fileLen, map[string]string{"container_id": "123"})
	log.Print("Outgoing hash: ", uploadHash)

	// Waiting for file to be processed
	waitProcessing(getURL, itm)

	// Downloading file
	downloadHash := download(downloadURL, itm)
	log.Print("Incoming hash: ", downloadHash)

	// Comparing hashes
	if !reflect.DeepEqual(uploadHash, downloadHash) {
		log.Fatal("Hashes are different")
	}

	log.Print("Hashes are equal")
}
