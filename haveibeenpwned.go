// Package haveibeenpwned provides access to the Have I been Pwned API, returning a BreachModel or a PasteModel
// if any breach/paste is found.
package haveibeenpwned

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

//API URL of haveibeenpwned.com
const API = "https://haveibeenpwned.com/api/v3/"

//BreachModel Each breach contains a number of attributes describing the incident. In the future, these attributes may expand without the API being versioned.
type BreachModel struct {
	Name         string   `json:"Name,omitempty"`
	Title        string   `json:"Title,omitempty"`
	Domain       string   `json:"Domain,omitempty"`
	BreachDate   string   `json:"BreachDate,omitempty"`
	AddedDate    string   `json:"AddedDate,omitempty"`
	ModifiedDate string   `json:"ModifiedDate,omitempty"`
	PwnCount     int      `json:"PwnCount,omitempty"`
	Description  string   `json:"Description,omitempty"`
	DataClasses  []string `json:"DataClasses,omitempty"`
	IsVerified   bool     `json:"IsVerified,omitempty"`
	IsFabricated bool     `json:"IsFabricated,omitempty"`
	IsSensitive  bool     `json:"IsSensitive,omitempty"`
	IsRetired    bool     `json:"IsRetired,omitempty"`
	IsSpamList   bool     `json:"IsSpamList,omitempty"`
	LogoPath     string   `json:"LogoPath,omitempty"`
}

//PasteModel Each paste contains a number of attributes describing it. In the future, these attributes may expand without the API being versioned.
type PasteModel struct {
	Source     string `json:"Source,omitempty"`
	ID         string `json:"Id,omitempty"`
	Title      string `json:"Title,omitempty"`
	Date       string `json:"Date,omitempty"`
	EmailCount int    `json:"EmailCount,omitempty"`
}

//BreachedAccount The most common use of the API is to return a list of all breaches a particular account has been involved in. The API takes a single parameter which is the account to be searched for. The account is not case sensitive and will be trimmed of leading or trailing white spaces. The account should always be URL encoded.
func BreachedAccount(account, domainFilter string, truncate, unverified bool) ([]BreachModel, error) {

	res, err := callService("breachedaccount", account, domainFilter, truncate, unverified)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	breaches := make([]BreachModel, 0)
	if err := json.Unmarshal(body, &breaches); err != nil {
		return nil, err
	}

	return breaches, nil
}

//Breaches Getting all breached sites in the system. A "breach" is an instance of a system having been compromised by an attacker and the data disclosed.
func Breaches(domainFilter string) ([]BreachModel, error) {

	res, err := callService("breaches", "", "", false, false)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	breaches := make([]BreachModel, 0)
	if err := json.Unmarshal(body, &breaches); err != nil {
		return nil, err
	}

	return breaches, nil

}

//Breach Sometimes just a single breach is required and this can be retrieved by the breach "name". This is the stable value which may or may not be the same as the breach "title" (which can change).
func Breach(name string) (BreachModel, error) {

	breach := new(BreachModel)
	res, err := callService("breach", name, "", false, false)
	if err != nil {
		return *breach, err
	}
	if res.StatusCode == http.StatusNotFound {
		return *breach, nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return *breach, err
	}
	defer res.Body.Close()

	if err := json.Unmarshal(body, &breach); err != nil {
		return *breach, err
	}

	return *breach, nil
}

//PasteAccount The API takes a single parameter which is the email address to be searched for. Unlike searching for breaches, usernames that are not email addresses cannot be searched for. The email is not case sensitive and will be trimmed of leading or trailing white spaces. The email should always be URL encoded.
func PasteAccount(email string) ([]PasteModel, error) {
	res, err := callService("pasteaccount", email, "", false, false)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	pastes := make([]PasteModel, 0)
	if err := json.Unmarshal(body, &pastes); err != nil {
		return nil, err
	}

	return pastes, nil

}

func callService(service, account, domainFilter string, truncate, unverified bool) (*http.Response, error) {
	client := &http.Client{}

	u, err := url.Parse(API)
	if err != nil {
		return nil, err
	}

	u.Path += service + "/" + account
	parameters := url.Values{}
	if domainFilter != "" {
		parameters.Add("domain", domainFilter)
	}
	if truncate == false {
		parameters.Add("truncateResponse", "false")
	}
	if unverified {
		parameters.Add("includeUnverified", "true")
	}
	u.RawQuery = parameters.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Go/1.15")
	req.Header.Set("hibp-api-key", os.Getenv("HIBP_API_KEY"))
	res, err := client.Do(req)

	switch res.StatusCode {
	case http.StatusBadRequest:
		return nil, errors.New("the account does not comply with an acceptable format")
	case http.StatusTooManyRequests:
		return nil, errors.New("too many requests — the rate limit has been exceeded")
	case http.StatusUnauthorized:
		return nil, errors.New("valid header `hibp-api-key` required")
	}

	if err != nil {
		return nil, err
	}
	return res, nil
}
