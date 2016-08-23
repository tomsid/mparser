package parser

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		TestTitle      string
		Message        string
		LinksContent   map[string]string
		ExpectedResult MessageInfo
	}{
		{
			TestTitle: "Simple test",
			Message:   "@asd http://google.com/ (sss)",
			LinksContent: map[string]string{
				"http://google.com/": "response <title>Google</title>",
			},
			ExpectedResult: MessageInfo{
				Emoticons: []string{"sss"},
				Mentions:  []string{"asd"},
				Links: []Link{
					{
						URL:   "http://google.com",
						Title: "Google",
					},
				},
			},
		},
		{
			TestTitle: "Test emoticons",
			Message:   "(sda)(), asd, (sads( sas), (абвгд) (morethanfifteencharactersemoticon)",
			ExpectedResult: MessageInfo{
				Emoticons: []string{"sda"},
				Mentions:  []string{},
				Links:     []Link{},
			},
		},
		{
			TestTitle: "Test mentions",
			Message:   "@@asd, @SDlds@asd, @asdasd",
			ExpectedResult: MessageInfo{
				Emoticons: []string{},
				Mentions:  []string{"asd", "SDlds", "asdasd"},
				Links:     []Link{},
			},
		},
		{
			TestTitle: "Test links",
			Message:   "Emxample http://foo. http://google.com http://site.com/ http://example.com/asd?sd=wd%2Fsd http://торрент.рф",
			LinksContent: map[string]string{
				"http://foo/":                       "Foo site without title",
				"http://google.com/":                "vlsvddv <title>Google</title> asdsa",
				"http://site.com/":                  "<title id=123>Site</title> asdasd",
				"http://example.com/asd?sd=wd%2Fsd": "<title>いくつかのテキスト</title>",
				"http://xn--e1aqbiajf.xn--p1ai/":    "<title \n id=3 >Торрент треккер</title>",
			},
			ExpectedResult: MessageInfo{
				Emoticons: []string{},
				Mentions:  []string{},
				Links: []Link{
					{
						URL:   "http://foo",
						Title: "Website",
					},
					{
						URL:   "http://google.com",
						Title: "Google",
					},
					{
						URL:   "http://site.com",
						Title: "Site",
					},
					{
						URL:   "http://example.com/asd?sd=wd%2Fsd",
						Title: "いくつかのテキスト",
					},
					{
						URL:   "http://торрент.рф",
						Title: "Торрент треккер",
					},
				},
			},
		},
	}

	for _, c := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			fmt.Fprint(w, c.LinksContent[r.URL.String()])
		}))

		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		}

		httpClient := &http.Client{Transport: transport}

		p := NewParser(httpClient, log.New(ioutil.Discard, "messageHandlerTest: ", log.LstdFlags))
		mInfo := p.Parse(c.Message)

		if !compareSlices(c.ExpectedResult.Mentions, mInfo.Mentions) {
			t.Logf("Test '%s' failed. Mentions are not equal.\n\n Expect: %#v\n Got: %#v", c.TestTitle, c.ExpectedResult.Mentions, mInfo.Mentions)
			t.Fail()
			continue
		}

		if !compareSlices(c.ExpectedResult.Emoticons, mInfo.Emoticons) {
			t.Logf("Test '%s' failed. Emoticons are not equal.\n\n Expect: %#v\n Got: %#v", c.TestTitle, c.ExpectedResult.Emoticons, mInfo.Emoticons)
			t.Fail()
			continue
		}

		if equal, err := compareLinks(c.ExpectedResult.Links, mInfo.Links); err != nil {
			t.Logf("Test '%s' failed. An error occured while comparing links: %s", c.TestTitle, err)
			t.Fail()
		} else if !equal {
			t.Logf("Test '%s' failed. Links are not equal.\n\n Expect: %#v\n Got: %#v", c.TestTitle, c.ExpectedResult.Links, mInfo.Links)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	messages := []string{
		`Hello Parser. Here some mentions: @one@two @three and some emoticons: (adsdasd2132das).
		And here are some URLs: http://asdas.com/as and http://c3dw.com`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "Doesn't matter here")
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}

	p := NewParser(httpClient, log.New(ioutil.Discard, "", log.LstdFlags))

	for _, message := range messages {
		for i := 0; i < b.N; i++ {
			p.Parse(message)
		}
	}
}

func compareSlices(slice1 []string, slice2 []string) bool {
	map1 := make(map[string]struct{})
	map2 := make(map[string]struct{})

	for _, value := range slice1 {
		map1[value] = struct{}{}
	}
	for _, value := range slice2 {
		map2[value] = struct{}{}
	}

	return reflect.DeepEqual(map1, map2)
}

func compareLinks(links1 []Link, links2 []Link) (bool, error) {
	expectedResult := make(map[Link]struct{})
	actualResult := make(map[Link]struct{})

	for _, value := range links1 {
		var err error
		value.URL, err = sanitizeUrl(value.URL)
		if err != nil {
			return false, err
		}
		expectedResult[value] = struct{}{}
	}
	for _, value := range links2 {
		var err error
		value.URL, err = sanitizeUrl(value.URL)
		if err != nil {
			return false, err
		}

		actualResult[value] = struct{}{}
	}

	return reflect.DeepEqual(expectedResult, actualResult), nil
}
