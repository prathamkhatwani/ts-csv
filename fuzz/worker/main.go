package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	csv "github.com/prathamkhatwani/ts-csv/src"
)

type Request struct {
	InputBase64 string `json:"input_b64"`
}

type Response struct {
	OK    bool       `json:"ok"`
	Rows  [][]string `json:"rows,omitempty"`
	Error string     `json:"error,omitempty"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	// Set large buffer size for long input lines
	const maxBuf = 10 * 1024 * 1024
	buf := make([]byte, maxBuf)
	scanner.Buffer(buf, maxBuf)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			resp := Response{OK: false, Error: fmt.Sprintf("invalid json req: %v", err)}
			out, _ := json.Marshal(resp)
			fmt.Println(string(out))
			continue
		}

		data, err := base64.StdEncoding.DecodeString(req.InputBase64)
		if err != nil {
			resp := Response{OK: false, Error: fmt.Sprintf("invalid b64: %v", err)}
			out, _ := json.Marshal(resp)
			fmt.Println(string(out))
			continue
		}

		input := string(data)
		parser := csv.NewParser()
		rows, parseErr := parser.Parse(input)

		if parseErr != nil {
			resp := Response{OK: false, Error: parseErr.Error()}
			out, _ := json.Marshal(resp)
			fmt.Println(string(out))
		} else {
			jsonRows := make([][]string, len(rows))
			for i, r := range rows {
				if r == nil {
					jsonRows[i] = []string{}
				} else {
					jsonRows[i] = r
				}
			}
			resp := Response{OK: true, Rows: jsonRows}
			out, _ := json.Marshal(resp)
			fmt.Println(string(out))
		}
	}
}
