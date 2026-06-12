package main

import (
	"encoding/json"
	"fmt"

	"github.com/sgdevelopers29-afk/GoKit-Lite/response"
)

func main() {

	user := map[string]string{
		"name": "Ganesh",
	}

	resp := response.Success(user)

	b, _ := json.MarshalIndent(resp, "", "  ")

	fmt.Println(string(b))
}
