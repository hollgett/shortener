package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ShortenerRequest struct {
	URL string `json:"url"`
}

type ShortenerResponse struct {
	ShortURL string `json:"result"`
}

func main() {
	JsonReq()
}

func JsonReq() {
	data, _ := json.Marshal(
		ShortenerRequest{
			URL: "ASDASDADASD",
		},
	)

	req, err := http.NewRequest("POST", "http://localhost:8080/api/shorten", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("post", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "gzip")
	client := http.Client{}
	resp, _ := client.Do(req)
	fmt.Println(resp.Status)
	fmt.Println("len", resp.ContentLength)
	fmt.Println("c", resp.Header.Get("Content-Encoding"))
	gzip, err := gzip.NewReader(resp.Body)
	if err != nil {
		fmt.Println("gzip", err)
		return
	}
	defer gzip.Close()
	bodyG, err := io.ReadAll(gzip)
	if err != nil {
		fmt.Println("read", err)
		return
	}
	respA := ShortenerResponse{}
	json.Unmarshal(bodyG, &respA)
	fmt.Println(respA)
}

func TestReq() {
	req, err := http.NewRequest("POST", "http://localhost:8080/test", nil)
	if err != nil {
		fmt.Println("post", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "gzip")
	client := http.Client{}
	resp, _ := client.Do(req)
	fmt.Println(resp.Status)

	fmt.Println("c", resp.Header.Get("Content-Encoding"))
	// body := resp.Body
	// sss, _ := io.ReadAll(body)
	// fmt.Println(string(sss))
	gzip, err := gzip.NewReader(resp.Body)
	if err != nil {
		fmt.Println("gzip", err)
		return
	}
	defer gzip.Close()
	bodyG, err := io.ReadAll(gzip)
	if err != nil {
		fmt.Println("read", err)
		return
	}
	fmt.Println(string(bodyG))
	// respA := ShortenerResponse{}
	// json.Unmarshal(bodyG, &respA)
	// fmt.Println(respA)
}
