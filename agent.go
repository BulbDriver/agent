package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Bulb struct {
	Name       string `json:"name"`
	Color      string `json:"color"`
	Brightness uint64 `json:"brightness"`
}

func main() {
	hub := "http://localhost:9393"

	if len(os.Args) > 1 {
		hub = os.Args[1]
	}

	var bulb = &Bulb{
		"My Bulb",
		"ffffff", // Default to White
		100,  // Default to full brightness
	}

	http.HandleFunc("/heartbeat", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Ping\n")
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
		switch req.Header.Get("Content-Type") {
		case "application/json":
			jsonData, err := json.Marshal(bulb)
			if err != nil {
				log.Println(err)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, string(jsonData))
		default:
			w.Header().Set("Content-Type", "text/html")

			fmt.Fprintf(w, `<html>
<head>
	<title>Status</title>
	<style type="text/css">
		span#preview {
			width: 100px;
			height: 100px;
			border: 1px solid black;
			display: block;
			background-color: #%s;
		}
	</style>
</head>
<body>
<div id="name">Name: %q</div>
<div id="color">
	Color: %q
	Brightness: %d
	<span id="preview"></span>
</div>
</body>
</html>`, bulb.Color, bulb.Name, bulb.Color, bulb.Brightness)
		}
	})

	http.HandleFunc("/name", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "GET":
			// get the color
			fmt.Fprintf(w, "%q", bulb.Name)
		case "PATCH":
			// update the color
			nameValue := req.FormValue("name")

			if nameValue == "" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "{\"error\": \"invalid name\"}")
			} else {
				w.WriteHeader(http.StatusAccepted)
				bulb.Name = nameValue
			}
		}
	})

	http.HandleFunc("/color", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "GET":
			// get the color
			fmt.Fprintf(w, "%q", bulb.Color)
		case "PATCH":
			// update the color
			colorValue := req.FormValue("color")

			if colorValue == "" || len(colorValue) < 3 {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "{\"error\": \"invalid color\"}")
			} else {
				w.WriteHeader(http.StatusAccepted)
				bulb.Color = colorValue
			}
		}
	})

	http.HandleFunc("/brightness", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "GET":
			// get the color
			fmt.Fprintf(w, "%d", bulb.Brightness)
		case "PATCH":
			// update the color
			brightnessValue, err := strconv.ParseUint(req.FormValue("brightness"), 10, 8)

			if err == nil {
				w.WriteHeader(http.StatusAccepted)
				bulb.Brightness = brightnessValue
			} else {
				w.WriteHeader(http.StatusBadRequest)
				log.Print(err)
				fmt.Fprintln(w, "{\"error\": \"invalid brightness\"}")
			}
		}
	})

	for {
		log.Printf("Trying to connect Bulb to hub (%s) ...", hub)

		req, err := http.Head(hub)

		if err != nil || req.StatusCode != http.StatusOK {
			if req != nil {
				if req.StatusCode != http.StatusOK {
					log.Printf("Hub responded code %d instead of %d!", req.StatusCode, http.StatusOK)
				}
			}

			time.Sleep(2 * time.Second)
		} else {
			jsonData, _ := json.Marshal(bulb)

			_, err := http.Post(hub + "/register", "application/json", bytes.NewBuffer(jsonData))

			if err != nil {
				log.Fatalf("%v", err)
			}

			break
		}
	}

	log.Fatal(http.ListenAndServe(":9494", nil))
}
