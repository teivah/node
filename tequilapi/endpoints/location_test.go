package endpoints

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mysterium/node/ip"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/mysterium/node/location"
	"github.com/mysterium/node/client/connection"
	"github.com/mysterium/node/identity"
)

type fakeManagerForLocation struct {
	onStatusReturn     connection.ConnectionStatus
}

func (fm *fakeManagerForLocation) Connect(consumerID identity.Identity, providerID identity.Identity) error {
	return nil
}

func (fm *fakeManagerForLocation) Status() connection.ConnectionStatus {
	return fm.onStatusReturn
}

func (fm *fakeManagerForLocation) Disconnect() error {
	return nil
}

func TestAddRoutesForLocationAddsRoutes(t *testing.T) {
	fakeManager := fakeManagerForLocation{}
	fakeManager.onStatusReturn = connection.ConnectionStatus{
		State: connection.Connected,
	}

	router := httprouter.New()

	currentIpResolver := ip.NewFakeResolver("123.123.123.123")
	currentLocationResolver := location.NewResolverFake("current country")
	currentLocationDetector := location.NewDetectorWithLocationResolver(currentIpResolver, currentLocationResolver)

	originalIpResolver := ip.NewFakeResolver("100.100.100.100")
	originalLocationResolver := location.NewResolverFake("original country")
	originalLocationDetector := location.NewDetectorWithLocationResolver(originalIpResolver, originalLocationResolver)
	originalLocationCache := location.NewLocationCache(originalLocationDetector)
	originalLocationCache.RefreshAndGet()

	AddRoutesForLocation(router, &fakeManager, currentLocationDetector, originalLocationCache)

	tests := []struct {
		method         string
		path           string
		body           string
		expectedStatus int
		expectedJSON   string
	}{
		{
			http.MethodGet, "/location", "",
			http.StatusOK,
			`{
				"original": {"ip": "100.100.100.100", "country": "original country"},
				"current":  {"ip": "123.123.123.123", "country": "current country"}
			}`,
		},
	}

	for _, test := range tests {
		resp := httptest.NewRecorder()
		req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		router.ServeHTTP(resp, req)
		assert.Equal(t, test.expectedStatus, resp.Code)
		if test.expectedJSON != "" {
			assert.JSONEq(t, test.expectedJSON, resp.Body.String())
		} else {
			assert.Equal(t, "", resp.Body.String())
		}
	}
}

func TestGetLocationWhenConnected(t *testing.T) {
	fakeManager := fakeManager{}
	fakeManager.onStatusReturn = connection.ConnectionStatus{
		State: connection.Connected,
	}

	currentIpResolver := ip.NewFakeResolver("123.123.123.123")
	currentLocationResolver := location.NewResolverFake("current country")
	currentLocationDetector := location.NewDetectorWithLocationResolver(currentIpResolver, currentLocationResolver)

	originalIpResolver := ip.NewFakeResolver("100.100.100.100")
	originalLocationResolver := location.NewResolverFake("original country")
	originalLocationDetector := location.NewDetectorWithLocationResolver(originalIpResolver, originalLocationResolver)
	originalLocationCache := location.NewLocationCache(originalLocationDetector)
	originalLocationCache.RefreshAndGet()

	connEndpoint := NewLocationEndpoint(&fakeManager, currentLocationDetector, originalLocationCache)
	resp := httptest.NewRecorder()

	connEndpoint.GetLocation(resp, nil, nil)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(
		t,
		`{
			"original": {"ip": "100.100.100.100", "country": "original country"},
			"current":  {"ip": "123.123.123.123", "country": "current country"}
		}`,
		resp.Body.String(),
	)
}

func TestGetLocationWhenNotConnected(t *testing.T) {
	originalIpResolver := ip.NewFakeResolver("100.100.100.100")
	originalLocationResolver := location.NewResolverFake("original country")
	originalLocationDetector := location.NewDetectorWithLocationResolver(originalIpResolver, originalLocationResolver)
	originalLocationCache := location.NewLocationCache(originalLocationDetector)
	originalLocationCache.RefreshAndGet()

	states := []connection.State{
		connection.NotConnected,
		connection.Connecting,
		connection.Disconnecting,
		connection.Reconnecting,
	}

	for _, state := range states {

		fakeManager := fakeManager{}
		fakeManager.onStatusReturn = connection.ConnectionStatus{
			State: state,
		}

		connEndpoint := NewLocationEndpoint(&fakeManager, nil, originalLocationCache)
		resp := httptest.NewRecorder()

		connEndpoint.GetLocation(resp, nil, nil)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.JSONEq(
			t,
			`{
			"original": {"ip": "100.100.100.100", "country": "original country"},
			"current":  {"ip": "100.100.100.100", "country": "original country"}
		}`,
			resp.Body.String(),
		)
	}
}
