package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"

	"google-vcard-import/client"
	"google-vcard-import/util"
)

func main() {
	var examplePath string = "example.vcf"
	var pathString = flag.String("p", examplePath, "The path to the vCard file or folder with vCard files to import")
	flag.Parse()

	var files []string
	var contacts []*people.ContactToCreate

	path, err := os.Stat(*pathString)
	if err == nil && path.IsDir() {
		files, err = util.ListFilePathsInDir(pathString)
	} else if err == nil && !path.IsDir() {
		files = append(files, *pathString)
	}
	if err != nil {
		log.Fatalf("Unable to open provided path. %v", err)
	}

	contacts = make([]*people.ContactToCreate, len(files))
	for i, file := range files {
		card, err := util.ReadVcardFromFile(file)
		if err != nil {
			log.Fatal(err)
		}
		person := util.ParseCardToPerson(card)
		contacts[i] = &people.ContactToCreate{ContactPerson: &person}
	}

	myClient := client.New("credentials.json", people.ContactsScope)

	peopleService, err := people.NewService(context.Background(), option.WithHTTPClient(myClient))
	if err != nil {
		log.Fatalf("Unable to create people Client %v", err)
	}

	createdContacts := createMultipleContacts(contacts, peopleService)

	label := createContactLabel(peopleService)

	resourceNames := make([]string, len(createdContacts))
	for i, person := range createdContacts {
		resourceNames[i] = person.RequestedResourceName
	}

	response := applyLabelToContacts(resourceNames, peopleService, label)
	fmt.Printf("Applied label to created contacts. Statuscode: %v", response.HTTPStatusCode)
}

func applyLabelToContacts(resourceNames []string, peopleService *people.Service, label *people.ContactGroup) *people.ModifyContactGroupMembersResponse {
	contactsToLabel := people.ModifyContactGroupMembersRequest{ResourceNamesToAdd: resourceNames}
	labellingResponse, err := peopleService.ContactGroups.Members.Modify(label.ResourceName, &contactsToLabel).Do()
	if err != nil {
		log.Fatalf("Unable to apply label to contacts. %v", err)
	}
	return labellingResponse
}

func createContactLabel(peopleService *people.Service) *people.ContactGroup {
	labelRequest := people.CreateContactGroupRequest{ContactGroup: &people.ContactGroup{Name: "Imported vCard: " + time.Now().Format(time.Kitchen)}}
	labelRequestResponse, err := peopleService.ContactGroups.Create(&labelRequest).Do()
	if err != nil {
		log.Fatalf("Unable to create contacts label. %v", err)
	}
	fmt.Printf("Created label: %v\n", labelRequestResponse.FormattedName)
	return labelRequestResponse
}

func createMultipleContacts(contacts []*people.ContactToCreate, peopleService *people.Service) []*people.PersonResponse {
	createdContacts := make([]*people.PersonResponse, 0)
	const batchSize int = 200
	for low := 0; low < len(contacts); low += batchSize {
		high := int(math.Min(float64(low+batchSize), float64(len(contacts))))

		request := people.BatchCreateContactsRequest{
			Contacts: contacts[low:high],
			ReadMask: "names",
		}
		result, err := peopleService.People.BatchCreateContacts(&request).Fields().Do()
		if err != nil {
			log.Fatalf("Unable to create contacts. %v", err)
		}
		createdContacts = append(createdContacts, result.CreatedPeople...)
	}
	return createdContacts
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
