package abra

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/h2non/gock"
)

const TEST_ABR_GUID = "TEST_ABR_GUID"

func init() {
	os.Setenv("ABR_GUID", TEST_ABR_GUID)
}

func TestSimple(t *testing.T) {
	defer gock.Off()

	gock.New("http://foo.com").
		Get("/bar").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	res, err := http.Get("http://foo.com/bar")
	if err != nil {
		t.Errorf("Expected %v, got %v", nil, err)
	}
	if res.StatusCode != 200 {
		t.Errorf("Expected %v, got %v", 200, res.StatusCode)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if string(body)[:13] != `{"foo":"bar"}` {
		t.Errorf("Expected %v, got %v", `{"foo":"bar"}`, string(body)[:13])
	}

	// Verify that we don't have pending mocks
	if !gock.IsDone() {
		t.Errorf("Expected %v, got %v", true, gock.IsDone())
	}
}

func TestABRClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Error(err)
		return
	}

	if client.BaseURL.String() != BaseURL {
		t.Errorf("Expected endpoint to be %s, got %s", BaseURL, client.BaseURL.String())
	}

	var c Abra
	c = client
	if c == nil {
		t.Errorf("This is just to test that Client implements Abra")
	}
}

func TestABRClientNoEnvSet(t *testing.T) {
	guid := os.Getenv("ABR_GUID")
	os.Unsetenv("ABR_GUID")
	defer os.Setenv("ABR_GUID", guid)

	_, err := NewClient()
	if err == nil {
		t.Errorf("Expected an error, none was raised")
	} else if err.Error() != MissingGUIDError {
		t.Error(err)
	}

	return
}

var abnTestCases = []struct {
	abn      string
	acn      string
	name     string
	filename string
}{
	{"99124391073", "", "COzero Pty Ltd", "abn/200/99124391073.xml"},
	{"26154482283", "", "Oneflare Pty Ltd", "abn/200/26154482283.xml"},
	{"65433405893", "", "STUART J AULD", "abn/200/65433405893.xml"},
}

func TestSearchByABNv201408(t *testing.T) {
	defer gock.Off()

	client, err := NewClient()
	if err != nil {
		t.Error(err)
		return
	}

	for _, c := range abnTestCases {
		body, err := ioutil.ReadFile(filepath.Join("testdata", c.filename))
		reqBody := url.Values{}
		reqBody.Set("authenticationGuid", TEST_ABR_GUID)
		reqBody.Add("includeHistoricalDetails", "Y")
		reqBody.Add("searchString", c.abn)

		gock.New("https://www.abn.business.gov.au/").
			Post("/abrxmlsearch/ABRXMLSearch.asmx/SearchByABNv201408").
			MatchType("url").
			BodyString(reqBody.Encode()).
			Reply(200).
			BodyString(string(body))

		entity, err := client.SearchByABNv201408(c.abn, true)
		if err != nil {
			t.Error(err)
			continue
		}

		if entity.Name() != c.name {
			t.Errorf("Expected %v, got %v", c.name, entity.Name())
		}

		if entity.ABN() != c.abn {
			t.Errorf("Expected %v, got %v", c.abn, entity.ABN())
		}
	}
	return
}

var asicTestCases = []struct {
	abn      string
	acn      string
	name     string
	filename string
}{
	{"78159033075", "159033075", "ENERGYLINK GLOBAL PTY LTD", "acn/200/159033075.xml"},
	{"26154482283", "154482283", "Oneflare Pty Ltd", "acn/200/154482283.xml"},
}

func TestSearchByASICv201408(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Error(err)
		return
	}

	for _, c := range asicTestCases {
		body, err := ioutil.ReadFile(filepath.Join("testdata", c.filename))
		reqBody := url.Values{}
		reqBody.Set("authenticationGuid", TEST_ABR_GUID)
		reqBody.Add("includeHistoricalDetails", "Y")
		reqBody.Add("searchString", c.acn)

		gock.New("https://www.abn.business.gov.au/").
			Post("/abrxmlsearch/ABRXMLSearch.asmx/SearchByASICv201408").
			MatchType("url").
			BodyString(reqBody.Encode()).
			Reply(200).
			BodyString(string(body))

		entity, err := client.SearchByASICv201408(c.acn, true)
		if err != nil {
			t.Error(err)
			continue
		}

		if entity.Name() != c.name {
			t.Errorf("Expected %v, got %v", c.name, entity.Name())
		}

		if entity.ABN() != c.abn {
			t.Errorf("Expected %v, got %v", c.abn, entity.ABN())
		}

		if entity.ASICNumber != c.acn {
			t.Errorf("Expected %v, got %v", c.acn, entity.ASICNumber)
		}
	}
	return
}

var nameSearchTestCases = []struct {
	name           string
	results        string
	exceedsMaximum string
	abn            string
	mainName       string
	postcode       string
	stateCode      string
	filename       string
}{
	{"COzero", "18", "N", "99124391073", "COzero Pty Ltd", "2000", "NSW", "name/200/COzero.xml"},
}

func TestSearchByNameWithEmptyString(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Error(err)
		return
	}

	result, err := client.SearchByName("  ", nil)

	if err == nil {
		t.Errorf("Expected empty request to return error, instead got success with %v\n", result)
	}

	return
}

func TestSearchByNameWithNonEmptyString(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Error(err)
		return
	}

	for _, c := range nameSearchTestCases {
		body, err := ioutil.ReadFile(filepath.Join("testdata", c.filename))
		reqBody := url.Values{}
		reqBody.Set("authenticationGuid", TEST_ABR_GUID)
		reqBody.Add("name", c.name)
		reqBody.Add("postcode", "")
		reqBody.Add("legalName", "Y")
		reqBody.Add("tradingName", "Y")
		reqBody.Add("businessName", "Y")
		reqBody.Add("activeABNsOnly", "Y")
		reqBody.Add("NSW", "Y")
		reqBody.Add("SA", "Y")
		reqBody.Add("ACT", "Y")
		reqBody.Add("VIC", "Y")
		reqBody.Add("WA", "Y")
		reqBody.Add("NT", "Y")
		reqBody.Add("QLD", "Y")
		reqBody.Add("TAS", "Y")
		reqBody.Add("searchWidth", "typical")
		reqBody.Add("minimumScore", "0")
		reqBody.Add("maxSearchResults", "50")

		gock.New("https://www.abn.business.gov.au/").
			Post("/abrxmlsearch/ABRXMLSearch.asmx/ABRSearchByNameAdvancedSimpleProtocol2017").
			MatchType("url").
			BodyString(reqBody.Encode()).
			Reply(200).
			BodyString(string(body))

		searchResults, err := client.SearchByNameAdvancedSimpleProtocol2017(c.name, nil)
		if err != nil {
			t.Error(err)
			continue
		}

		if searchResults.NumberOfRecords != 18 {
			t.Errorf("Incorrect `NumberOfRecords` value of %d", searchResults.NumberOfRecords)
		}

		if searchResults.ExceedsMaximum != "N" {
			t.Errorf("Incorrect `ExceedsMaximum` value of %s", searchResults.ExceedsMaximum)
		}

		if len(searchResults.SearchResultsRecord) != int(searchResults.NumberOfRecords) {
			t.Errorf("Counts do not match. Expected %d received %d.\n", len(searchResults.SearchResultsRecord), searchResults.NumberOfRecords)
		}
	}
	return
}

func TestEntityNumberFromString(t *testing.T) {
	number, ty := entityNumberFromString(" 0123456789 ")
	if ty != numberTypeNone {
		t.Errorf("Expected numberTypeNone, got %d", ty)
	}

	if number != "" {
		t.Errorf("Expected blank, got %v", number)
	}

	number, ty = entityNumberFromString("12-345-678-912")
	if ty != numberTypeNone {
		t.Errorf("Expected numberTypeNone, got %d", ty)
	}

	if number != "" {
		t.Errorf("Expected blank, got %v", number)
	}

	number, ty = entityNumberFromString(" 12 34 56 789 ")
	if ty != numberTypeACN {
		t.Errorf("Expected numberTypeACN, got %d", ty)
	}

	if number != "123456789" {
		t.Errorf("Expected 123456789, got %v", number)
	}

	number, ty = entityNumberFromString(" 12 34 56 789 12")
	if ty != numberTypeABN {
		t.Errorf("Expected numberTypeABN, got %d", ty)
	}

	if number != "12345678912" {
		t.Errorf("Expected 12345678912, got %v", number)
	}

	number, ty = entityNumberFromString(" 3z4 56 789 12")
	if ty != numberTypeNone {
		t.Errorf("Expected numberTypeABN, got %d", ty)
	}

	if number != "" {
		t.Errorf("Expected blank string, got %v", number)
	}

	number, ty = entityNumberFromString(" 3z 56 789 12")
	if ty != numberTypeNone {
		t.Errorf("Expected numberTypeABN, got %d", ty)
	}

	if number != "" {
		t.Errorf("Expected blank string, got %v", number)
	}
}

func TestSearchResultsName(t *testing.T) {
	s := &SearchResultsRecord{
		MainName: &SearchResultName{
			OrganisationName: "Bob's Warehouse",
		},
	}

	if s.Name() != "Bob's Warehouse" {
		t.Errorf("Expected Bob's Warehouse, got %v", s.Name())
	}

	s = &SearchResultsRecord{
		MainTradingName: &SearchResultName{
			OrganisationName: "Bob's Warehouse",
		},
	}

	if s.Name() != "Bob's Warehouse" {
		t.Errorf("Expected Bob's Warehouse, got %v", s.Name())
	}

	s = &SearchResultsRecord{
		BusinessName: &SearchResultName{
			OrganisationName: "Bob's Warehouse",
		},
	}

	if s.Name() != "Bob's Warehouse" {
		t.Errorf("Expected Bob's Warehouse, got %v", s.Name())
	}

	s = &SearchResultsRecord{
		OtherTradingName: &SearchResultName{
			OrganisationName: "Bob's Warehouse",
		},
	}

	if s.Name() != "Bob's Warehouse" {
		t.Errorf("Expected Bob's Warehouse, got %v", s.Name())
	}

	s = &SearchResultsRecord{
		LegalName: &SearchResultName{
			FullName: "Bob's Warehouse",
		},
	}

	if s.Name() != "Bob's Warehouse" {
		t.Errorf("Expected Bob's Warehouse, got %v", s.Name())
	}

	s = &SearchResultsRecord{}

	if s.Name() != "" {
		t.Errorf("Expected an empty string, got %v", s.Name())
	}
}

func TestSearchResultsScore(t *testing.T) {
	s := &SearchResultsRecord{
		MainName: &SearchResultName{
			Score: 69,
		},
	}

	if s.Score() != 69 {
		t.Errorf("Expected 69, got %d", s.Score())
	}

	s = &SearchResultsRecord{
		MainTradingName: &SearchResultName{
			Score: 69,
		},
	}

	if s.Score() != 69 {
		t.Errorf("Expected 69, got %d", s.Score())
	}

	s = &SearchResultsRecord{
		BusinessName: &SearchResultName{
			Score: 69,
		},
	}

	if s.Score() != 69 {
		t.Errorf("Expected 69, got %d", s.Score())
	}

	s = &SearchResultsRecord{
		OtherTradingName: &SearchResultName{
			Score: 69,
		},
	}

	if s.Score() != 69 {
		t.Errorf("Expected 69, got %d", s.Score())
	}

	s = &SearchResultsRecord{
		LegalName: &SearchResultName{
			Score: 69,
		},
	}

	if s.Score() != 69 {
		t.Errorf("Expected 69, got %d", s.Score())
	}

	s = &SearchResultsRecord{}

	if s.Score() != 0 {
		t.Errorf("Expected 0, got %d", s.Score())
	}
}

func TestSearchResultsIsCurrentIndicator(t *testing.T) {
	s := &SearchResultsRecord{
		MainName: &SearchResultName{
			IsCurrentIndicator: "yo",
		},
	}

	if s.IsCurrentIndicator() != "yo" {
		t.Errorf("Expected yo, got %v", s.IsCurrentIndicator())
	}

	s = &SearchResultsRecord{
		MainTradingName: &SearchResultName{
			IsCurrentIndicator: "yo",
		},
	}

	if s.IsCurrentIndicator() != "yo" {
		t.Errorf("Expected yo, got %v", s.IsCurrentIndicator())
	}

	s = &SearchResultsRecord{
		BusinessName: &SearchResultName{
			IsCurrentIndicator: "yo",
		},
	}

	if s.IsCurrentIndicator() != "yo" {
		t.Errorf("Expected yo, got %v", s.IsCurrentIndicator())
	}

	s = &SearchResultsRecord{
		OtherTradingName: &SearchResultName{
			IsCurrentIndicator: "yo",
		},
	}

	if s.IsCurrentIndicator() != "yo" {
		t.Errorf("Expected yo, got %v", s.IsCurrentIndicator())
	}

	s = &SearchResultsRecord{
		LegalName: &SearchResultName{
			IsCurrentIndicator: "yo",
		},
	}

	if s.IsCurrentIndicator() != "yo" {
		t.Errorf("Expected yo, got %v", s.IsCurrentIndicator())
	}

	s = &SearchResultsRecord{}

	if s.IsCurrentIndicator() != "" {
		t.Errorf("Expected blank string, got %v", s.IsCurrentIndicator())
	}
}

func TestCharCheck(t *testing.T) {
	c := charCheck([]byte("a")[0])
	if charCheckNotNumber != c {
		t.Errorf("Expected %d, got %d", charCheckNotNumber, c)
	}

	c = charCheck([]byte(" ")[0])
	if charCheckWhitespace != c {
		t.Errorf("Expected %d, got %d", charCheckWhitespace, c)
	}

	c = charCheck([]byte("3")[0])
	if charCheckNumber != c {
		t.Errorf("Expected %d, got %d", charCheckNumber, c)
	}
}
