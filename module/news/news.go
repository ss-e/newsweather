package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

//HeadlineDB contains all headlines
var HeadlineDB []string
var redditAccessToken string
var redditAccessTokenExpiry int64

//Startup starts authentication and headline scheduling
func Startup() error {
	redditOAuth()
	getCurrentHeadlines()
	t1 := schedule(getCurrentHeadlines, 3*time.Minute)
	_ = t1
	return nil
}

func schedule(f func(), interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			f()
		}
	}()
	return ticker
}

//ReadHeadlineDB return weatherdb
func ReadHeadlineDB() []string {
	//return []string{"test","2222"}
	return HeadlineDB
}

func redditOAuth() {
	url := "https://api.reddit.com/api/v1/access_token"
	post := strings.NewReader("grant_type=password&username=newsweather&password=u8uNcbQWtzmeWhDgRK8v")
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", url, post)
	req.SetBasicAuth("vhIMcPFb0A2OHA", "lCLwLSPVOKsDFQhIYgVvQy-7Ta4")
	req.Header.Set("user-agent", "newsweather/0.1")
	response, err := netClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	//var temp2 []weatherData
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		fmt.Println("error:", err)
		fmt.Println("dump:", response)
	} else {
		//fmt.Println("dump:", jsonResponse)
		//fmt.Println("key is: ", jsonResponse["access_token"], " expires in: ", jsonResponse["expires_in"])
		redditAccessToken = jsonResponse["access_token"].(string)
		thisTime := time.Now()
		redditAccessTokenExpiry = thisTime.Unix() + int64(jsonResponse["expires_in"].(float64))
		//fmt.Println("key is: ", redditAccessToken, " expires at: ", redditAccessTokenExpiry)
	}
}

//getCurrentHeadlines poll for headlines
func getCurrentHeadlines() {
	fmt.Println("key is: ", redditAccessToken, " expires at: ", redditAccessTokenExpiry)
	var url = "https://oauth.reddit.com/r/worldnews/hot.json?limit=25"
	thisTime := time.Now()
	//fmt.Println("comparing", thisTime.Unix(), "and", redditAccessTokenExpiry)
	if thisTime.Unix() > redditAccessTokenExpiry {
		fmt.Println("OAuth is expired, renewing")
		redditOAuth()
		return
	}
	//fmt.Println("passed check")
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "bearer "+redditAccessToken)
	req.Header.Set("user-agent", "newsweather/0.1")
	response, err := netClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		fmt.Println("error:", err)
		fmt.Println("dump:", response)
	} else {
		//fmt.Println("response:", jsonResponse)
		tdb1 := jsonResponse["data"].(map[string]interface{})
		//fmt.Println("tdb1 success")
		tdb2 := tdb1["children"].([]interface{})
		//fmt.Println("tdb2 success, len:", len(tdb2))
		for i := 0; i < 25; i++ {
			tdb3 := tdb2[i].(map[string]interface{})
			tdb4 := tdb3["data"].(map[string]interface{})
			if tdb4["stickied"] == false {
				HeadlineDB = append(HeadlineDB, tdb4["title"].(string))
			}
		}
		HeadlineDB = HeadlineDB[:25]
	}
}
