package seedco_test

import (
	"log"

	"github.com/orijtech/seedco"
)

func Example_client_AuthToken() {
	client, err := seedco.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	token, err := client.AuthToken("seedco-5ecR3t-username", "seedco-p4S5W0r6")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Token: %#v\n", token)
}

func Example_client_RefreshToken() {
	client, err := seedco.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	token, err := client.RefreshToken("2.a.gxDCXy_JTcGPFYb8M6zVgg.coXKLXuO8JWzSZKZ0TaQGKZMyY8Pkk-vSX7RgimNubE.YyWuB6YlCiSp1jPQEyv71EDpZbFxCgsWTaxdcqNsExw")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("RefreshedToken: %#v\n", token)
}

func Example_client_APIVersion() {
	client, err := seedco.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	apiVersion, err := client.APIVersion()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Latest APIVersion: %+v\n", apiVersion)
}
