package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func displayHelp() {

	fmt.Printf(
		"usage: %s [serve [--address ADDRESS:PORT] [--file JSON_PATH]" 												+ "\n"		+
																													  "\n" 		+
		"Running the executable with undefined subcommands or arguments will print this page" 						+ "\n"		+
																													  "\n"		+
		"SUBCOMMAND: serve"																							+ "\n"		+
		"The serve command starts the server using the default settings."											+ "\n"		+
		"    --address is used to specify the IPv4 address for the server to listen to. (Default: localhost:80)"	+ "\n"		+
		"    --file JSON_PATH is used to specify a JSON file-based database for simple persistence mechanism"		+ "\n"		+
																													  "\n"		+
		"AUTHOR: Muhammad Raznan"																					+ "\n",

		filepath.Base(os.Args[0]))
}

func handleCreateData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		// Create data
	}
}

func handleGetData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		// Get data
	}
}

func handleUpdateData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		// Get data
	}
}

func handleDeleteData(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "DELETE":
		// Get data
	}
}

func main() {
	displayHelp()

	http.HandleFunc("/create-data", handleCreateData)
	http.HandleFunc("/get-data", handleGetData)
	http.HandleFunc("/update-data", handleUpdateData)
	http.HandleFunc("/delete-data", handleDeleteData)
}
