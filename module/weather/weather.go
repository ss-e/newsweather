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

const (
	weatherSite      = "https://api.openweathermap.org/data/2.5/"
	delayCurrentTemp = 20
	delay6hrTemp     = 6
	delayBatch       = 30
)

func debugOutput(t string) {
	debug.Output("weather", t)
	return
}

//ReadWeatherDB return weatherdb
func ReadWeatherDB() []Data {
	return weatherDB
}

func schedule(f func(*http.Client), interval time.Duration, nc *http.Client) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			f(nc)
		}
	}()
	return ticker
}

//Startup performs actions to be done on startup, start timers
func Startup(nc *http.Client) error {
	readToDB("weather")
	getCurrentTemp(nc)
	go get6hrTemp(nc)
	t1 := schedule(getCurrentTemp, delayCurrentTemp*time.Minute, nc)
	_ = t1
	t2 := schedule(get6hrTemp, delay6hrTemp*time.Hour, nc)
	_ = t2
	return nil
}

//ReadToDB read cities in database
func readToDB(dbname string) {
	// open json file
	jsonFile, err := ioutil.ReadFile("./db/" + dbname + ".json")
	if err != nil {
		debugOutput("error reading db: " + err.Error())
		return
	}
	err = json.Unmarshal(jsonFile, &weatherDB)
	if err != nil {
		debugOutput("error unmarshalling db: " + err.Error())
		return
	}
	return
}

//GetCurrentTemp poll database entries for current temperature
func getCurrentTemp(nc *http.Client) {
	//we can do batches of 20 requests at a time in order to limit HTTP overhead
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
	//leftovers
	temp = append(temp, t2)
	//send batches
	for i := 0; i < len(temp); i++ {
		debugOutput("Loading map temperature batch: " + fmt.Sprintf("%d", i+1) + "/" + fmt.Sprintf("%d", len(temp)))
		//var url = weatherSite + "group?id=" + strings.Join(temp[i], ",") + "&units=metric&appid=" + weatherAPIKey
		//response, err := nc.Get(url)
		response, err := nc.Get(weatherSite + "group?id=" + strings.Join(temp[i], ",") + "&units=metric&appid=" + weatherAPIKey)
		if err != nil {
			debugOutput("Error in getCurrentTemp get request: " + err.Error())
			continue
		}
		defer response.Body.Close()
		if response.Body == nil {
			debugOutput("Did not recieve a response from server.")
			return
		}
		var jsonResponse map[string]interface{}
		err2 := json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err2 != nil {
			debugOutput("error decoding getCurrentTemp: " + err.Error())
			continue
		} else {
			responseArr, ok := jsonResponse["list"].([]interface{})
			if !ok {
				debugOutput("error decoding response from getcurrenttemp")
				continue
				/*message, ok := jsonResponse["message"].([]interface{})
				if !ok {
					debugOutput("error decoding response from getcurrenttemp, response: " + response)
				} else {
					debugOutput("error decoding response from getcurrenttemp with message:" + message)
				}*/
			} else {
				for j := 0; j < int(jsonResponse["cnt"].(float64)); j++ {
					//extract information
					temp2 := responseArr[j].(map[string]interface{})
					t3 := temp2["main"].(map[string]interface{})
					t4 := temp2["weather"].([]interface{})
					t5 := t4[0].(map[string]interface{})
					id := int(temp2["id"].(float64))
					nowtemp := int(math.Round(t3["temp"].(float64)))
					nowid := int(t5["id"].(float64))
					//apply information to weatherDB
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
		//once we get batch, wait delayBatch seconds for API to be ready again
		debugOutput("got batch, waiting " + fmt.Sprintf("%d", delayBatch) + " seconds")
		time.Sleep(time.Second * delayBatch)
		debugOutput("done sleeping")
	}
	debugOutput("getCurrentTemp completed, waiting " + fmt.Sprintf("%d", delayCurrentTemp) + " minutes.")
	return
}

func get6hrTemp(nc *http.Client) {
	for i := range weatherDB {
		time.Sleep(5 * time.Second)
		//var url = weatherSite + "onecall?lat=" + weatherDB[i].Lat + "&lon=" + weatherDB[i].Lon + "&exclude=minutely,current&units=metric&appid=" + weatherAPIKey
		//var url = weatherSite + "forecast?id=" + weatherDB[i].ID + "&appid=" + weatherAPIKey + "&units=metric&cnt=19"
		//response, err := nc.Get(url)
		response, err := nc.Get(weatherSite + "forecast?id=" + weatherDB[i].ID + "&appid=" + weatherAPIKey + "&units=metric&cnt=19")
		if err != nil {
			debugOutput("err getting 6hr temp data: " + err.Error())
			continue
		}
		defer response.Body.Close()
		if response.Body == nil {
			debugOutput("Did not recieve a response from server.")
			return
		}
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			debugOutput("error decoding get6hrTemp response: " + err.Error())
			continue
		} else {
			responseArr, ok := jsonResponse["list"].([]interface{})
			if !ok {
				debugOutput("error decoding response from get6hrTemp")
				continue
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
				debugOutput("grabbed 6hr index for item: " + fmt.Sprintf("%d", i))
			}
		}
	}
	debugOutput("get6hrTemp completed, waiting " + fmt.Sprintf("%d", delay6hrTemp) + " minutes.")
	return
}
