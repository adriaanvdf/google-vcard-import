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

	"vdfeltz.com/workspace/google-vcard-import/client"
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

	y := people.ModifyContactGroupMembersRequest{
		ResourceNamesToAdd:    []string{results3.ResourceName},
		ResourceNamesToRemove: []string{},
		ForceSendFields:       []string{},
		NullFields:            []string{},
	}
	results2, err := peopleService.ContactGroups.Members.Modify(results1.ResourceName, &y).Do()
	if err != nil {
		log.Fatalf("Unable to apply label to contacts. %v", err)
	}
	fmt.Printf("results1: %v\n", results2)

	results4, err := peopleService.People.Connections.List("people/me").PageSize(10).
		PersonFields("names,emailAddresses").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve people. %v", err)
	}
	if len(results4.Connections) > 0 {
		fmt.Print("List 10 connection names:\n")
		for _, c := range results4.Connections {
			names := c.Names
			if len(names) > 0 {
				name := names[0].DisplayName
				fmt.Printf("%s\n", name)
			}
		}
	} else {
		fmt.Print("No connections found.\n")
	}
}

func parseCardToPerson(card vcard.Card) people.Person {
	var formattedName = card.PreferredValue(vcard.FieldFormattedName)
	var bday = card.PreferredValue(vcard.FieldBirthday)
	t, err := time.Parse("20060102", bday)
	if err != nil {
		t, err = time.Parse("2006-01-02", bday)
	}
	if err != nil {
		t, err = time.Parse("01-02", bday)
	}
	if err != nil {
		log.Fatalf("Unable to parse birthday date. %v", err)
	}

	var birthDayDate = people.Date{Year: int64(t.Year()), Month: int64(t.Month()), Day: int64(t.Day())}
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

func readVcardFromFile(filePath string) (vcard.Card, error) {
	var card vcard.Card
	file, err := os.Open(filePath)
	if err == nil {
		card, err = vcard.NewDecoder(file).Decode()
	}
	defer file.Close()
	return card, err
}
