package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	API_KEY = "4b0d355121fe4e0bb3d86e902efe9f20"

	USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15"
	API_BASE   = "https://israelrail.azurefd.net/rjpa-prod/api/v1"
)

type Rail struct {
	Result Line1 `json:"result,omitempty"`
}
type Line1 struct {
	NumResults int     `json:"numOfResultsToShow,omitempty"`
	StartFrom  int     `json:"startFromIndex,omitempty"`
	Travels    []Train `json:"travels,omitempty"`
}
type Train struct {
	Trains []TrainRoutePart `json:"trains,omitempty"`
}
type TrainRoutePart struct {
	OrignStation       int    `json:"orignStation,omitempty"`
	DestinationStation int    `json:"destinationStation,omitempty"`
	ArrivalTime        string `json:"arrivalTime,omitempty"`
	DepartureTime      string `json:"departureTime,omitempty"`
	OriginPlatform     int    `json:"originPlatform,omitempty"`
	DestPlatform       int    `json:"destPlatform,omitempty"`
}

type Ans map[string]map[string]TrainRoutePartS

type TrainRoutePartS struct {
	OrignStation       string `json:"orignStation,omitempty"`
	DepartureTime      string `json:"departureTime,omitempty"`
	OriginPlatform     int    `json:"originPlatform,omitempty"`
	DestinationStation string `json:"destinationStation,omitempty"`
	ArrivalTime        string `json:"arrivalTime,omitempty"`
	DestPlatform       int    `json:"destPlatform,omitempty"`
}

var DEFAULT_HEADERS = url.Values{
	"User-Agent":                []string{USER_AGENT},
	"ocp-apim-subscription-key": []string{API_KEY},
}

type RailApi struct {
	url       string
	params    url.Values
	headers   url.Values
	arguments url.Values
}

func getRailSchedule(userName, from, to string) string {
	ans := getSchedule(userName, from, to)
	prettyJSON, err := json.MarshalIndent(ans, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return err.Error()
	}

	// Print the formatted JSON
	return string(prettyJSON)
}

func getSchedule(userName, from, to string) Ans {
	cache := NewCache()
	ans := cache.Get(userName, from, to)
	if ans == nil {
		params := url.Values{}
		params.Add("fromStation", from)
		params.Add("toStation", to)
		params.Add("date", time.Now().Format(time.DateOnly))
		params.Add("hour", time.Now().Format(time.TimeOnly))
		params.Add("scheduleType", "2")
		params.Add("systemType", "1")
		params.Add("languageId", "Hebrew")

		body := callRailAPI(params)

		parse, err := parseBody(body)
		if err != nil {
			log.Fatal(err)
		}
		ans = parse
		cache.Set(userName, from, to, ans)
	}
	return ans.(Ans)
}
func callRailAPI(params url.Values) []byte {
	fullUrl := fmt.Sprintf("%s?%s", "https://israelrail.azurefd.net/rjpa-prod/api/v1/timetable/searchTrainLuzForDateTime", params.Encode())
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("User-Agent", USER_AGENT)
	req.Header.Set("ocp-apim-subscription-key", API_KEY)

	// Send the request using http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	return body
}

func parseBody(body []byte) (Ans, error) {
	var rail Rail
	err := json.Unmarshal(body, &rail)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	// take only 5 relevant results from all day list
	rail.Result.Travels = rail.Result.Travels[rail.Result.StartFrom : rail.Result.StartFrom+rail.Result.NumResults]
	ans := make(map[string]map[string]TrainRoutePartS)
	for j, travel := range rail.Result.Travels {
		innerMap := make(map[string]TrainRoutePartS)

		for i, train := range travel.Trains {
			innerMap[fmt.Sprintf("🚂 %v", i)] = TrainRoutePartS{
				OrignStation:       getStation(train.OrignStation),
				DepartureTime:      extractTime(train.DepartureTime),
				OriginPlatform:     train.OriginPlatform,
				DestinationStation: getStation(train.DestinationStation),
				ArrivalTime:        extractTime(train.ArrivalTime),
				DestPlatform:       train.DestPlatform,
			}
		}
		ans[fmt.Sprintf("🚆- %v", j)] = innerMap
	}
	return ans, nil
}
func getStation(i int) string {
	if val, ok := STATIONS[strconv.Itoa(i)]["Heb"]; ok {
		return val
	}
	return ""
}

func extractTime(dateStr string) string {
	parsedTime, err := time.Parse("2006-01-02T15:04:05", dateStr)
	if err != nil {
		return "Invalid date format"
	}
	return parsedTime.Format("15:04:05")
}

func (r RailApi) RailApiCreate(url string, params url.Values, headers url.Values) {

	r.url = joinURL(url)
	r.params = params

	if len(headers) > 0 {
		r.headers = headers
	} else {
		r.headers = DEFAULT_HEADERS
	}
}

func joinURL(url string) string {
	// Trim any trailing slash from base and leading slash from url
	return API_BASE + "/" + url
}
