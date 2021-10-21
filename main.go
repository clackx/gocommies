package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/antonholmquist/jason"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/goodsign/monday"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type config struct {
	TgToken string `json:"tgToken"`
	DBPath  string `json:"dbPath"`
	ChatId  int64  `json:"chatId"`
}

type people struct {
	bday  string
	bdate string
	wde   string
	cate  string
	name  string
	photo string
	rank  int
}

type peoples = []people

type details struct {
	cate string
	rank int
}

type orderedMap struct {
	m    map[string]details
	keys []string
}

func main() {
	config := getConfig()

	bot, err := tgbotapi.NewBotAPI(config.TgToken)
	if err != nil {
		log.Panic(err)
	}
	//bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	for {
		tnow := time.Now().Format("15:04")
		if tnow == "09:00" {
			bday := time.Now().Format("01.02")
			peoples := getRed(bday, config.DBPath)
			dayPersons := newOrderedMap()
			for _, p := range peoples {
				dayPersons.set(p.name, details{p.cate, p.rank})
			}
			for k := range dayPersons.keys {
				text := makeText(dayPersons, k)
				msg := tgbotapi.NewMessage(config.ChatId, text)
				msg.ParseMode = "html"
				msg.DisableWebPagePreview = false
				msg.DisableNotification = true
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(err)
				}
				if len(dayPersons.keys) > 10 {
					time.Sleep(time.Second * 30 * 60)
				} else {
					time.Sleep(time.Second * 60 * 60)
				}
			}
		}
		time.Sleep(time.Second * 30)
	}
}

func newOrderedMap() orderedMap {
	return orderedMap{make(map[string]details), []string{}}
}

func (o *orderedMap) set(k string, v details) {
	data, present := o.m[k]
	if !present {
		o.keys = append(o.keys, k)
		v.cate = "Категория: " + v.cate
	} else {
		v.cate = data.cate + "; " + v.cate
	}
	o.m[k] = v
}

func getConfig() config {
	data, err := ioutil.ReadFile("conf.json")
	if err != nil {
		panic(err)
	}
	var config config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
	}
	return config
}

func getRed(bday string, dbPath string) peoples {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT * FROM commies WHERE bday='%s' ORDER BY rank DESC", bday)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var peoples peoples
	for rows.Next() {
		p := people{}
		err := rows.Scan(&p.bday, &p.bdate, &p.wde, &p.cate, &p.name, &p.photo, &p.rank)
		if err != nil {
			fmt.Println(err)
			continue
		}
		peoples = append(peoples, p)
	}
	return peoples
}

func getInfo(name string) string {
	link := fmt.Sprintf("https://ru.wikipedia.org/w/api.php?action=query&"+
		"format=json&prop=extracts&explaintext=1&exintro=1&titles=%s", url.PathEscape(name))

	response, err := http.Get(link)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	b, err := io.ReadAll(response.Body)
	rJson, _ := jason.NewObjectFromBytes(b)
	qJson, _ := rJson.GetObject("query")
	pJson, _ := qJson.GetObject("pages")

	result := ""
	for _, element := range pJson.Map() {
		gJson, _ := element.Object()
		result, _ = gJson.GetString("extract")
	}
	return result
}

func makeText(dayPersons orderedMap, i int) string {
	name := dayPersons.keys[i]
	rank := dayPersons.m[name].rank
	catg := dayPersons.m[name].cate
	stars := "☆✫✮✭★✯"
	bars := stars[:len(fmt.Sprintf("%d", rank))*3]
	bdlc := monday.Format(time.Now(), "2 January", monday.LocaleRuRU)
	log.Printf("%s %s", name, bars)

	info := getInfo(name)
	info = strings.Join(strings.Split(strings.TrimRight(info, "\n"), "\n"), "\n\n")
	href := fmt.Sprintf("<a href='https://ru.wikipedia.org/wiki/%s'>%s</a>",
		strings.ReplaceAll(name, " ", "_"), name)

	text := fmt.Sprintf("<i>%s</i> родился товарищ <b>%s</b> %s\n\n%s\n\n<i>%s</i>\n",
		bdlc, href, bars, info, catg)
	return text
}
