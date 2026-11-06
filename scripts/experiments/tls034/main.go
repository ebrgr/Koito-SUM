// Command tls034 checks Go's HTTPS trust in the disposable issue 034 container.
package main

import (
	"fmt"
	"net/http"
	"time"
)

// main performs an HTTPS request using Go's default system certificate pool.
func main() {
	client := http.Client{Timeout: 15 * time.Second}
	response, err := client.Get("https://example.com/")
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		panic(response.Status)
	}
	fmt.Println("TLS verified")
}
