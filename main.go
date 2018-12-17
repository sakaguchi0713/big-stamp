package main

import (
	"encoding/json"
	"fmt"
	"github.com/lob-inc/rssp/server/shared/logger"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	logger.Infof("Start big-stamp server.")
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("SLASHCOMMAND")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}
	channelID := r.Form.Get("channel_id")
	text := r.Form.Get("text")

	sendMsgUrl := "https://slack.com/api/chat.postMessage"

	emojiMsg := emojiList(w, token)
	for k, imgUrl := range emojiMsg {
		if fmt.Sprintf(":%s:", k) == text {
			sendMsgUrlOption := url.Values{}
			sendMsgUrlOption.Add("token", token)
			sendMsgUrlOption.Add("channel", channelID)
			sendMsgUrlOption.Add("attachments", "[{\"\": \"\", \"text\": \"\", \"image_url\": \""+imgUrl.(string)+"\"}]")
			_, err := http.Post(sendMsgUrl+"?"+sendMsgUrlOption.Encode(), "", nil)
			if err != nil {
				http.Error(w, "Error parse json.", http.StatusBadRequest)
			}
		}
	}
}

func emojiList(w http.ResponseWriter, token string) (map[string]interface{}) {
	urlOption := url.Values{}
	urlOption.Add("token", token)

	url := "https://slack.com/api/emoji.list?" + urlOption.Encode()
	resp, err := http.Post(url, "", nil)
	if err != nil {
		http.Error(w, "can not get emoji list.", http.StatusBadRequest)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "cnan not response read.", http.StatusBadRequest)
	}

	var result interface{}
	if err = json.Unmarshal(b, &result); err != nil {
		http.Error(w, "Error parse json.", http.StatusBadRequest)
	}
	msg := result.(map[string]interface{})

	var emoji interface{}
	eb, err := json.Marshal(msg["emoji"])
	if err != nil {
		logger.Errorf("can not emoji marshal err: %v", err)
		http.Error(w, "can not emoji.list.", http.StatusBadRequest)
	}

	if eb == nil {
		logger.Error("not set emoji.")
	}

	json.Unmarshal(eb, &emoji)
	emojiMsg := emoji.(map[string]interface{})

	return emojiMsg
}
