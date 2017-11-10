package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	r := []byte(`[["a", 1, "c"]]`)
	var arr [][]interface{}
	err := json.Unmarshal(r, &arr)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+#v", arr[0])
	fmt.Println(time.Now())
}
