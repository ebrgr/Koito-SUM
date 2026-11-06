// Inventory records capabilities of the actual local native build, not a universal promise.
package main

import (
	"encoding/json"
	"github.com/h2non/bimg"
	"os"
)

func main() {
	out := map[string]any{}
	for kind, name := range bimg.ImageTypes {
		out[name] = bimg.IsImageTypeSupportedByVips(kind)
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
