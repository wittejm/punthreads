package oldcode

import (
	"encoding/json"
	"fmt"

	//"io"
	//"net/http"
	"math"
	"os"
	"strings"
)

func ParseJSONFile(jsonFile *os.File) (map[string]interface{}, error) {
	fmt.Println("in Parse_json_file")

	var data map[string]interface{}

	// Decode the JSON data directly from the file
	decoder := json.NewDecoder(jsonFile)
	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func makeArrayOfArrays() {
	// bonus: using a closure function!
	Pic := func(dx, dy int) [][]uint8 {
		img := make([][]uint8, dy)
		for i, _ := range img {
			img[i] = make([]uint8, dx)
			for j, _ := range img[i] {
				img[i][j] = uint8(math.Pow(float64(i), 1.2) + math.Pow(float64(j), 1.2))
			}
		}
		// fmt.Println(img)
		return img
	}
	Pic(3, 4)

}

// Check it out!
func WordCount(s string) map[string]int {
	fmt.Println("fields", strings.Fields(s))
	result := make(map[string]int)
	for _, f := range strings.Fields(s) {
		if _, ok := result[f]; ok {
			result[f] = result[f] + 1
		} else {
			result[f] = 1
		}
	}
	return result
}
