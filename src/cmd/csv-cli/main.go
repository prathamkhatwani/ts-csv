package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	csv "github.com/prathamkhatwani/ts-csv/src"
)

type Request struct {
	FieldSeparator string `json:"fieldSeparator"`
	LineSeparator  string `json:"lineSeparator"`
	Quote          string `json:"quote"`
	Text           string `json:"text"`
}

type Response struct {
	Rows  [][]string          `json:"rows,omitempty"`
	JSON  []map[string]string `json:"json,omitempty"`
	Error string              `json:"error,omitempty"`
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		sendError("Failed to read stdin: " + err.Error())
		return
	}

	var req Request
	if err := json.Unmarshal(input, &req); err != nil {
		sendError("Failed to parse request JSON: " + err.Error())
		return
	}

	config := csv.Configuration{
		FieldSeparator: req.FieldSeparator,
		LineSeparator:  req.LineSeparator,
		Quote:          req.Quote,
	}

	parser := csv.NewParser(config)
	rows, err := parser.Parse(req.Text)

	if err != nil {
		sendError(err.Error())
		return
	}

	res := Response{
		Rows: rows,
		JSON: parser.JSON(),
	}

	out, err := json.Marshal(res)
	if err != nil {
		sendError("Failed to marshal response JSON: " + err.Error())
		return
	}

	fmt.Print(string(out))
}

func sendError(msg string) {
	res := Response{Error: msg}
	out, _ := json.Marshal(res)
	fmt.Print(string(out))
	os.Exit(0)
}
