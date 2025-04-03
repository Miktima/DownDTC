package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func getRes(url string) (int, error) {
	// функция получения ресурсов по указанному адресу url с использованием User-Agent
	// возвращает загруженный HTML контент
	client := &http.Client{}
	var Scode int
	Scode = 0

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Cannot create new request  %s, error: %v\n", url, err)
		return Scode, err
	}

	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		//fmt.Printf("Error with GET request: %v\n", err)
		return Scode, err
	}
	Scode = resp.StatusCode
	defer resp.Body.Close()

	return Scode, nil
}

func telega(apikey, resource, err string, code int, chats []string) {

	var tgbody string

	reli := regexp.MustCompile(`<.*?>`)
	resmb := regexp.MustCompile(`([_\*\[\]\(\)~\>\#\+\-\=\|\{\}\.!])`)

	resource = reli.ReplaceAllString(resource, "")
	resource = resmb.ReplaceAllString(resource, "\\$1")

	tgbody = "*" + resource + "*\n"
	tgbody += "Error code: " + strconv.Itoa(code) + "\n"

	err = reli.ReplaceAllString(err, "")
	err = resmb.ReplaceAllString(err, "\\$1")

	tgbody += err

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Send message to Telegram
	client := &http.Client{}

	url := "https://api.telegram.org/bot" + apikey + "/sendMessage"

	// Если тело сообщения не превышает предельный размер, то отсылаем сообщение как обычно
	for _, tgid := range chats {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Cannot create new request  %s, error: %v\n", url, err)
		}
		q := req.URL.Query()
		q.Add("parse_mode", "MarkdownV2")
		q.Add("chat_id", tgid)
		q.Add("disable_web_page_preview", "1")
		q.Add("text", tgbody)
		req.URL.RawQuery = q.Encode()
		// Отправляем запрос
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error with GET request: %v\n", err)
		}
		if resp.StatusCode > 299 {
			fmt.Println("Message with was not sent")
			fmt.Println("Resource: ", resource)
			fmt.Println("Error: ", err)
			fmt.Println("Error code: ", code)
			fmt.Println("TGBODE:", tgbody)
		}
		defer resp.Body.Close()
	}
}

func main() {
	// Определяем путь
	path, _ := os.Executable()
	path = path[:strings.LastIndex(path, "/")+1]

	type Resources struct {
		Resource string
		Cron     string
		Chats    []string
	}

	type Configtg struct {
		APIkey string
	}

	var resources []Resources
	var configtg Configtg
	// Читаем файл с ресурсами
	if _, err := os.Stat(path + "/resources.json"); err == nil {
		// Open our jsonFile
		byteValue, err := os.ReadFile(path + "/resources.json")
		// if we os.ReadFile returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		// defer the closing of our jsonFile so that we can parse it later on
		// var listHash []ArticleH
		err = json.Unmarshal(byteValue, &resources)
		if err != nil {
			fmt.Println(err)
		}
	}

	// Читаем файл с настройками telegram
	if _, err := os.Stat(path + "/botkey.json"); err == nil {
		// Open our jsonFile
		byteValue, err := os.ReadFile(path + "/botkey.json")
		// if we os.ReadFile returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		// defer the closing of our jsonFile so that we can parse it later on
		// var listHash []ArticleH
		err = json.Unmarshal(byteValue, &configtg)
		if err != nil {
			fmt.Println(err)
		}
	}

	// Читаем файлы со списками. Файлы в порядке, указанном в конфигурационном файле
	for _, res := range resources {

		statuscode, err := getRes(res.Resource)
		if err != nil || statuscode != 200 {
			// TODO: call telegram
			telega(configtg.APIkey, res.Resource, err.Error(), statuscode, res.Chats)
			//fmt.Printf("Error - %v\n", err)
			//fmt.Printf("Resource - %s\n", res.Resource)
			//fmt.Printf("Status Code - %d\n", statuscode)
		}
	}
}
