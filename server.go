package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
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
			"    --address is used to specify the address for the server to listen to. (default \"localhost:80\")"+"\n"+
			"    --file JSON_PATH is used to specify a JSON file-based database for simple persistence mechanism. If the file path does not exists, it will create one, assuming the directory is present. (default \"file.json\")"+"\n"+
			"\n"+
			"JSON FILE SCHEMATICS"+"\n"+
			"The JSON file must follow the following scheme: \"[{id: <UINT>, name: <STRING>, value: <STRING>, last_modified: <STRING>}, ...]\""+"\n"+
			"where the 'id' field signifies a unique identifier for a specified entry, and the 'value' field signifies the value of the entry."+"\n"+
			"\n"+
			"API USAGE"+"\n"+
			"	PUT /create-data?name=<STRING>[&value=<STRING>] is used to create a new data entry. The 'name' parameter defines the entry, and it must be unique. An optional 'value' parameter can also be set. If not, the entry will have an emptry string value. The response status 201 indicates successful entry creation. The response status 400 will be returned for malformed queries."+"\n"+
			"\n"+
			"	GET /get-data?<id=<INT>|name=<STRING>> returns a JSON data entry specified by the 'id' parameter. Setting 'id' -1 will return all JSON entries. You can also retrieve the data entry by 'name'. You can't use both 'id' and 'name' in conjunction. The response status 200 indicates succesful entry retrieval, which carries the message as JSON format. The response status 400 will be returned for malformed queries."+"\n"+
			"\n"+
			"	POST /update-data?name=<STRING>[&value=<STRING>] updates the specified entry value. The 'name' parameters specifies the target entry, and an optional 'value' can be set, if not the value will be an empty string. The status code 204 indicates successful modification. The response status 400 will be returned for malformed queries and undefined entries."+"\n"+
			"\n"+
			"	DELETE /delete-data?<id=<INT>|name=<STRING>> deletes the specified entry. A 'name' or 'id' parameter can be used to specify the target entry, but it can't be used in conjunction. Setting 'id' to -1 will delete all entries an returns a null message. Successful deletion will return status 204. Malformed queries will return 400 response status."+"\n"+
			"\n"+
			"AUTHOR: Muhammad Raznan"+"\n",

		filepath.Base(os.Args[0]))
}

func findData(vecData []DataJsonMap, id int) (data *DataJsonMap, found bool) {
	var datum *DataJsonMap
	found = false

	for i := range vecData {
		datum = &vecData[i]

		if id == datum.Id {
			data = datum
			found = true
			break
		}
	}

	return
}

func findDataByName(vecData []DataJsonMap, name string) (data *DataJsonMap, found bool) {
	var datum *DataJsonMap
	found = false

	for i := range vecData {
		datum = &vecData[i]

		if name == datum.Name {
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
			reqDataName  string
			reqData      *DataJsonMap
			haveData     bool
		)

		qParams, err := url.ParseQuery(request.URL.RawQuery)

		if err != nil {
			http.Error(writer, ":(", http.StatusBadRequest)
			return
		}

		if len(qParams) == 0 || len(qParams) > 2 {
			http.Error(writer, ":-(", http.StatusBadRequest)
			return
		}

		if len(qParams) == 2 && !qParams.Has("value") {
			http.Error(writer, ":)", http.StatusBadRequest)
			return
		}

		if !qParams.Has("name") {
			http.Error(writer, ":?(", http.StatusBadRequest)
			return
		}

		reqDataName = qParams.Get("name")
		reqData, haveData = findDataByName(sh.vecJsonMap, reqDataName)

		if !haveData {
			reqDataValue = qParams.Get("value")
			newData := CreateDataJson(reqDataName, reqDataValue)
			reqData = &newData

			sh.mux.Lock()

			sh.vecJsonMap = append(sh.vecJsonMap, *reqData)

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
		}

		writer.Header().Set("Content-Location", fmt.Sprintf("%s?id=%d", kREAD_DATA_API, reqData.Id))
		writer.Header().Set("Last-Modified", reqData.LastModified)
		writer.WriteHeader(http.StatusCreated)
	}
}

func (sh *sharedHandler) handleGetData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		var (
			resVecData []DataJsonMap
			haveData   bool
		)

		qParams, err := url.ParseQuery(request.URL.RawQuery)

		if err != nil {
			http.Error(writer, ":(", http.StatusBadRequest)
			return
		}

		if len(qParams) != 1 {
			http.Error(writer, ":?<", http.StatusBadRequest)
			return
		}

		if qParams.Has("id") {
			var reqDataId int

			reqDataId, err = strconv.Atoi(qParams.Get("id"))

			if err != nil {
				http.Error(writer, ":-(", http.StatusBadRequest)
				return
			}

			if reqDataId == -1 {
				resVecData = sh.vecJsonMap
				haveData = true
			} else {
				var resData *DataJsonMap
				resData, haveData = findData(sh.vecJsonMap, reqDataId)
				resVecData = []DataJsonMap{*resData}
			}
		} else if qParams.Has("name") {
			var (
				reqDataName string
				resData     *DataJsonMap
			)

			reqDataName = qParams.Get("name")
			resData, haveData = findDataByName(sh.vecJsonMap, reqDataName)
			resVecData = []DataJsonMap{*resData}
		} else {
			http.Error(writer, "3:?<", http.StatusBadRequest)
			return
		}

		if !haveData {
			http.Error(writer, ":O", http.StatusBadRequest)
			return
		}

		json.NewEncoder(writer).Encode(resVecData)
	}
}

func (sh *sharedHandler) handleUpdateData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		var (
			reqDataValue string
			reqDataName  string
			editedData   *DataJsonMap
			haveData     bool
		)

		qParams, err := url.ParseQuery(request.URL.RawQuery)

		if err != nil {
			http.Error(writer, ":(", http.StatusBadRequest)
			return
		}

		if !qParams.Has("name") && !qParams.Has("value") {
			http.Error(writer, ":(", http.StatusBadRequest)
			return
		}

		reqDataName = qParams.Get("name")
		reqDataValue = qParams.Get("value")

		editedData, haveData = findDataByName(sh.vecJsonMap, reqDataName)

		if !haveData {
			http.Error(writer, ":O", http.StatusBadRequest)
			return
		}

		sh.mux.Lock()
		editedData.Value = reqDataValue
		editedData.LastModified = time.Now().UTC().Format(http.TimeFormat)
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

		writer.WriteHeader(http.StatusNoContent)
	}
}

func (sh *sharedHandler) handleDeleteData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "DELETE":
		var (
			newVecData  []DataJsonMap
			deletedData *DataJsonMap
			canDelete   bool
		)

		qParams, err := url.ParseQuery(request.URL.RawQuery)

		if err != nil {
			http.Error(writer, ":(", http.StatusBadRequest)
			return
		}

		if len(qParams) != 1 {
			http.Error(writer, ":?<", http.StatusBadRequest)
			return
		}

		if qParams.Has("id") {
			reqDataId, err := strconv.Atoi(qParams.Get("id"))

			if err != nil {
				http.Error(writer, ":-(", http.StatusBadRequest)
				return
			}

			if reqDataId == -1 {
				sh.vecJsonMap = nil
				canDelete = true
			} else {
				deletedData, canDelete = findData(sh.vecJsonMap, reqDataId)
			}
		} else if qParams.Has("name") {
			reqDataName := qParams.Get("name")
			deletedData, canDelete = findDataByName(sh.vecJsonMap, reqDataName)
		}

		if !canDelete {
			http.Error(writer, ":<", http.StatusBadRequest)
			return
		}

		for i := range sh.vecJsonMap {
			datum := &sh.vecJsonMap[i]

			if datum == deletedData {
				continue
			}

			newVecData = append(newVecData, *datum)
		}

		sh.mux.Lock()
		sh.vecJsonMap = newVecData

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

		writer.WriteHeader(http.StatusNoContent)
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
		ipAddrPtr := serveCmd.String("address", "localhost:80", "used to specify the IPv4 address and port for the server to listen to.")
		jsonPathPtr := serveCmd.String("file", "file.json", "accepts a JSON_PATH used to specify a JSON file-based database for simple persistence mechanism.")

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

		if portAsInt > 65535 && portAsInt < 0 {
			log.Fatalf("ERROR: the port %s is not IANA standard port\n", words[1])
		}

		_, err = os.Stat(*jsonPathPtr)

		if os.IsNotExist(err) {
			log.Printf("ERROR: %s does not exist on the current directory, creating...\n", *jsonPathPtr)

			vecJsonMap = []DataJsonMap{CreateDataJson("Auriga", "Some value :3")}

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
