package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func getHtmlPage(url, userAgent string) ([]byte, int, error) {
	// функция получения ресурсов по указанному адресу url с использованием User-Agent
	// возвращает загруженный HTML контент
	client := &http.Client{}
	var Scode int
	Scode = 0

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Cannot create new request  %s, error: %v\n", url, err)
		return nil, Scode, err
	}

	// с User-agent по умолчанию контент не отдается, используем свой
	req.Header.Set("User-Agent", userAgent)

	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error with GET request: %v\n", err)
		return nil, Scode, err
	}
	Scode = resp.StatusCode
	defer resp.Body.Close()

	// Получаем контент и возвращаем его, как результат работы функции
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error ReadAll")
		return nil, Scode, err
	}
	return body, Scode, nil
}

func getListFsm(body []byte, tag string) []string {
	// Функция получения списка из html контента
	// Список достается из тега tag
	// Возвращает список
	tkn := html.NewTokenizer(bytes.NewReader(body))
	depth := 0
	var flist []string
	errorCode := false
	var trimedstr string

	// Проходим по всему дереву тегов (пока не встретится html.ErrorToken)
	for !errorCode {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			errorCode = true
		case html.TextToken:
			if depth > 0 {
				trimedstr = strings.Trim(string(tkn.Text()), " \n\t")
				if len(trimedstr) > 0 { //Пустые строки не забираем
					flist = append(flist, trimedstr) // Если внутри нужного тега, забираем текст из блока
				}
			}
		case html.StartTagToken, html.EndTagToken:
			tn, tAttr := tkn.TagName()
			if string(tn) == "div" { // выбираем нужный tag
				if tAttr {
					key, attr, _ := tkn.TagAttr()
					if tt == html.StartTagToken && string(key) == "id" && string(attr) == tag {
						depth++ // нужный тег открывается
					}
				} else if tt == html.EndTagToken && depth >= 1 {
					depth--
				}
			}
		}
	}
	return flist
}

func getListMinjust(body []byte) []string {
	// Функция получения списка из html контента
	// Список достается из тега tag
	// Возвращает список
	tkn := html.NewTokenizer(bytes.NewReader(body))
	depth := 0
	var flist []string
	errorCode := false
	var trimedstr string
	acctext := ""
	validNum := regexp.MustCompile(`^[ ]*[0-9]+.+`)

	// Проходим по всему дереву тегов (пока не встретится html.ErrorToken)
	for !errorCode {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			errorCode = true
		case html.TextToken:
			if depth > 0 {
				trimedstr = strings.Trim(string(tkn.Text()), " \n\t")
				if len(trimedstr) > 0 { //Пустые строки не забираем
					acctext += trimedstr + " " // Если внутри нужного тега, забираем текст из блока
				}
			}
		case html.StartTagToken, html.EndTagToken:
			tn, _ := tkn.TagName()
			if string(tn) == "tr" { // выбираем нужный tag
				if tt == html.StartTagToken {
					depth++ // нужный тег открывается
				} else if tt == html.EndTagToken && depth >= 1 {
					if validNum.MatchString(acctext) { // Строка должна начинаться с числа
						flist = append(flist, acctext) // При закрытии тега добавляем в список
					}
					acctext = ""
					depth--
				}
			}
		}
	}
	return flist
}

func getListSpimex(body []byte) []string {
	// Функция получения списка из html контента
	// Список достается из тега tag
	// Возвращает список
	tkn := html.NewTokenizer(bytes.NewReader(body))
	depth := 0
	othertag := 0
	var flist []string
	errorCode := false
	var trimedstr string
	acctext := ""

	// Проходим по всему дереву тегов (пока не встретится html.ErrorToken)
	for !errorCode {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			errorCode = true
		case html.TextToken:
			if depth > 0 {
				trimedstr = strings.Trim(string(tkn.Text()), " \n\t")
				if len(trimedstr) > 0 { //Пустые строки не забираем
					acctext += trimedstr + " " // Если внутри нужного тега, забираем текст из блока
				}
			}
		case html.StartTagToken, html.EndTagToken:
			tn, tAttr := tkn.TagName()
			if string(tn) == "div" { // выбираем нужный tag
				// fmt.Println("depth:", depth, "     othertag:", othertag)
				if tAttr {
					key, attr, _ := tkn.TagAttr()
					if tt == html.StartTagToken {
						// fmt.Println("key:", string(key), "     attr:", string(attr))
						if depth == 1 {
							othertag++ // считаем другие такие же теги внутри нужного
						}
						if string(key) == "class" && string(attr) == "news-item" {
							depth++ // нужный тег открывается
						}
					}
				} else if tt == html.EndTagToken && depth == 1 {
					if othertag == 0 {
						flist = append(flist, acctext) // При закрытии тега добавляем в список
						acctext = ""
						depth--
					} else {
						othertag--
					}
				}
			}
		}
	}
	return flist
}

func getListAcra(body []byte) []string {
	// Функция получения списка из html контента
	// Список достается из тега tag
	// Возвращает список
	tkn := html.NewTokenizer(bytes.NewReader(body))
	depth := 0
	othertag := 0
	var flist []string
	errorCode := false
	var trimedstr string
	acctext := ""

	// Проходим по всему дереву тегов (пока не встретится html.ErrorToken)
	for !errorCode {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			errorCode = true
		case html.TextToken:
			if depth > 0 {
				trimedstr = strings.Trim(string(tkn.Text()), " \n\t")
				if len(trimedstr) > 0 { //Пустые строки не забираем
					acctext += trimedstr + " " // Если внутри нужного тега, забираем текст из блока
				}
			}
		case html.StartTagToken, html.EndTagToken:
			tn, tAttr := tkn.TagName()
			if string(tn) == "div" { // выбираем нужный tag
				// fmt.Println("depth:", depth, "     othertag:", othertag)
				if tAttr {
					key, attr, _ := tkn.TagAttr()
					if tt == html.StartTagToken {
						// fmt.Println("key:", string(key), "     attr:", string(attr))
						if depth == 1 {
							othertag++ // считаем другие такие же теги внутри нужного
						}
						if string(key) == "class" && string(attr) == "documents-row__wrapper search-table-row__wrapper" {
							depth++ // нужный тег открывается
						}
					}
				} else if tt == html.EndTagToken && depth == 1 {
					if othertag == 0 {
						flist = append(flist, acctext) // При закрытии тега добавляем в список
						acctext = ""
						depth--
					} else {
						othertag--
					}
				}
			}
		}
	}
	return flist
}

func getListMintrans(body []byte) []string {
	// Функция получения списка из html контента
	// Список достается из тега tag
	// Возвращает список
	tkn := html.NewTokenizer(bytes.NewReader(body))
	depth := 0
	othertag := 0
	var flist []string
	errorCode := false
	var trimedstr string
	acctext := ""

	// Проходим по всему дереву тегов (пока не встретится html.ErrorToken)
	for !errorCode {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			errorCode = true
		case html.TextToken:
			if depth > 0 {
				trimedstr = strings.Trim(string(tkn.Text()), " \n\t")
				if len(trimedstr) > 0 { //Пустые строки не забираем
					acctext += trimedstr + " " // Если внутри нужного тега, забираем текст из блока
				}
			}
		case html.StartTagToken, html.EndTagToken:
			tn, tAttr := tkn.TagName()
			if string(tn) == "div" { // выбираем нужный tag
				// fmt.Println("depth:", depth, "     othertag:", othertag)
				if tAttr {
					key, attr, _ := tkn.TagAttr()
					if tt == html.StartTagToken {
						// fmt.Println("key:", string(key), "     attr:", string(attr))
						if depth == 1 {
							othertag++ // считаем другие такие же теги внутри нужного
						}
						if string(key) == "class" && string(attr) == "news-list-item" {
							depth++ // нужный тег открывается
						}
					}
				} else if tt == html.EndTagToken && depth == 1 {
					if othertag == 0 {
						acctext += "[PAD]"
						flist = append(flist, acctext) // При закрытии тега добавляем в список
						acctext = ""
						depth--
					} else {
						othertag--
					}
				}
			}
		}
	}
	return flist
}

func testEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func newList(get_list []string, byteValue []byte, rexp, order string) []string {
	// возвращает список только измененных значений
	// order = 'asc', если более свежие строики в конце списка, 'desc' - наоборот
	// выбираем все признаки строк
	var new_list []string
	reorder := regexp.MustCompile(rexp)
	order_list := reorder.FindAll(byteValue, -1)
	if len(order_list) == 0 {
		return new_list
	}
	// Добавляем в список только новые строки
	if order == "asc" {
		newlines := 0
		for _, v := range get_list {
			if newlines == 1 {
				new_list = append(new_list, "<li style=\"background-color:#ffff99\">"+v+"</li>")
			}
			if bytes.Contains([]byte(v), order_list[len(order_list)-1]) {
				newlines = 1
			}
		}
	} else if order == "desc" {
		newlines := 1
		for _, v := range get_list {
			if bytes.Contains([]byte(v), order_list[0]) {
				newlines = 0
			}
			if newlines == 1 {
				new_list = append(new_list, "<li style=\"background-color:#ffff99\">"+v+"</li>")
			}
		}
	}
	return new_list
}

func mail(newlist []string, listName, urlList string, addressList []string) {
	var title string
	var titleLink string
	var subject string
	if strings.Contains(listName, "UL") {
		title = "Юридические лица"
	} else if strings.Contains(listName, "FL") {
		title = "Физические лица"
	}
	if strings.Contains(listName, "UL") || strings.Contains(listName, "FL") {
		subject = "Subject: New list Federal Financial Monitoring Service: " + title + "\n"
	} else if listName == "Minjust" {
		subject = "Subject: New list Minjust: нежелательные организации\n"
	} else if listName == "Spimex" {
		subject = "Subject: Биржевая информация: статистические материалы\n"
	} else if listName == "Acra" {
		subject = "Subject: АКРА рейтинг\n"
	} else if listName == "Mintrans" {
		subject = "Subject: Министерство транспорта Российской Федерации\n"
	}
	adfrom := "From: robot@ria.ru\nReply-To: robot@ria.ru\n"
	address := "To: "
	n_address := 0
	for _, a := range addressList {
		if n_address > 0 {
			address += ", "
		}
		address += a
		n_address += 1
	}
	address += "\n"
	htmlhead := "<html>"
	if strings.Contains(listName, "UL") || strings.Contains(listName, "FL") {
		htmlhead += "<head><title>New list Federal Financial Monitoring Service</title>"
	} else if listName == "Minjust" {
		htmlhead += "<head><title>New list Minjust: нежелательные организации</title>"
	} else if listName == "Spimex" {
		htmlhead += "<head><title>Биржевая информация: статистические материалы</title>"
	} else if listName == "Acra" {
		htmlhead += "<head><title>АКРА рейтинг: пресс-релизы</title>"
	} else if listName == "Mintrans" {
		htmlhead += "<head><title>Министерство транспорта Российской Федерации</title>"
	}
	htmlhead += "<meta charset=\"utf-8\">"
	htmlhead += "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">"
	if strings.Contains(listName, "add") {
		title += " (включенные)"
		htmlhead += "</head><body><h1>" + title + "</h1>"
	} else if strings.Contains(listName, "del") {
		title += " (исключённые)"
		htmlhead += "</head><body><h1>" + title + "</h1>"
	} else if listName == "Minjust" {
		htmlhead += "</head><body><h1>Перечень иностранных и международных неправительственных организаций, деятельность которых признана нежелательной на территории Российской Федерации</h1>"
	} else if listName == "Spimex" {
		htmlhead += "</head><body><h1>Статистические материалы</h1>"
	} else if listName == "Acra" {
		htmlhead += "</head><body><h1>АКРА рейтинг: пресс-релизы</h1>"
	} else if listName == "Mintrans" {
		htmlhead += "</head><body><h1>Министерство транспорта Российской Федерации: новости</h1>"
	}
	if strings.Contains(listName, "add") {
		titleLink = "Перечень террористов и экстремистов (включённые)"
	} else if strings.Contains(listName, "del") {
		titleLink = "Перечень террористов и экстремистов (исключённые)"
	} else if listName == "Minjust" {
		titleLink = "Перечень нежелательных организаций"
	} else if listName == "Spimex" {
		titleLink = "Биржевая информация: статистические материалы"
	} else if listName == "Acra" {
		titleLink = "Пресс-релизы"
	} else if listName == "Mintrans" {
		titleLink = "Новости"
	}
	htmlhead += "<a href=\"" + urlList + "\">" + titleLink + "</a><br><br><br><div><ul>"
	headers := []byte(subject + adfrom + address + "Content-Type: text/html\nMIME-Version: 1.0\n\n" + htmlhead)
	htmlfooter := []byte("</ul></div></body></html>")

	var combined_string []byte
	if strings.Contains(listName, "UL") || strings.Contains(listName, "FL") {
		combined_string = []byte(strings.Join(newlist, "<br>"))
	} else {
		combined_string = []byte(strings.Join(newlist, "\n"))
	}
	headers = append(headers, combined_string...)
	msg := append(headers, htmlfooter...)
	sendmail := exec.Command("/usr/sbin/sendmail", "-t")
	stdin, err := sendmail.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}
	_, err = sendmail.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	sendmail.Start()
	stdin.Write([]byte(msg))
	stdin.Close()
	// sentBytes, _ := ioutil.ReadAll(stdout)
	sendmail.Wait()

}

func telega(newlist []string, listName, urlList, apikey string, chatList []string) {
	var title string
	var titleLink string
	var tgbody string
	var listbulk string
	var niter int

	const MAX_TG_SIZE = 4096

	if strings.Contains(listName, "UL") {
		title = "New list Federal Financial Monitoring Service: Юридические лица "
	} else if strings.Contains(listName, "FL") {
		title = "New list Federal Financial Monitoring Service: Физические лица "
	} else if listName == "Spimex" {
		title = "Биржевая информация: статистические материалы"
	} else if listName == "Acra" {
		title = "АКРА рейтинг"
	} else if listName == "Mintrans" {
		title = "Министерство транспорта Российской Федерации"
	}

	if (strings.Contains(listName, "UL") || strings.Contains(listName, "FL")) && strings.Contains(listName, "add") {
		title += "\\(включённые\\)"
	} else if (strings.Contains(listName, "UL") || strings.Contains(listName, "FL")) && strings.Contains(listName, "del") {
		title += "\\(исключённые\\)"
	}

	if strings.Contains(listName, "add") {
		titleLink = "Перечень террористов и экстремистов \\(включённые\\)"
	} else if strings.Contains(listName, "del") {
		titleLink = "Перечень террористов и экстремистов \\(исключённые\\)"
	} else if listName == "Minjust" {
		titleLink = "Перечень нежелательных организаций"
	} else if listName == "Spimex" {
		titleLink = "Биржевая информация: статистические материалы"
	} else if listName == "Acra" {
		titleLink = "Пресс\\-релизы"
	} else if listName == "Mintrans" {
		titleLink = "Новости"
	}
	tgbody = "*" + title + "*\n\n"
	tgbody += "[" + titleLink + "](" + urlList + ")\n\n"

	reli := regexp.MustCompile(`<.*?>`)
	resmb := regexp.MustCompile(`([_\*\[\]\(\)~\>\#\+\-\=\|\{\}\.!])`)

	for _, v := range newlist {
		v = reli.ReplaceAllString(v, "")
		v = resmb.ReplaceAllString(v, "\\$1")
		listbulk += "\\> " + v + "\n"
	}
	tgbody += listbulk
	// Проверяем, что тело сообщения не превышает предельный размер
	niter = len(tgbody) / MAX_TG_SIZE

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Send message to Telegram
	client := &http.Client{}

	url := "https://api.telegram.org/bot" + apikey + "/sendMessage"

	// Если тело сообщения не превышает предельный размер, то отсылаем сообщение как обычно
	if niter == 0 {
		for _, tgid := range chatList {
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
				fmt.Println("Title: ", title)
				fmt.Println("Title Link: ", titleLink)
				fmt.Println("URL link: ", urlList)
				fmt.Println("Listbulk:", listbulk)
			}

			defer resp.Body.Close()
		}
		listbulk = ""
	} else {
		// Если тело сообщения превышает предельный размер, то делим сообщение на чанки
		var divided [][]string
		chunkSize := len(newlist) / (niter + 1)
		for i := 0; i < len(newlist); i += chunkSize {
			end := i + chunkSize
			if end > len(newlist) {
				end = len(newlist)
			}
			divided = append(divided, newlist[i:end])
		}
		k := 0
		for _, dv := range divided {
			listbulk = ""
			if k == 0 {
				tgbody = "*" + title + "*\n\n"
				tgbody += "[" + titleLink + "](" + urlList + ")\n\n"
			} else {
				tgbody = "*" + title + " \\(Продолжение\\) *\n\n"
			}
			for _, vc := range dv {
				vc = reli.ReplaceAllString(vc, "")
				vc = resmb.ReplaceAllString(vc, "\\$1")
				listbulk += "\\> " + vc + "\n"
			}
			tgbody += listbulk
			for _, tgid := range chatList {
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
					fmt.Println("Chunk:", k)
					fmt.Println("tgbody:", tgbody)
					fmt.Println("resp.StatusCode:", resp.StatusCode)
				}
				defer resp.Body.Close()
				k++
			}
		}
	}
}

func main() {
	var userAgent string

	// Ключи для командной строки
	flag.StringVar(&userAgent, "uagent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit", "User Agent")

	flag.Parse()

	path, _ := os.Executable()
	path = path[:strings.LastIndex(path, "/")+1]

	var byteValue_list []byte

	type Configlist struct {
		List   string
		Emails []string
		Chats  []string
		Url    string
	}

	type Configtg struct {
		APIkey string
	}

	var configlist []Configlist
	var configtg Configtg
	// Читаем файл с адресами
	if _, err := os.Stat(path + "/emails.json"); err == nil {
		// Open our jsonFile
		byteValue, err := os.ReadFile(path + "/emails.json")
		// if we os.ReadFile returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		// defer the closing of our jsonFile so that we can parse it later on
		// var listHash []ArticleH
		err = json.Unmarshal(byteValue, &configlist)
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
	for key, el := range configlist {
		keystr := strconv.Itoa(key)
		byteValue_list = []byte{}
		if _, err := os.Stat(path + "/file_" + keystr + ".txt"); err == nil {
			// Open our jsonFile
			byteValue_list, err = os.ReadFile(path + "/file_" + keystr + ".txt")
			// if we os.ReadFile returns an error then handle it
			if err != nil {
				fmt.Println(err)
			}
		}

		body, statuscode, err := getHtmlPage(el.Url, userAgent)
		if err != nil || statuscode != 200 {
			fmt.Printf("Error getHtmlPage - %v\n", err)
			fmt.Printf("URL - %s\n", el.Url)
			fmt.Printf("Status Code - %d\n", statuscode)
		} else {
			// Получаем список
			var get_list []string
			if el.List == "ULadd" || el.List == "ULdel" {
				get_list = getListFsm(body, "russianUL")
			} else if el.List == "FLadd" || el.List == "FLdel" {
				get_list = getListFsm(body, "russianFL")
			} else if el.List == "Minjust" {
				get_list = getListMinjust(body)
			} else if el.List == "Spimex" {
				get_list = getListSpimex(body)
			} else if el.List == "Acra" {
				get_list = getListAcra(body)
			} else if el.List == "Mintrans" {
				get_list = getListMintrans(body)
			}
			combined_string := []byte(strings.Join(get_list, ""))
			if !testEq(byteValue_list, combined_string) {
				err := os.WriteFile(path+"/file_"+keystr+".txt", combined_string, 0666)
				fmt.Println("Update from URL: ", el.Url)
				if err != nil {
					fmt.Println("Error : ", err)
				}
				// Для определенных сайтов отправляем только новые строки
				if el.List == "ULadd" || el.List == "ULdel" || el.List == "FLadd" || el.List == "FLdel" {
					mail(get_list, el.List, el.Url, el.Emails)
					telega(get_list, el.List, el.Url, configtg.APIkey, el.Chats)
				} else if el.List == "Minjust" {
					if len(byteValue_list) > 0 {
						new_list := newList(get_list, byteValue_list, `[0-9]+ № [\d]+-[[\p{Cyrillic} ]+[\d.]+`, "asc")
						if len(new_list) > 0 {
							mail(new_list, el.List, el.Url, el.Emails)
							telega(new_list, el.List, el.Url, configtg.APIkey, el.Chats)
						}
					} else {
						if len(get_list) > 0 {
							mail(get_list, el.List, el.Url, el.Emails)
							telega(get_list, el.List, el.Url, configtg.APIkey, el.Chats)
						}
					}
				} else if el.List == "Spimex" {
					if len(byteValue_list) > 0 {
						new_list := newList(get_list, byteValue_list, `^[0-9]{2} \p{Cyrillic}{3} .{1} [0-9]{2}`, "desc")
						if len(new_list) > 0 {
							mail(new_list, el.List, el.Url, el.Emails)
							telega(new_list, el.List, el.Url, configtg.APIkey, el.Chats)
						}
					} else {
						if len(get_list) > 0 {
							mail(get_list, el.List, el.Url, el.Emails)
							telega(get_list, el.List, el.Url, configtg.APIkey, el.Chats)
						}
					}
				} else if el.List == "Acra" {
					if len(byteValue_list) > 0 {
						new_list := newList(get_list, byteValue_list, `.*?\d{1,2} \p{Cyrillic}{3} \d{4}`, "desc")
						if len(new_list) > 0 {
							mail(new_list, el.List, el.Url, el.Emails)
							telega(new_list, el.List, el.Url, configtg.APIkey, el.Chats)
						}
					} else {
						if len(get_list) > 0 {
							mail(get_list, el.List, el.Url, el.Emails)
							telega(get_list, el.List, el.Url, configtg.APIkey, el.Chats)
						}
					}
				} else if el.List == "Mintrans" {
					if len(byteValue_list) > 0 {
						new_list := newList(get_list, byteValue_list, `.*?[PAD]`, "desc")
						for i := 0; i < len(new_list); i++ {
							new_list[i] = strings.ReplaceAll(new_list[i], "[PAD]", "")
						}
						if len(new_list) > 0 {
							mail(new_list, el.List, el.Url, el.Emails)
							telega(new_list, el.List, el.Url, configtg.APIkey, el.Chats)
						}
					} else {
						for i := 0; i < len(get_list); i++ {
							get_list[i] = strings.ReplaceAll(get_list[i], "[PAD]", "")
						}
						if len(get_list) > 0 {
							mail(get_list, el.List, el.Url, el.Emails)
							telega(get_list, el.List, el.Url, configtg.APIkey, el.Chats)
						}
					}
				}
			}
		}
	}
}
