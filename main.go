package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"time"

	"github.com/emersion/go-vcard"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"

	"google-vcard-import/client"
)

func main() {
	var examplePath string = "example.vcf"
	var pathString = flag.String("p", examplePath, "The path to the vCard file or folder with vCard files to import")
	flag.Parse()
	
	var files []string
	var contacts []*people.ContactToCreate

	path, err := os.Stat(*pathString)
	if err == nil && path.IsDir() {
		files, err = listFilePathsInDir(pathString)
	} else if err == nil && !path.IsDir() {
		files = append(files, *pathString)
	} 
	if err != nil {
		log.Fatalf("Unable to open provided path. %v", err)
	}

	contacts = make([]*people.ContactToCreate, len(files))
	for i, file := range files {
		card, err := readVcardFromFile(file)
		if err != nil {
			log.Fatal(err)
		}
		person := parseCardToPerson(card)
		contacts[i] = &people.ContactToCreate{ContactPerson: &person}
	}

	myClient := client.GetClient("credentials.json", people.ContactsScope)

	peopleService, err := people.NewService(context.Background(), option.WithHTTPClient(myClient))
	if err != nil {
		log.Fatalf("Unable to create people Client %v", err)
	}

	createdContacts := make([]*people.PersonResponse, 0)
	const batchSize int = 200
	for low := 0; low < len(contacts); low += batchSize {
		high := int(math.Min(float64(low + batchSize), float64(len(contacts))))
		
		request := people.BatchCreateContactsRequest{
			Contacts:        contacts[low:high],
			ReadMask:        "names",
		}
		result, err := peopleService.People.BatchCreateContacts(&request).Fields().Do()
		if err != nil {
			log.Printf("result.HTTPStatusCode: %v\n", result.HTTPStatusCode)
			log.Fatalf("Unable to create contacts. %v", err)
		}
		createdContacts = append(createdContacts, result.CreatedPeople...)
	}
	
	resourceNames := make([]string, len(createdContacts))
	for i, person := range createdContacts {
		resourceNames[i] = person.RequestedResourceName
	}

	labelRequest := people.CreateContactGroupRequest{ContactGroup: &people.ContactGroup{Name: "Imported vCard: " + time.Now().Format(time.Kitchen)}}
	labelRequestResponse, err := peopleService.ContactGroups.Create(&labelRequest).Do()
	if err != nil {
		log.Fatalf("Unable to create contacts label. %v", err)
	}
	fmt.Printf("Created label: %v\n", labelRequestResponse.FormattedName)

	tagLabelRequest := people.ModifyContactGroupMembersRequest{ResourceNamesToAdd: resourceNames}
	taggingResponse, err := peopleService.ContactGroups.Members.Modify(labelRequestResponse.ResourceName, &tagLabelRequest).Do()
	if err != nil {
		log.Fatalf("Unable to apply label to contacts. %v", err)
	}
	fmt.Printf("Applied label to created contacts. Statuscode: %v", taggingResponse.HTTPStatusCode)
}

func listFilePathsInDir(path *string) ([]string, error) {
	var entries []fs.DirEntry
	var filePaths []string
	dir, err := os.Open(*path)
	if err != nil {
		log.Fatalf("Unable to open path. %v", err)
	}
	defer dir.Close()
	entries, err = dir.ReadDir(0)
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			filePath := dir.Name() + entry.Name()
			filePaths = append(filePaths, filePath)
		}
	}

	return filePaths, err
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
