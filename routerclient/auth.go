package routerclient

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
)

func generateNonce() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	uuidText, err := uuid.MarshalBinary()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(uuidText), nil
}

func generateClientNonce() (string, error) {
	uuid1, err := generateNonce()
	if err != nil {
		return "", err
	}

	uuid2, err := generateNonce()
	if err != nil {
		return "", err
	}

	return uuid1 + uuid2, nil
}

func (c *RouterClient) updateVerificationTokenFromHeaders(httpResponse *http.Response) error {
	token := httpResponse.Header[requestVerificationToken]
	if token == nil {
		return fmt.Errorf("missing %s token in the response", requestVerificationToken)
	}

	c.requestVerificationToken = token[0][0:32]

	return nil
}

func (c *RouterClient) getServerToken() error {
	resp, err := c.client.Get(c.routerURL + tokenURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type Response struct {
		Token string `xml:"token"`
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	v := Response{}
	err = xml.Unmarshal(responseData, &v)
	if err != nil {
		return err
	}

	c.requestVerificationToken = v.Token[32:]
	return nil
}

func (c *RouterClient) challengeLogin(clientNonce string) (int, string, string, error) {
	type ChallengeLoginRequest struct {
		XMLName    xml.Name `xml:"request"`
		Username   string   `xml:"username"`
		Firstnonce string   `xml:"firstnonce"`
		Mode       int      `xml:"mode"`
	}

	challengeLoginRequest := &ChallengeLoginRequest{
		Username:   c.username,
		Firstnonce: clientNonce,
		Mode:       1,
	}

	challengeLogin, err := xml.Marshal(challengeLoginRequest)
	if err != nil {
		return 0, "", "", err
	}

	challengeLoginWithHeader := append([]byte(xml.Header), challengeLogin...)

	req, err := http.NewRequest("POST", c.routerURL+challengeLoginURL, bytes.NewReader(challengeLoginWithHeader))
	req.Header.Add("Content-Type", "text/html")
	req.Header.Add(requestVerificationToken, c.requestVerificationToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, "", "", err
	}

	type ChallengeLoginResponse struct {
		Iterations  int    `xml:"iterations"`
		Servernonce string `xml:"servernonce"`
		Salt        string `xml:"salt"`
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", "", err
	}

	c.updateVerificationTokenFromHeaders(resp)

	v := ChallengeLoginResponse{}
	err = xml.Unmarshal(responseData, &v)
	if err != nil {
		return 0, "", "", err
	}

	return v.Iterations, v.Servernonce, v.Salt, nil
}

func calculateClientProof(password string, clientNonce string, iterations int, serverNonce string, salt string) (string, error) {
	msg := fmt.Sprintf("%s,%s,%s", clientNonce, serverNonce, serverNonce)
	saltArray, err := hex.DecodeString(salt)
	if err != nil {
		return "", err
	}

	saltedPass := pbkdf2.Key([]byte(password), saltArray, iterations, 32, sha256.New)
	clientKeyHash := hmac.New(sha256.New, []byte("Client Key"))
	_, err = clientKeyHash.Write(saltedPass)
	if err != nil {
		return "", err
	}

	clientKeyDigest := clientKeyHash.Sum(nil)
	storedKey := sha256.Sum256(clientKeyDigest)
	signature := hmac.New(sha256.New, []byte(msg))
	_, err = signature.Write(storedKey[:])
	if err != nil {
		return "", err
	}

	signatureDigest := signature.Sum(nil)
	clientProof := make([]byte, len(clientKeyDigest))

	for i := 0; i < len(clientKeyDigest); i++ {
		clientProof[i] = clientKeyDigest[i] ^ signatureDigest[i]
	}

	return hex.EncodeToString(clientProof), nil
}

func (c *RouterClient) authLogin(clientNonce string, iterations int, serverNonce string, salt string) error {
	type AuthLoginRequest struct {
		XMLName     xml.Name `xml:"request"`
		ClientProof string   `xml:"clientproof"`
		FinalNonce  string   `xml:"finalnonce"`
	}

	clientProof, err := calculateClientProof(c.password, clientNonce, iterations, serverNonce, salt)

	authLoginRequest := &AuthLoginRequest{
		ClientProof: clientProof,
		FinalNonce:  serverNonce,
	}

	authLogin, err := xml.Marshal(authLoginRequest)
	if err != nil {
		return err
	}

	authLoginWithHeader := append([]byte(xml.Header), authLogin...)

	req, err := http.NewRequest("POST", c.routerURL+authLoginURL, bytes.NewReader(authLoginWithHeader))
	req.Header.Add("Content-Type", "text/html")
	req.Header.Add(requestVerificationToken, c.requestVerificationToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	c.updateVerificationTokenFromHeaders(resp)

	return nil

}

//Login initialize the sessio and logs in to router
func (c *RouterClient) Login() error {
	err := c.initSession()
	if err != nil {
		return err
	}

	err = c.getServerToken()
	if err != nil {
		return err
	}

	clientNonce, err := generateClientNonce()
	if err != nil {
		return err
	}

	iterations, serverNonce, salt, err := c.challengeLogin(clientNonce)
	if err != nil {
		return err
	}

	err = c.authLogin(clientNonce, iterations, serverNonce, salt)
	if err != nil {
		return err
	}

	return nil
}
