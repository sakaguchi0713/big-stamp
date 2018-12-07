package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"github.com/lob-inc/rssp/server/shared/logger"
)

var (
	port  string = "80"
	token string
)

func main() {
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}

	//channel := r.Form.Get("channel")
	emojiListUrlOption := url.Values{}
	token := os.Getenv("BIGSTAMP")
	//token := "xoxp-382633581043-383781965799-496956374679-271a7ddf620b6f39538939b7f8e74465"
	emojiListUrlOption.Add("token", token)
	resp, err := http.Post("https://slack.com/api/emoji.list"+"?"+emojiListUrlOption.Encode(), "", nil)
	if err != nil {
		http.Error(w, "Error get emoji.list.", http.StatusBadRequest)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error don't response read.", http.StatusBadRequest)
	}

	var result interface{}
	if err = json.Unmarshal(b, &result); err != nil {
		http.Error(w, "Error parse json.", http.StatusBadRequest)
	}
	msg := result.(map[string]interface{})

	var emoji interface{}
	eb, _ := json.Marshal(msg["emoji"])

	json.Unmarshal(eb, &emoji)
	if emoji == nil {
		logger.Warn("emoji in json is nil.")
	}

	emojiMsg := emoji.(map[string]interface{})
	sendMsgUrl := "https://slack.com/api/chat.postMessage"

	text := r.Form.Get("text")
	for k, imgUrl := range emojiMsg {
		if strings.Contains(text, k) {
			sendMsgUrlOption := url.Values{}
			sendMsgUrlOption.Add("token", token)
			sendMsgUrlOption.Add("channel", "#random")
			sendMsgUrlOption.Add("attachments", "[{\"\": \"\", \"text\": \"\", \"image_url\": \""+imgUrl.(string)+"\"}]")
			_, err := http.Post(sendMsgUrl+"?"+sendMsgUrlOption.Encode(), "", nil)
			if err != nil {
				http.Error(w, "Error parse json.", http.StatusBadRequest)
			}
		}
	}
}
