package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const kCREATE_DATA_API = "/create-data"
const kREAD_DATA_API = "/get-data"
const kUPDATE_DATA_API = "/update-data"
const kDELETE_DATA_API = "/delete-data"

type sharedHandler struct {
	jsonFilePath string
	vecJsonMap   []DataJsonMap
	mux          sync.Mutex
}

func displaySummaryHelp() {
	fmt.Printf(
		"usage: %s serve [--address ADDRESS:PORT] [--file JSON_PATH]"+"\n"+
			"\n"+
			"Running the executable with undefined subcommands or arguments will print this page"+"\n"+
			"\n"+
			"SUBCOMMAND: serve"+"\n"+
			"The serve command starts the server using the default settings."+"\n"+
			"    --address is used to specify the IPv4 address for the server to listen to. (Default: localhost:80)"+"\n"+
			"    --file JSON_PATH is used to specify a JSON file-based database for simple persistence mechanism"+"\n"+
			"\n"+
			"JSON FILE SCHEMATICS"+"\n"+
			"The JSON file must follow the following scheme: \"[{id: <UINT>, value: <STRING>}, ...]\""+"\n"+
			"where the 'id' field signifies a unique identifier for a specified entry, and the 'value' field signifies the value of the entry."+"\n"+
			"\n"+
			"API USAGE"+"\n"+
			"	GET /create-data?id=<INT> returns a JSON data entry specified by the 'id' parameter."+"\n"+
			"Setting the id parameter to -1 will return all JSON entries"+"\n"+
			"\n"+
			"	PUT /create-data?value=<STRING> creates a data entry with the specified value."+"\n"+
			"\n"+
			"AUTHOR: Muhammad Raznan"+"\n",

		filepath.Base(os.Args[0]))
}

func findData(vecData []DataJsonMap, id int) (data DataJsonMap, found bool) {
	found = false

	for _, datum := range vecData {
		if id == datum.Id {
			data = datum
			found = true
			break
		}
	}

	return
}

func (sh *sharedHandler) handleCreateData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "PUT":
		var (
			reqDataValue string
			newData      DataJsonMap
		)

		qParams, err := url.ParseQuery(request.URL.RawQuery)

		if err != nil {
			http.Error(writer, ":(", http.StatusBadRequest)
			return
		}

		reqDataValue = qParams.Get("value")
		newData = DataJsonMap{Id: rand.Int(), Value: reqDataValue, LastModified: time.Now().UTC().Format(http.TimeFormat)}

		sh.mux.Lock()

		sh.vecJsonMap = append(sh.vecJsonMap, newData)

		jsonBytes, err := json.Marshal(sh.vecJsonMap)

		if err != nil {
			http.Error(writer, "Sorry :x", http.StatusInternalServerError)
			return
		}

		err = os.WriteFile(sh.jsonFilePath, jsonBytes, 0644)

		if err != nil {
			http.Error(writer, "Sorry :o", http.StatusInternalServerError)
			return
		}

		sh.mux.Unlock()

		writer.Header().Set("Last-Modified", newData.LastModified)
		writer.Header().Set("Content-Location", fmt.Sprintf("%s?id=%d", kREAD_DATA_API, newData.Id))
		writer.WriteHeader(http.StatusCreated)
	}
}

func (sh *sharedHandler) handleGetData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		var reqDataId int

		qParams, err := url.ParseQuery(request.URL.RawQuery)

		if err != nil {
			http.Error(writer, ":(", http.StatusBadRequest)
			return
		}

		reqDataId, err = strconv.Atoi(qParams.Get("id"))

		if err != nil {
			http.Error(writer, ":-(", http.StatusBadRequest)
			return
		}

		if reqDataId == -1 {
			json.NewEncoder(writer).Encode(sh.vecJsonMap)
		} else {
			responseData, ok := findData(sh.vecJsonMap, reqDataId)

			if !ok {
				http.Error(writer, ":O", http.StatusBadRequest)
				return
			}

			json.NewEncoder(writer).Encode(responseData)
		}
	}
}

func (sh *sharedHandler) handleUpdateData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		// Get data
	}
}

func (sh *sharedHandler) handleDeleteData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "DELETE":
		// Get data
	}
}

func runServe(sh *sharedHandler, ipaddr, port string) {
	http.HandleFunc(kCREATE_DATA_API, sh.handleCreateData)
	http.HandleFunc(kREAD_DATA_API, sh.handleGetData)
	http.HandleFunc(kUPDATE_DATA_API, sh.handleUpdateData)
	http.HandleFunc(kDELETE_DATA_API, sh.handleDeleteData)

	if err := http.ListenAndServe(ipaddr+":"+port, nil); err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
}

func main() {
	if len(os.Args) < 2 {
		displaySummaryHelp()
		return
	}

	if os.Args[1] == "serve" {
		var (
			words      []string
			ipaddr     string
			port       string
			portAsInt  int
			err        error
			vecJsonMap []DataJsonMap
			jsonBytes  []byte
			shandler   sharedHandler
		)

		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		ipAddrPtr := serveCmd.String("address", "localhost:80", "is used to specify the IPv4 address for the server to listen to. (Default: localhost:80)")
		jsonPathPtr := serveCmd.String("file", "file.json", "JSON_PATH is used to specify a JSON file-based database for simple persistence mechanism")

		if len(os.Args) > 2 {
			serveCmd.Parse(os.Args[2:])

			if os.Args[2] == "help" {
				serveCmd.Usage()
				os.Exit(1)
			}
		}

		words = strings.Split(*ipAddrPtr, ":")

		if len(words) != 2 {
			log.Fatalf("ERROR: %s is not a valid ADDRESS:PORT\n", os.Args[4])
		}

		if words[0] != "localhost" && net.ParseIP(words[0]) == nil {
			log.Fatalf("ERROR: %s is not a valid address\n", words[0])
		}

		portAsInt, err = strconv.Atoi(words[1])

		if err != nil {
			log.Fatalf("ERROR: %s is not a valid port number\n", words[1])
		}

		if portAsInt < 0 && portAsInt >= 65535 {
			log.Fatalf("ERROR: the port %s is not IANA standard port\n", words[1])
		}

		_, err = os.Stat(*jsonPathPtr)

		if os.IsNotExist(err) {
			log.Printf("ERROR: %s does not exist on the current directory, creating...\n", *jsonPathPtr)

			vecJsonMap = []DataJsonMap{CreateDataJson("Some value :3")}

			jsonBytes, err = json.Marshal(vecJsonMap)

			if err != nil {
				log.Fatalf("ERROR: failed to marshal. Reason %s\n", err.Error())
			}

			err = os.WriteFile(*jsonPathPtr, jsonBytes, 0644)

			if err != nil {
				log.Fatalf("ERROR: failed writing to %s. Reason %s\n", *jsonPathPtr, err.Error())
			}
		} else {
			jsonBytes, err = os.ReadFile(*jsonPathPtr)

			if err != nil {
				log.Fatalf("ERROR: failed to read file %s. Reason %s\n", *jsonPathPtr, err.Error())
			}

			err = json.Unmarshal(jsonBytes, &vecJsonMap)

			if err != nil {
				log.Fatalf("ERROR: failed to unmarshal. Reason %s\n", err.Error())
			}
		}

		ipaddr = words[0]
		port = words[1]
		shandler.jsonFilePath = *jsonPathPtr
		shandler.vecJsonMap = vecJsonMap

		runServe(&shandler, ipaddr, port)
	} else {
		displaySummaryHelp()
	}
}
