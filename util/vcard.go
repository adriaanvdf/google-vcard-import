package util

import (
	"log"
	"os"
	"time"

	"github.com/emersion/go-vcard"
	"google.golang.org/api/people/v1"
)

func ParseCardToPerson(card vcard.Card) people.Person {
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

func ReadVcardFromFile(filePath string) (vcard.Card, error) {
	var card vcard.Card
	file, err := os.Open(filePath)
	if err == nil {
		card, err = vcard.NewDecoder(file).Decode()
	}
	defer file.Close()
	return card, err
}
