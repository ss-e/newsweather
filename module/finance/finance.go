package finance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"../debug"
)

//Item db, chartdata is [timestamp, open, high, low, close]
type Item struct {
	Name          string
	Ticker        string
	Value         float64
	Open          float64
	ChangePercent float64
	Chartdata     [][]float64
}

//StockDB contains all stock info
var StockDB []Item

//FxDB contains all foreign currency info
var FxDB []Item

//CryptoDB contains all crypto info
var CryptoDB []Item

var iexapikey = os.Getenv("IEX_APIKEY")
var iexsite = "https://cloud.iexapis.com/"
var cryptoapi = "https://api.cryptowat.ch/markets/binance/"

//ReadStockDB return weatherdb
func ReadStockDB() []Item {
	return StockDB
}

//ReadFxDB return weatherdb
func ReadFxDB() []Item {
	return FxDB
}

//ReadCryptoDB return weatherdb
func ReadCryptoDB() []Item {
	return CryptoDB
}

func debugOutput(t string) {
	debug.Output("finance", t)
}

func getFxInfo() {
	//thisTime := time.Now()
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	ta := []string{}
	for i := range FxDB {
		ta = append(ta, FxDB[i].Ticker)
	}
	temp := strings.Join(ta, ",")
	req, err := http.NewRequest("GET", iexsite+"stable/fx/latest?symbols="+temp+"&token="+iexapikey, nil)
	req.Header.Set("user-agent", "newsweather/0.1")
	response, err := netClient.Do(req)
	if err != nil {
		debugOutput("Error getting fxinfo: " + err.Error())
	}
	defer response.Body.Close()
	var jsonResponse []map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		debugOutput("Error decoding json: " + err.Error())
	} else {
		for i := range jsonResponse {
			ticker, ok := jsonResponse[i]["symbol"].(string)
			if ok {
				for j := range FxDB {
					if ticker == FxDB[j].Ticker {
						value, ok := jsonResponse[i]["rate"].(float64)
						//debugOutput("for: " + ticker + " got value: " + value)
						if ok {
							FxDB[j].Value = value
						} else {
							debugOutput("error copying response in getfxinfo")
						}
						break
					}
				}
			}
		}
	}
}

func getStockInfo() {
	for i := range StockDB {
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", iexsite+"stable/stock/"+StockDB[i].Ticker+"/book?token="+iexapikey, nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			debugOutput("err getting stock data:" + err.Error())
			continue
		}
		defer response.Body.Close()
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			debugOutput("Error decoding getStockInfo() json: " + err.Error())
		} else {
			tdb1, ok := jsonResponse["quote"].(map[string]interface{})
			if !ok {
				debugOutput("Error response to getstockinfo for ticker item " + StockDB[i].Ticker)
			} else {
				value, ok := tdb1["latestPrice"].(float64)
				open, ok := tdb1["open"].(float64)
				changepercent, ok := tdb1["changePercent"].(float64)
				debugOutput("for: " + StockDB[i].Ticker + " got value: " + fmt.Sprintf("%f", value) + " open: " + fmt.Sprintf("%f", open) + " changepercent: " + fmt.Sprintf("%f", changepercent))
				if ok {
					StockDB[i].Value = value
					StockDB[i].Open = open
					StockDB[i].ChangePercent = changepercent * 100
				} else {
					debugOutput("error copying response in getstockinfo for item " + StockDB[i].Ticker)
				}
			}
		}
	}
}

func getStockChartData() {
	for i := range StockDB {
		debugOutput("getStockChartData for: " + StockDB[i].Ticker)
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", iexsite+"stable/stock/"+StockDB[i].Ticker+"/intraday-prices?chartInterval=5&token="+iexapikey, nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			debugOutput("err getStockChartData:" + err.Error())
			continue
		}
		defer response.Body.Close()
		var jsonResponse []interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			debugOutput("Error decoding getStockChartData json: " + err.Error())
		} else {
			StockDB[i].Chartdata = nil
			for j := range jsonResponse {
				tdb1, ok := jsonResponse[j].(map[string]interface{})
				if !ok {
					debugOutput("error response to chartdatainfo for ticker item " + StockDB[i].Ticker)
				} else {
					temp, err := time.Parse("2006-01-02 15:04", tdb1["date"].(string)+" "+tdb1["minute"].(string))
					if err != nil {
						//debugOutput("for " + StockDB[i].Ticker + ", error parsing stock chart data, date " + err.Error())
					}
					date := temp.Unix()
					open, err3 := tdb1["open"].(float64)
					if !err3 {
						//debugOutput("for " + StockDB[i].Ticker + ", error parsing stock chart data, date " + err.Error())
					}
					high, err4 := tdb1["high"].(float64)
					if !err4 {
						//debugOutput("for " + StockDB[i].Ticker + ", error parsing stock chart data, date " + err.Error())
					}
					low, err5 := tdb1["low"].(float64)
					if !err5 {
						//debugOutput("for " + StockDB[i].Ticker + ", error parsing stock chart data, date " + err.Error())
					}
					close, err6 := tdb1["close"].(float64)
					if !err6 {
						//debugOutput("for " + StockDB[i].Ticker + ", error parsing stock chart data, date " + err.Error())
					}
					if err == nil && err3 && err4 && err5 && err6 {
						//debugOutput("inputting entry. [date:", float64(date)*1000, ",open:", open, ",high:", high, ",low:", low, ",close:", close)
						entry := []float64{float64(date) * 1000, open, high, low, close}
						StockDB[i].Chartdata = append(StockDB[i].Chartdata, entry)
					}
				}
			}
		}
	}
}

func getCryptoInfo() {
	for i := range CryptoDB {
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", cryptoapi+CryptoDB[i].Ticker+"/summary", nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			debugOutput("err getting getCryptoInfo http:" + err.Error())
			continue
		}
		defer response.Body.Close()
		//if we put too many requests across, stop immediately
		if response.StatusCode == 429 {
			debugOutput("recieved 429 code doing getCryptoInfo" + err.Error())
			return
		}
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			debugOutput("Error decoding getCryptoInfo json: " + err.Error())
		} else {
			tdb1, ok := jsonResponse["result"].(map[string]interface{})
			if !ok {
				debugOutput("Finance error response to getCryptoInfo for ticker item " + CryptoDB[i].Ticker)
			} else {
				tdb2 := tdb1["price"].(map[string]interface{})
				tdb3 := tdb2["change"].(map[string]interface{})
				value, ok := tdb2["last"].(float64)
				changepercent, ok := tdb3["percentage"].(float64)
				if ok {
					CryptoDB[i].Value = value
					CryptoDB[i].ChangePercent = changepercent * 100
					debugOutput("for: " + CryptoDB[i].Ticker + " got value: " + fmt.Sprintf("%f", CryptoDB[i].Value) + " changepercent: " + fmt.Sprintf("%f", CryptoDB[i].ChangePercent))
				} else {
					debugOutput("error copying getCryptoInfo response")
				}
			}
		}
	}
}

func getCryptoChartData() {
	for i := range CryptoDB {
		t1 := time.Now()
		t2 := t1.Add(-24 * time.Hour)
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		thisTime := fmt.Sprintf("%v", t2.Unix())
		debugOutput("getting crypto chart data for: " + CryptoDB[i].Ticker + " for time: " + thisTime)
		req, err := http.NewRequest("GET", cryptoapi+CryptoDB[i].Ticker+"/ohlc?periods=1800&after="+thisTime, nil)
		req.Header.Set("user-agent", "newsweather/0.1")
		response, err := netClient.Do(req)
		if err != nil {
			debugOutput("err getting crypto chart data: " + err.Error())
			continue
		}
		defer response.Body.Close()
		//if we put too many requests across, stop immediately
		if response.StatusCode == 429 {
			debugOutput("recieved 429 code doing getCryptoChartData" + err.Error())
			return
		}
		var jsonResponse map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			debugOutput("Error decoding getCryptoChartData json: " + err.Error())
		} else {
			CryptoDB[i].Chartdata = nil
			tdb1, ok := jsonResponse["result"].(map[string]interface{})
			if !ok {
				debugOutput("error response to crypto chartdata for ticker item " + CryptoDB[i].Ticker)
			} else {
				tdb2, ok := tdb1["1800"].([]interface{})
				if !ok {
					debugOutput("error response to 1800 for ticker item " + CryptoDB[i].Ticker)
				} else {
					for j := range tdb2 {
						tdb3, ok := tdb2[j].([]interface{})
						if ok {
							date, ok2 := tdb3[0].(float64)
							open, ok3 := tdb3[1].(float64)
							high, ok4 := tdb3[2].(float64)
							low, ok5 := tdb3[3].(float64)
							close, ok6 := tdb3[4].(float64)
							if ok2 && ok3 && ok4 && ok5 && ok6 {
								//debugOutput("inputting entry. [date:", float64(date)*1000, ",open:", open, ",high:", high, ",low:", low, ",close:", close)
								entry := []float64{float64(date) * 1000, open, high, low, close}
								CryptoDB[i].Chartdata = append(CryptoDB[i].Chartdata, entry)
							} else {
								debugOutput("error 2 parsing crypto chart data for ticker item " + CryptoDB[i].Ticker)
							}
						} else {
							debugOutput("error 1 parsing crypto chart data for ticker item " + CryptoDB[i].Ticker)
						}
					}
				}
			}
		}
	}
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

//ReadToDB read tickers to check
func readToDB(dbname string, database *[]Item) {
	jsonFile, err := ioutil.ReadFile("./db/" + dbname + ".json")
	if err != nil {
		debugOutput("Error reading db file" + err.Error())
	}
	err = json.Unmarshal(jsonFile, &database)
	if err != nil {
		debugOutput("Error unmarshalling db file" + err.Error())
	}
}

//Startup starts authentication and headline scheduling
func Startup() error {
	readToDB("stock", &StockDB)
	readToDB("crypto", &CryptoDB)
	getStockInfo()
	getStockChartData()
	getCryptoInfo()
	getCryptoChartData()
	t1 := schedule(getStockInfo, 3*time.Minute)
	_ = t1
	t2 := schedule(getStockChartData, 15*time.Minute)
	_ = t2
	t3 := schedule(getCryptoInfo, 5*time.Minute)
	_ = t3
	t4 := schedule(getCryptoChartData, 15*time.Minute)
	_ = t4
	return nil
}
