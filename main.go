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
	"strings"
)

func main() {
	//
	logger.Infof("Start big-stamp server.")
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

type profile struct {
	Ok      bool `json:"ok"`
	Members []struct {
		ID       string `json:"id"`
		TeamID   string `json:"team_id"`
		Name     string `json:"name"`
		Deleted  bool   `json:"deleted"`
		Color    string `json:"color"`
		RealName string `json:"real_name"`
		Tz       string `json:"tz"`
		TzLabel  string `json:"tz_label"`
		TzOffset int    `json:"tz_offset"`
		Profile  struct {
			AvatarHash            string `json:"avatar_hash"`
			StatusText            string `json:"status_text"`
			StatusEmoji           string `json:"status_emoji"`
			RealName              string `json:"real_name"`
			DisplayName           string `json:"display_name"`
			RealNameNormalized    string `json:"real_name_normalized"`
			DisplayNameNormalized string `json:"display_name_normalized"`
			Email                 string `json:"email"`
			Image24               string `json:"image_24"`
			Image32               string `json:"image_32"`
			Image48               string `json:"image_48"`
			Image72               string `json:"image_72"`
			Image192              string `json:"image_192"`
			Image512              string `json:"image_512"`
			Team                  string `json:"team"`
		} `json:"profile"`
		IsAdmin           bool `json:"is_admin"`
		IsOwner           bool `json:"is_owner"`
		IsPrimaryOwner    bool `json:"is_primary_owner"`
		IsRestricted      bool `json:"is_restricted"`
		IsUltraRestricted bool `json:"is_ultra_restricted"`
		IsBot             bool `json:"is_bot"`
		Updated           int  `json:"updated"`
		IsAppUser         bool `json:"is_app_user,omitempty"`
		Has2Fa            bool `json:"has_2fa"`
	} `json:"members"`
	CacheTs          int `json:"cache_ts"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

func handle(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("SLASHCOMMAND")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}

	channelID := r.Form.Get("channel_id")
	texts := strings.Split(r.Form.Get("text"), " ")
	userID := r.Form.Get("user_id")

	option := url.Values{}
	option.Add("token", token)

	usersURL := "https://slack.com/api/users.list?"
	usersOption := url.Values{}
	usersOption.Add("token", token)
	resp, err := http.Get(usersURL + usersOption.Encode())
	if err != nil {
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}
	var users profile
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error parsing form.", http.StatusBadRequest)
		return
	}
	json.Unmarshal(b, &users)
	username := map[string]string{}
	usersicon := map[string]string{}
	for _, m := range users.Members {
		username[m.ID] = m.Profile.DisplayName
		usersicon[m.ID] = m.Profile.Image512
	}

	stampMap := map[string]bool{}
	for _, text := range texts {
		stampMap[text] = false
	}
	sendMsgUrl := "https://slack.com/api/chat.postMessage"

	emojiMsg := emojiList(w, token)
	for k, imgUrl := range emojiMsg {
		ek := fmt.Sprintf(":%s:", k)
		if _, ok := stampMap[ek]; ok {
			option := url.Values{}
			option.Add("token", token)
			option.Add("channel", channelID)
			option.Add("attachments", "[{\"\": \"\", \"text\": \"\", \"image_url\": \""+imgUrl.(string)+"\"}]")
			option.Add("as_user", "false")
			option.Add("username", username[userID])
			option.Add("icon_url", usersicon[userID])
			_, err := http.Post(sendMsgUrl+"?"+option.Encode(), "", nil)
			if err != nil {
				http.Error(w, "can't http post.", http.StatusBadRequest)
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
