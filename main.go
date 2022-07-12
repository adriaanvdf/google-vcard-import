package main

import (
        "context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"

	"vdfeltz.com/workspace/google-vcard-import/client"
)

func main() {
	myClient := client.GetClient("credentials.json", people.ContactsReadonlyScope)

	peopleService, err := people.NewService(context.Background(), option.WithHTTPClient(myClient))
	if err != nil {
		log.Fatalf("Unable to create people Client %v", err)
	}

	results, err := peopleService.People.Connections.List("people/me").PageSize(10).
		PersonFields("names,emailAddresses").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve people. %v", err)
	}
	if len(results.Connections) > 0 {
		fmt.Print("List 10 connection names:\n")
		for _, c := range results.Connections {
			names := c.Names
			if len(names) > 0 {
				name := names[0].DisplayName
				fmt.Printf("%s\n", name)
			}
		}
	} else {
		fmt.Print("No connections found.")
	}
}
