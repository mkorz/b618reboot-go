package routerclient

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

const (
	tokenURL                 = "/api/webserver/token"
	challengeLoginURL        = "/api/user/challenge_login"
	authLoginURL             = "/api/user/authentication_login"
	signalURL                = "/api/device/signal"
	controlURL               = "/api/device/control"
	requestVerificationToken = "__requestverificationtoken"
)

// RouterClient is a client used for connecting and managing router
//
// The preferred way of constructing client is calling NewRouterClient() method
type RouterClient struct {
	client                   *http.Client
	routerURL                string
	username                 string
	password                 string
	requestVerificationToken string
}

type signalBandwidth struct {
	Upload   int
	Download int
}

type signalEARFCN struct {
	Uplink   int
	Downlink int
}

type signalPower struct {
	PUSCH int
	PUCCH int
	SRS   int
	PRACH int
}

//Signal stores signal parameters
type Signal struct {
	RSRQ      int
	RSRP      int
	RSSI      int
	SINR      int
	Bandwidth signalBandwidth
	Power     signalPower
	EARFCN    signalEARFCN
}

// NewRouterClient constructs new Routerclient object, validating provided arguments
// It does not log in to router nor it creates the session
func NewRouterClient(routerURL string, username string, password string) (*RouterClient, error) {

	if routerURL == "" {
		return nil, errors.New("routerURL cannot be empty")
	}

	url, err := url.Parse(routerURL)
	if err != nil {
		return nil, err
	}

	if username == "" && url.User.Username() != "" {
		username = url.User.Username()
	}

	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	if pass, set := url.User.Password(); password == "" && set {
		password = pass
	}

	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}

	u := url.Scheme + "://" + url.Host

	routerClient := &RouterClient{
		routerURL: u,
		username:  username,
		password:  password,
		client: &http.Client{
			Jar: jar,
		},
	}

	return routerClient, nil
}

func (c RouterClient) initSession() error {
	_, err := c.client.Get(c.routerURL)

	if err != nil {
		return err
	}

	return nil
}

func getdBValue(v string) int {
	if v == "" || v[len(v)-2:] != "dB" {
		return 0
	}
	val, _ := strconv.Atoi(v[0 : len(v)-2])
	return val
}

func getdBMValue(v string) int {
	if v == "" || v[len(v)-3:] != "dBm" {
		return 0
	}
	val, _ := strconv.Atoi(v[0 : len(v)-3])
	return val
}

func getMHzValue(v string) int {
	if v == "" || v[len(v)-3:] != "MHz" {
		return 0
	}
	val, _ := strconv.Atoi(v[0 : len(v)-3])
	return val
}

func getEARFCN(v string) signalEARFCN {
	e := signalEARFCN{}
	for _, f := range strings.Fields(v) {
		if len(f) < 4 {
			continue
		}
		val, err := strconv.Atoi(f[3:])
		if err != nil {
			continue
		}

		switch f[0:3] {
		case "UL:":
			e.Uplink = val
		case "DL:":
			e.Downlink = val
		}
	}
	return e
}

func getSignalPower(v string) signalPower {
	p := signalPower{}
	for _, f := range strings.Fields(v) {
		e := strings.FieldsFunc(f, func(c rune) bool { return c == ':' })
		if len(f) < 2 {
			continue
		}
		switch e[0] {
		case "PPusch":
			p.PUSCH = getdBMValue(e[1])
		case "PPucch":
			p.PUCCH = getdBMValue(e[1])
		case "PSrs":
			p.SRS = getdBMValue(e[1])
		case "PPrach":
			p.PRACH = getdBMValue(e[1])
		}
	}
	return p
}

// GetSignalStats connects to router and fetches current signal stats
func (c *RouterClient) GetSignalStats() (Signal, error) {
	resp, err := c.client.Get(c.routerURL + signalURL)
	if err != nil {
		return Signal{}, err
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Signal{}, err
	}

	type SignalResponse struct {
		RSRQ              string `xml:"rsrq"`
		RSRP              string `xml:"rsrp"`
		RSSI              string `xml:"rssi"`
		SINR              string `xml:"sinr"`
		UploadBandwidth   string `xml:"ulbandwidth"`
		DownloadBandwidth string `xml:"dlbandwidth"`
		TXPower           string `xml:"txpower"`
		EARFCN            string `xml:"earfcn"`
	}

	v := SignalResponse{}

	err = xml.Unmarshal(responseData, &v)
	if err != nil {
		return Signal{}, err
	}
	signal := Signal{
		RSRQ: getdBValue(v.RSRQ),
		RSRP: getdBMValue(v.RSRP),
		RSSI: getdBMValue(v.RSSI),
		SINR: getdBValue(v.SINR),
		Bandwidth: signalBandwidth{
			Upload:   getMHzValue(v.UploadBandwidth),
			Download: getMHzValue(v.DownloadBandwidth),
		},
		EARFCN: getEARFCN(v.EARFCN),
		Power:  getSignalPower(v.TXPower),
	}

	return signal, nil

}

// Reboot reboots the router ;)
func (c RouterClient) Reboot() error {
	type RebootRequest struct {
		XMLName xml.Name `xml:"request"`
		Control int      `xml:"Control"`
	}

	rebootRequest := RebootRequest{
		Control: 1,
	}

	reboot, err := xml.Marshal(rebootRequest)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.routerURL+controlURL, bytes.NewReader(reboot))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Content-Length", strconv.Itoa(len(reboot)))
	req.Header.Add(requestVerificationToken, c.requestVerificationToken)
	response, err := c.client.Do(req)

	if err != nil {
		return err
	}

	type ErrorReponse struct {
		XMLName xml.Name `xml:"error"`
		Code    int      `xml:"code"`
		Message string   `xml:"message"`
	}

	responseData, _ := ioutil.ReadAll(response.Body)
	e := ErrorReponse{}
	err = xml.Unmarshal(responseData, &e)

	if err == nil {
		return fmt.Errorf("error rebooting router, response %v", e)
	}

	return nil

}
