package routerclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanRetrieveToken(t *testing.T) {
	expectedToken := "S7xTD7LGjRhXAtLCQRYWUKc5YdaRuzJJ"
	serverToken := "TJECW4tvlJ2fGiClXMwew8wiGWRKlzvz" + expectedToken

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Wrong request method, should be GET, got %s", r.Method)
		}
		correctURL := "/api/webserver/token"
		if r.URL.RequestURI() != correctURL {
			t.Errorf("Wrong URL called, should be %s got %s", correctURL, r.URL.Path)
		}

		fmt.Fprintf(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><response><token>%s</token></response>", serverToken)
	}))
	defer ts.Close()

	client, err := NewRouterClient(ts.URL, "user", "pass")

	assert.Nil(t, err, "error creating client: %q", err)

	err = client.getServerToken()

	assert.Nil(t, err, "error retrieving token %q", err)
	assert.Equal(t, expectedToken, client.requestVerificationToken, "invalid token returned, expected")
}

func TestCookieIsSendBack(t *testing.T) {
	const CookieName = "test-cookie"
	const CookieValue = "sample-cookie-value"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Wrong request method, should be GET, got %s", r.Method)
		}
		switch r.URL.RequestURI() {
		case "/":
			http.SetCookie(w, &http.Cookie{Name: CookieName, Value: CookieValue})
		case "/api/webserver/token":
			c, err := r.Cookie(CookieName)
			if err != nil {
				t.Errorf("Error retrieving value for cookie: %q", err)
			}
			if c.Value != CookieValue {
				t.Errorf("Wrong value for cookie %s received, expected %s got %s", CookieName, c.Value, CookieValue)
			}
		default:
			t.Errorf("wrong URL called %s", r.URL.RequestURI())
		}
	}))
	defer ts.Close()

	client, err := NewRouterClient(ts.URL, "user", "pass")
	assert.Nil(t, err, "error creating client: %q", err)
	client.initSession()
	client.getServerToken()
}

func TestNonceHasCorrectLength(t *testing.T) {

	nonce, err := generateClientNonce()
	if err != nil {
		t.Errorf("Error generating nonce: %q", err)
	}
	assert.NotNil(t, nonce)
	assert.Len(t, nonce, 64)
}

func TestCanDoChallengeLogin(t *testing.T) {
	token := "TJECW4tvlJ2fGiClXMwew8wiGWRKlabcS7xTD7LGjRhXAtLCQRYWUKc5YdaRuzJJ"
	clientNonce := "a8da1039e4ff4b71ba402591a7a324a7c400c068ed6c4697b670ec9002da816b"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("Wrong request method, should be GET, got %s", r.Method)
		}

		correctURL := "/api/user/challenge_login"
		if r.URL.RequestURI() != correctURL {
			t.Errorf("Wrong URL called, should be %s got %s", correctURL, r.URL.Path)
		}

		rvt := r.Header["__requestverificationtoken"]
		if rvt == nil {
			t.Errorf("Missing RequestVerificationToken in the request.")
		}

		if rvt[0] != token {
			t.Errorf("Invalid header received, expected %s got %s", token, rvt[0])
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Error reading body: %q", err)
		}

		body := string(b)

		if body != fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<request><username>admin</username><firstnonce>%s</firstnonce><mode>1</mode></request>", clientNonce) {
			t.Errorf("Invalid body received: %s", body)
		}
		fmt.Fprintf(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><response><iterations>100</iterations><servernonce>3325010dd3ff01f13ae515b4d8705c62ee34c019c64f83e059e3a6e3376f9b51Lpcw0a320YeprpYH8kURAUwfyTbYtHUA</servernonce><modeselected>1</modeselected><salt>fd4b1e6ad1b05db6ff288928fed3005ef4fdc9ade8be276220a8f41adcccda29</salt><newType>0</newType></response>")
	}))

	defer ts.Close()
	client, err := NewRouterClient(ts.URL, "admin", "pass")

	assert.Nil(t, err, "error creating client: %q", err)

	client.requestVerificationToken = token
	iterations, serverNonce, salt, err := client.challengeLogin(clientNonce)

	assert.Nil(t, err, "error retrieving token %q", err)

	assert.Equal(t, 100, iterations)
	assert.Equal(t, "3325010dd3ff01f13ae515b4d8705c62ee34c019c64f83e059e3a6e3376f9b51Lpcw0a320YeprpYH8kURAUwfyTbYtHUA", serverNonce)
	assert.Equal(t, "fd4b1e6ad1b05db6ff288928fed3005ef4fdc9ade8be276220a8f41adcccda29", salt)
}

func TestCanDoAuthLogin(t *testing.T) {
	token := "TJECW4tvlJ2fGiClXMwew8wiGWRKlabcS7xTD7LGjRhXAtLCQRYWUKc5YdaRuzJJ"
	clientNonce := "6ae8b6a8273fa166c24fb11f63d2910db3a4602411023982578d03ea6caa1c54"
	serverNonce := "6ae8b6a8273fa166c24fb11f63d2910db3a4602411023982578d03ea6caa1c54QscEJiy2Dbs0RJAsUN4rz4r8eZnfqbof"
	salt := "fd4b1e6ad1b05db6ff288928fed3005ef4fdc9ade8be276220a8f41adcccda29"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Wrong request method, should be GET, got %s", r.Method)
		}

		correctURL := "/api/user/authentication_login"
		if r.URL.RequestURI() != correctURL {
			t.Errorf("Wrong URL called, should be %s got %s", correctURL, r.URL.Path)
		}

		rvt := r.Header["__requestverificationtoken"]
		if rvt == nil {
			t.Errorf("Missing RequestVerificationToken in the request.")
		}

		if rvt[0] != token {
			t.Errorf("Invalid header received, expected %s got %s", token, rvt[0])
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Error reading body: %q", err)
		}

		body := string(b)

		if body != fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<request><clientproof>2afb0409edf53eec750bc0ec65d2802d518cc125bbe925799e5a721b8cfe96a1</clientproof><finalnonce>%s</finalnonce></request>", serverNonce) {
			t.Errorf("Invalid body received: %s", body)
		}
		w.Header().Add("__RequestVerificationToken", "25ae2067cf278b183daab21a32d133e5")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><response><rsan>94be3e6833721570f6a210e7312c8ea223f641919398b2f2683e4f2f661225775a02893e11c35c70c49ee6064df4d1bcfe9bc90f1dad3b9e2a2c506736cca048b8bd11b868805badb9b10c2cc11b677487a0ab84c1ef675a3fb6f40023971d4211f6d508bdcebaed5ef935911589c4d076ae59bcbbb094e6bde43ebd04c43025f3dfef243a84d4c267bf4c1361f3126e55a989f55a70b44dde4d84be518caed1d83f287efafde1da665a61a6f95ee3ad551e60178ae321d0772267aff83e5a76ae76146885ec09c705f6adafbe38b32015cf18385f166921f963ccc7e531496c0efd5f5a7659e5ad2e6bac2731b6e908aa4d2712b7e18d583105f1b25a68260d</rsan><rsae>010001</rsae><serversignature>0ad0937ab319b6ccdc442e3aa91c099975d4631165c0faca1e1f376e2bd39445</serversignature><rsapubkeysignature>868feb7fcd8dc67021affa011ac02a9d49b55c0ed9a8248cbb0c0aac260c0e90</rsapubkeysignature></response>")
	}))

	defer ts.Close()
	client, err := NewRouterClient(ts.URL, "admin", "MySecretPassword")
	client.requestVerificationToken = token

	assert.Nil(t, err, "error creating client: %q", err)

	err = client.authLogin(clientNonce, 100, serverNonce, salt)

	assert.Nil(t, err, "error retrieving token %q", err)

	assert.Equal(t, "25ae2067cf278b183daab21a32d133e5", client.requestVerificationToken)
}

func TestCanCalculateClientProof(t *testing.T) {
	clientNonce := "6ae8b6a8273fa166c24fb11f63d2910db3a4602411023982578d03ea6caa1c54"
	serverNonce := "6ae8b6a8273fa166c24fb11f63d2910db3a4602411023982578d03ea6caa1c54QscEJiy2Dbs0RJAsUN4rz4r8eZnfqbof"
	salt := "fd4b1e6ad1b05db6ff288928fed3005ef4fdc9ade8be276220a8f41adcccda29"
	res, err := calculateClientProof("MySecretPassword", clientNonce, 100, serverNonce, salt)

	assert.Nil(t, err, "error calculating proof %q", err)

	expectedClientProof := "2afb0409edf53eec750bc0ec65d2802d518cc125bbe925799e5a721b8cfe96a1"

	assert.Equal(t, expectedClientProof, res)
}
