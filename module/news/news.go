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
		return
	}
	defer response.Body.Close()
	//var temp2 []weatherData
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		fmt.Println("Error decoding reddit access token resonse:", err)
		fmt.Println("dump:", response)
	} else {
		//fmt.Println("dump:", jsonResponse)
		//fmt.Println("key is: ", jsonResponse["access_token"], " expires in: ", jsonResponse["expires_in"])
		redditAccessTokenTemp, ok := jsonResponse["access_token"].(string)
		if !ok {
			fmt.Println("Error with reddit access token")
		} else {
			redditAccessToken = redditAccessTokenTemp
			thisTime := time.Now()
			redditAccessTokenExpiryTemp, ok := jsonResponse["expires_in"].(float64)
			if !ok {
				fmt.Println("Error with reddit access token expiry time")
			} else {
				redditAccessTokenExpiry = thisTime.Unix() + int64(redditAccessTokenExpiryTemp)
				//fmt.Println("key is: ", redditAccessToken, " expires at: ", redditAccessTokenExpiry)
			}
		}
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
	if err != nil {
		fmt.Println("Error creating news request:", err)
	}
	req.Header.Add("Authorization", "bearer "+redditAccessToken)
	req.Header.Set("user-agent", "newsweather/0.1")
	response, err := netClient.Do(req)
	if err != nil {
		fmt.Println("Error executing news request:", err)
		return
	}
	defer response.Body.Close()
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		fmt.Println("Error decoding news response:", err)
		fmt.Println("dump:", response)
		return
	}
	//fmt.Println("response:", jsonResponse)
	tdb1, ok := jsonResponse["data"].(map[string]interface{})
	if !ok {
		fmt.Println("Error with news response data:", err)
		fmt.Println("dump:", jsonResponse["data"])
		return
	}
	//fmt.Println("tdb1 success")
	tdb2, ok := tdb1["children"].([]interface{})
	if !ok {
		fmt.Println("Error with news response data children:", err)
		fmt.Println("dump:", tdb1["children"])
		return
	}
	fmt.Println("tdb2 success, len:", len(tdb2))
	for i := 0; i < 25; i++ {
		tdb3, ok := tdb2[i].(map[string]interface{})
		if !ok {
			fmt.Println("Error with news response data tdb3:", err)
			fmt.Println("dump:", tdb1)
		} else {
			tdb4, ok := tdb3["data"].(map[string]interface{})
			if !ok {
				fmt.Println("Error with news response data tdb4:", err)
				fmt.Println("dump:", tdb1)
			} else {
				if tdb4["stickied"] == false {
					HeadlineDB = append(HeadlineDB, tdb4["title"].(string))
				}
			}
		}
	}
	fmt.Println("headlinedb len: ", len(HeadlineDB))
	if len(HeadlineDB) < 24 {
		HeadlineDB = HeadlineDB[:25]
		fmt.Println("first news grab, getting first 25")
	} else {
		HeadlineDB = HeadlineDB[25:]
	}
	fmt.Println("got news")
}
