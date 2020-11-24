package routerclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientRequireUsername(t *testing.T) {

	_, err := NewRouterClient("http://localhost", "", "Password")

	assert.NotNil(t, err)
}

func TestCanGetSignal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Wrong request method, should be GET, got %s", r.Method)
		}
		correctURL := "/api/device/signal"
		if r.URL.RequestURI() != correctURL {
			t.Errorf("Wrong URL called, should be %s got %s", correctURL, r.URL.Path)
		}

		fmt.Fprint(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<response>\n<pci>43</pci>\n<sc></sc>\n<cell_id>44294436</cell_id>\n<rsrq>-14dB</rsrq>\n<rsrp>-86dBm</rsrp>\n<rssi>-61dBm</rssi>\n<sinr>10dB</sinr>\n<rscp></rscp>\n<ecio></ecio>\n<mode>7</mode>\n<ulbandwidth>15MHz</ulbandwidth>\n<dlbandwidth>15MHz</dlbandwidth>\n<txpower>PPusch:8dBm PPucch:-5dBm PSrs:0dBm PPrach:-4dBm</txpower>\n<tdd></tdd>\n<ul_mcs>mcsUpCarrier1:22</ul_mcs>\n<dl_mcs>mcsDownCarrier1Code0:25 mcsDownCarrier1Code1:25 </dl_mcs>\n<earfcn>DL:3025 UL:21025</earfcn>\n<rrc_status></rrc_status>\n<rac></rac>\n<lac></lac>\n<tac>57332</tac>\n<band>7</band>\n<nei_cellid>No1:42No2:19No3:43No4:44No5:20</nei_cellid>\n<plmn>26003</plmn>\n<ims>0</ims>\n</response>\n")
	}))
	defer ts.Close()
	client, err := NewRouterClient(ts.URL, "user", "pass")
	if err != nil {
		t.Errorf("error creating RouterClient %q", err)
	}

	signal, err := client.GetSignalStats()
	if err != nil {
		t.Errorf("error getting signal stats %q", err)
	}
	assert.EqualValues(t, Signal{
		RSRQ: -14,
		RSRP: -86,
		RSSI: -61,
		SINR: 10,
		Bandwidth: struct {
			Upload   int
			Download int
		}{
			Upload:   15,
			Download: 15,
		},
		Power: struct {
			PUSCH int
			PUCCH int
			SRS   int
			PRACH int
		}{
			PUSCH: 8,
			PUCCH: -5,
			SRS:   0,
			PRACH: -4,
		},
		EARFCN: struct {
			Uplink   int
			Downlink int
		}{
			Uplink:   21025,
			Downlink: 3025,
		},
	}, signal)
}

func TestErrorDuringReboot(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("Wrong request method, should be POST, got %s", r.Method)
		}
		correctURL := "/api/device/control"
		if r.URL.RequestURI() != correctURL {
			t.Errorf("Wrong URL called, should be %s got %s", correctURL, r.URL.Path)
		}
		fmt.Fprint(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<error>\n<code>100003</code>\n<message/>\n</error>\n")
	}))
	defer ts.Close()
	client, err := NewRouterClient(ts.URL, "user", "pass")

	assert.Nil(t, err, "error creating RouterClient %q", err)

	err = client.Reboot()
	assert.NotNil(t, err, "reboot should trigger error")
}

func TestRouterClientRequiresUsername(t *testing.T) {
	_, err := NewRouterClient("http://localhost", "", "pass")
	assert.EqualError(t, err, "username cannot be empty")
}

func TestRouterClientRequiresPassword(t *testing.T) {
	_, err := NewRouterClient("http://localhost", "user", "")
	assert.EqualError(t, err, "password cannot be empty")
}

func TestRouterClientRequiresUrl(t *testing.T) {
	_, err := NewRouterClient("", "user", "pass")
	assert.EqualError(t, err, "routerURL cannot be empty")
}

func TestRouterClientAcceptsUserAndPasswordInURL(t *testing.T) {
	_, err := NewRouterClient("https://user:pass@localhost/", "", "")
	assert.Nil(t, err)
}
