package utils

// Code from https://github.com/ssgelm/cookiejarparser under MIT License

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/publicsuffix"
)

func TestParseCookieLine(t *testing.T) {
	// normal
	cookie, err := parseCookieLine("example.com	FALSE	/	FALSE	0	test_cookie	1", 1)
	sampleCookie := &http.Cookie{
		Domain:   "example.com",
		Path:     "/",
		Name:     "test_cookie",
		Value:    "1",
		HttpOnly: false,
		Secure:   false,
	}

	if !reflect.DeepEqual(cookie, sampleCookie) || err != nil {
		c1, _ := json.Marshal(cookie)
		c2, _ := json.Marshal(sampleCookie)

		t.Errorf("Parsing normal cookie failed.  Expected:\n  cookie: %s err: nil,\ngot:\n  cookie: %s err: %s", c2, c1, err)
	}
	// httponly
	cookieHttp, err := parseCookieLine("#HttpOnly_example.com	FALSE	/	FALSE	0	test_cookie_httponly	1", 1)
	sampleCookieHttp := &http.Cookie{
		Domain:   "example.com",
		Path:     "/",
		Name:     "test_cookie_httponly",
		Value:    "1",
		HttpOnly: true,
		Secure:   false,
	}

	if !reflect.DeepEqual(cookieHttp, sampleCookieHttp) || err != nil {
		c1, _ := json.Marshal(cookieHttp)
		c2, _ := json.Marshal(sampleCookieHttp)

		t.Errorf("Parsing httpOnly cookie failed.  Expected:\n  cookie: %s err: nil,\ngot:\n  cookie: %s err: %s", c2, c1, err)
	}

	// comment
	cookieComment, err := parseCookieLine("# This is a comment", 1)
	if cookieComment != nil || err != nil {
		t.Errorf("Parsing comment failed.  Expected cookie: nil err: nil, got cookie: %s err: %s", cookie, err)
	}

	cookieBlank, err := parseCookieLine("", 1)
	if cookieBlank != nil || err != nil {
		t.Errorf("Parsing blank line failed.  Expected cookie: nil err: nil, got cookie: %s err: %s", cookie, err)
	}
}

func TestLoadCookieJarFile(t *testing.T) {
	exampleURL := &url.URL{
		Scheme: "http",
		Host:   "example.com",
	}
	sampleCookies := []*http.Cookie{
		{
			Domain:   "example.com",
			Path:     "/",
			Name:     "test_cookie",
			Value:    "1",
			HttpOnly: false,
			Secure:   false,
		},
		{
			Domain:   "example.com",
			Path:     "/",
			Name:     "test_cookie_httponly",
			Value:    "1",
			HttpOnly: true,
			Secure:   false,
		},
	}
	sampleCookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	sampleCookieJar.SetCookies(exampleURL, sampleCookies)

	cookieJar, err := LoadCookieJarFile("testdata/cookies.txt")
	require.NoError(t, err)

	c1, _ := json.Marshal(cookieJar.Cookies(exampleURL))
	c2, _ := json.Marshal(sampleCookieJar.Cookies(exampleURL))

	if !reflect.DeepEqual(c1, c2) || err != nil {
		t.Errorf("Cookie jar creation failed.  Expected:\n  cookieJar: %s err: nil,\ngot:\n  cookieJar: %s err: %s", c2, c1, err)
	}
}
