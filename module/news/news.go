package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"../debug"
)

//HeadlineDB contains all headlines
var HeadlineDB []string
var redditUsername string = os.Getenv("REDDIT_USERNAME")
var redditPassword string = os.Getenv("REDDIT_PASSWORD")
var redditAppUsername string = os.Getenv("REDDIT_APP_USERNAME")
var redditAppSecret string = os.Getenv("REDDIT_APP_SECRET")
var redditAccessToken string
var redditAccessTokenExpiry int64

func debugOutput(t string) {
	debug.Output("news", t)
}

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

//ReadHeadlineDB return HeadlineDB
func ReadHeadlineDB() []string {
	//return []string{"test","2222"}
	return HeadlineDB
}

//redditOAuth regenerate reddit oauth token
func redditOAuth() {
	url := "https://api.reddit.com/api/v1/access_token"
	post := strings.NewReader("grant_type=password&username=" + redditUsername + "&password=" + redditPassword)
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", url, post)
	req.SetBasicAuth(redditAppUsername, redditAppSecret)
	req.Header.Set("user-agent", "newsweather/0.1")
	response, err := netClient.Do(req)
	if err != nil {
		debugOutput("Oauth post errror: " + err.Error())
		return
	}
	defer response.Body.Close()
	//var temp2 []weatherData
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		debugOutput("Error decoding reddit access token response:" + err.Error())
	} else {
		//debugOutput("key is: " + jsonResponse["access_token"] + " expires in: " + jsonResponse["expires_in"])
		redditAccessTokenTemp, ok := jsonResponse["access_token"].(string)
		if !ok {
			debugOutput("Error with reddit access token")
		} else {
			redditAccessToken = redditAccessTokenTemp
			thisTime := time.Now()
			redditAccessTokenExpiryTemp, ok := jsonResponse["expires_in"].(float64)
			if !ok {
				debugOutput("Error with reddit access token expiry time")
			} else {
				redditAccessTokenExpiry = thisTime.Unix() + int64(redditAccessTokenExpiryTemp)
				//debugOutput("key is: ", redditAccessToken, " expires at: ", redditAccessTokenExpiry)
			}
		}
	}
}

//getCurrentHeadlines poll for headlines
func getCurrentHeadlines() {
	//debugOutput("key is: " + redditAccessToken + " expires at: " + redditAccessTokenExpiry)
	var url = "https://oauth.reddit.com/r/worldnews/hot.json?limit=25"
	thisTime := time.Now()
	if thisTime.Unix() > redditAccessTokenExpiry {
		debugOutput("OAuth is expired, renewing")
		redditOAuth()
		return
	}
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		debugOutput("Error creating news request:" + err.Error())
	}
	req.Header.Add("Authorization", "bearer "+redditAccessToken)
	req.Header.Set("user-agent", "newsweather/0.1")
	response, err := netClient.Do(req)
	if err != nil {
		debugOutput("Error executing news request:" + err.Error())
		return
	}
	defer response.Body.Close()
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		debugOutput("Error decoding news response:" + err.Error())
		return
	}
	tdb1, ok := jsonResponse["data"].(map[string]interface{})
	if !ok {
		//debugOutput("Error with news response data:" + err.Error() + " dump: " + jsonResponse["data"].(string))
		debugOutput("Error with news response data:" + err.Error())
		return
	}
	tdb2, ok := tdb1["children"].([]interface{})
	if !ok {
		//debugOutput("Error with news response data children:" + err.Error() + "dump:" + tdb1["children"].(string))
		debugOutput("Error with news response data children:" + err.Error())
		return
	}
	//debugOutput("Got temp db2, len:" + fmt.Sprintf("%d", len(tdb2)))
	for i := 0; i < 25; i++ {
		tdb3, ok := tdb2[i].(map[string]interface{})
		if !ok {
			debugOutput("Error with news response data tdb3:" + err.Error())
		} else {
			tdb4, ok := tdb3["data"].(map[string]interface{})
			if !ok {
				debugOutput("Error with news response data tdb4:" + err.Error())
			} else {
				if tdb4["stickied"] == false {
					HeadlineDB = append(HeadlineDB, tdb4["title"].(string))
				}
			}
		}
	}
	if len(HeadlineDB) <= 25 {
		debugOutput("first news grab, getting first 25")
		HeadlineDB = HeadlineDB[:25]
	} else {
		debugOutput("headlinedb len: " + fmt.Sprintf("%d", len(HeadlineDB)))
		HeadlineDB = HeadlineDB[25:]
	}
}
