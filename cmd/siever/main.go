package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wwmoraes/siever/pkg/managesieve"
)

var errLog = log.New(os.Stderr, "", log.LstdFlags)

func assert(err error) {
	if err == nil {
		return
	}

	errLog.Println(err)
	os.Exit(1)
}

func main() {
	SERVER := os.Getenv("SIEVER_SERVER")
	USERNAME := os.Getenv("SIEVER_USERNAME")
	PASSWORD := os.Getenv("SIEVER_PASSWORD")
	PORT := os.Getenv("SIEVER_PORT")
	if PORT == "" {
		PORT = "4190"
	}

	fmt.Println("creating client")
	client, err := managesieve.NewClient(
		fmt.Sprintf("%s:%s", SERVER, PORT),
		managesieve.WithLogger(log.New(os.Stdout, "DEBUG: ", 0)),
	)
	assert(err)
	defer client.Close()

	fmt.Println("starting TLS")
	message, err := client.StartTLS()
	assert(err)
	fmt.Println(message)

	fmt.Println("checking capabilites post-TLS")
	message, err = client.Capability()
	assert(err)
	fmt.Println(message)

	fmt.Println("logging in")
	response, err := client.Login(USERNAME, PASSWORD)
	assert(err)
	defer client.Logout() //nolint:errcheck // no need to check err on defer
	fmt.Println(response)

	fmt.Println("checking capabilites post-login")
	message, err = client.Capability()
	assert(err)
	fmt.Println(message)

	fmt.Println("listing scripts")
	scripts, response, err := client.ListScripts()
	assert(err)
	fmt.Println(response)
	fmt.Println(scripts)

	fmt.Printf("%#+v\n", scripts)

	fmt.Println("checking space for test script")
	response, err = client.HaveSpace("test3", 4096)
	assert(err)
	fmt.Printf("%#+v\n", response)

	fmt.Println("DONE")
}
