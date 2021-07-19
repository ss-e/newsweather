package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"../debug"
)

//Data contains critical data for weather
type Data struct {
	Name string
	ID   string
	Tz   string
	Lat  string
	Lon  string
	Now  [2]int
	W    [3][2]int
}

var weatherDB []Data
var weatherAPIKey = os.Getenv("OWM_APIKEY")
var weatherSite string = "https://api.openweathermap.org/data/2.5/"

func debugOutput(t string) {
	debug.Output("weather", t)
}

//ReadWeatherDB return weatherdb
func ReadWeatherDB() []Data {
	return weatherDB
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

//Startup performs actions to be done on startup, start timers
func Startup() error {
	readToDB("weather")
	getCurrentTemp()
	go get6hrTemp()
	t1 := schedule(getCurrentTemp, 20*time.Minute)
	_ = t1
	t2 := schedule(get6hrTemp, 6*time.Hour)
	_ = t2
	return nil
}

//ReadToDB read cities in database
func readToDB(dbname string) {
	// open json file
	jsonFile, err := ioutil.ReadFile("./db/" + dbname + ".json")
	if err != nil {
		debugOutput("error reading db: " + err.Error())
	}
	err = json.Unmarshal(jsonFile, &weatherDB)
	if err != nil {
		debugOutput("error unmarshalling db: " + err.Error())
	}
}

//GetCurrentTemp poll database entries for current temperature
func getCurrentTemp() {
	temp := make([][]string, 0)
	t2 := make([]string, 0)
	for k := 0; k < len(weatherDB); k++ {
		if len(t2) >= 20 {
			temp = append(temp, t2)
			t2 = nil
			t2 = make([]string, 0)
			t2 = append(t2, weatherDB[k].ID)
		} else {
			t2 = append(t2, weatherDB[k].ID)
		}
	}
	temp = append(temp, t2)
	for i := 0; i < len(temp); i++ {
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		debugOutput("loading map temperature batch:" + fmt.Sprintf("%d", i+1) + "/" + fmt.Sprintf("%d", len(temp)))
		var url = weatherSite + "group?id=" + strings.Join(temp[i], ",") + "&units=metric&appid=" + weatherAPIKey
		response, err := netClient.Get(url)
		if err != nil {
			debugOutput("Error getcurrenttemp()" + err.Error())
			continue
		}
		defer response.Body.Close()
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			debugOutput("error decoding getCurrentTemp:" + err.Error())
		} else {
			responseArr, ok := jsonResponse["list"].([]interface{})
			if !ok {
				debugOutput("error decoding response from getcurrenttemp")
				/*message, ok := jsonResponse["message"].([]interface{})
				if !ok {
					debugOutput("error decoding response from getcurrenttemp, response: " + response)
				} else {
					debugOutput("error decoding response from getcurrenttemp with message:" + message)
				}*/
			} else {
				for j := 0; j < int(jsonResponse["cnt"].(float64)); j++ {
					temp2 := responseArr[j].(map[string]interface{})
					t3 := temp2["main"].(map[string]interface{})
					t4 := temp2["weather"].([]interface{})
					t5 := t4[0].(map[string]interface{})
					id := int(temp2["id"].(float64))
					nowtemp := int(math.Round(t3["temp"].(float64)))
					nowid := int(t5["id"].(float64))
					for index := range weatherDB {
						t0, _ := strconv.Atoi(weatherDB[index].ID)
						if t0 == id {
							weatherDB[index].Now[0] = nowtemp
							weatherDB[index].Now[1] = nowid
							break
						}
					}
				}
			}
		}
		debugOutput("got batch, waiting 30 seconds")
		time.Sleep(30 * time.Second)
		debugOutput("done sleeping")
	}
}

func get6hrTemp() {
	for i := range weatherDB {
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		time.Sleep(5 * time.Second)
		//var url = weatherSite + "onecall?lat=" + weatherDB[i].Lat + "&lon=" + weatherDB[i].Lon + "&exclude=minutely,current&units=metric&appid=" + weatherAPIKey
		var url = weatherSite + "forecast?id=" + weatherDB[i].ID + "&appid=" + weatherAPIKey + "&units=metric&cnt=19"
		response, err := netClient.Get(url)
		if err != nil {
			debugOutput("err getting 6hr temp data:" + err.Error())
			continue
		}
		defer response.Body.Close()
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			debugOutput("error decoding get6hrTemp response:" + err.Error())
		} else {
			responseArr, ok := jsonResponse["list"].([]interface{})
			if !ok {
				debugOutput("error decoding response from get6hrTemp")
				/*message, ok2 := jsonResponse["message"].([]interface{})
				if !ok2 {
					debugOutput("error decoding response for 6 hour temp for index: "+fmt.Sprintf("%d",i)+" message dump: ", response, ",", jsonResponse)
				} else {
					debugOutput("error decoding response for 6 hour temp for index: "+fmt.Sprintf("%d",i)+" with message", message)
				}*/
			} else {
				nowHour := time.Now().Hour()
				h := 6 - (nowHour % 6)
				k := 0
				for j := h; j < 19; j = j + 6 {
					//get main temp
					main := responseArr[j].(map[string]interface{})
					t1 := main["main"].(map[string]interface{})
					weatherDB[i].W[k][0] = int(t1["temp"].(float64))
					//get weather status
					t2 := main["weather"].([]interface{})
					t3 := t2[0].(map[string]interface{})
					weatherDB[i].W[k][1] = int(t3["id"].(float64))
					k++
				}
				//debugOutput("weather 6hr index:" + fmt.Sprintf("%d", i) + "w1:" + fmt.Sprintf("%d", weatherDB[i].W[0][0]) + "," + fmt.Sprintf("%d", weatherDB[i].W[0][1]) + "w2:" + fmt.Sprintf("%d", weatherDB[i].W[1][0]) + "," + fmt.Sprintf("%d", weatherDB[i].W[1][1]) + "w3:" + fmt.Sprintf("%d", weatherDB[i].W[2][0]) + "," + fmt.Sprintf("%d", weatherDB[i].W[2][1]))
				debugOutput("grabbed 6hr index for item:" + fmt.Sprintf("%d", i))
			}
		}
	}
}
