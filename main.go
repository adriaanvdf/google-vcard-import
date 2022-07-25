package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/emersion/go-vcard"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"

	"google-vcard-import/client"
)

func main() {
	card, err := readVcardFromFile("/Users/adriaan/workspace/google-vcard-import/example.vcf")
	fmt.Printf("card: %v\n", card)
	if err != nil {
		log.Fatal(err)
	}	
	contact := parseCardToPerson(card)

	

	fmt.Printf("contact: %v\n", contact)

	myClient := client.GetClient("credentials.json", people.ContactsScope)

	peopleService, err := people.NewService(context.Background(), option.WithHTTPClient(myClient))
	if err != nil {
		log.Fatalf("Unable to create people Client %v", err)
	}

	results3, err := peopleService.People.CreateContact(&contact).Do()
	if err != nil {
		log.Fatalf("Unable to create contact. %v", err)
	}
	fmt.Printf("result.HTTPStatusCode: %v\n", results3.HTTPStatusCode)
	x := people.CreateContactGroupRequest{ContactGroup: &people.ContactGroup{Name: "Created: " + time.Now().Format(time.Kitchen)}}
	results1, err := peopleService.ContactGroups.Create(&x).Do()
	if err != nil {
		log.Fatalf("Unable to create contacts label. %v", err)
	}
	fmt.Printf("results1: %v\n", results1)

}

func createSingleContact(peopleService *people.Service, contact *people.Person) (*people.Person, error) {
	results, err := peopleService.People.CreateContact(contact).Do()
	if err != nil {
		log.Fatalf("Unable to create contact. %v", err)
	}
	return results, err
}

func listConnections(peopleService *people.Service) ([]*people.Person, error) {
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
		fmt.Print("No connections found.\n")
	}
	return results.Connections, err
}

func parseCardToPerson(card vcard.Card) people.Person {
	var formattedName = card.PreferredValue(vcard.FieldFormattedName)
	var birthDayDate = parseVcardDate(card.PreferredValue(vcard.FieldBirthday))
	var gender = card.PreferredValue(vcard.FieldGender)
	var org = card.PreferredValue(vcard.FieldOrganization)

	contact := people.Person{
		Birthdays: []*people.Birthday{{Date: &birthDayDate}},
		Genders:   []*people.Gender{{Value: gender}},
		Names: []*people.Name{{
			FamilyName:  card.Name().FamilyName,
			GivenName:   card.Name().GivenName,
			DisplayName: formattedName,
		}},
		Organizations: []*people.Organization{{Name: org}},
	}
	return contact
}

func parseVcardDate(bday string) people.Date {
	var t, err = time.Parse("2006-01-02", bday)
	for err != nil {
		t, err = time.Parse("20060102", bday)
		t, err = time.Parse("01-02", bday)
		log.Printf("Unable to parse birthday date. %v", err)
		return people.Date{}
	}
	return people.Date{Year: int64(t.Year()), Month: int64(t.Month()), Day: int64(t.Day())}
}

func readVcardFromFile(filePath string) (vcard.Card, error) {
	var card vcard.Card
	file, err := os.Open(filePath)
	if err == nil {
		card, err = vcard.NewDecoder(file).Decode()
	}
	defer file.Close()
	return card, err
}
