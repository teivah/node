package openvpn

import (
	"errors"
	"github.com/mysterium/node/openvpn/management"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestClientReceivesEventsFromOpenvpnManagement(t *testing.T) {
	config := NewClientConfig("1.2.3.4", "", "", "")

	middleware := newCollectingMiddleware()
	var startCalled = false
	middleware.afterStart = func() {
		startCalled = true
	}
	var stopCalled = false
	middleware.afterStop = func() {
		stopCalled = true
	}

	client := NewClient("testdata/openvpn-client-process.sh", config, "testdataoutput", middleware)

	assert.NoError(t, client.Start())
	time.Sleep(time.Second)
	assert.NoError(t, client.Stop())
	assert.NoError(t, client.Wait())

	assert.True(t, startCalled)
	assert.Equal(
		t,
		[]string{
			">INFO:OpenVPN Management Interface Version 1 -- type 'help' for more info",
			">PASSWORD:Need 'Auth' username/password",
			">STATE:1522855903,CONNECTING,,,,,,",
			">STATE:1522855903,WAIT,,,,,,",
			">STATE:1522855903,AUTH,,,,,,",
			">STATE:1522855904,GET_CONFIG,,,,,,",
			">STATE:1522855904,ASSIGN_IP,,10.8.0.133,,,,",
			">STATE:1522855905,CONNECTED,SUCCESS,10.8.0.133,1.2.3.4,1194,,",
			">BYTECOUNT:36987,32252",
			">STATE:1522855911,EXITING,SIGTERM,,,,,",
		},
		middleware.lines,
	)
	assert.True(t, stopCalled)
}

func TestClientSendsCommandsToOpenvpnProcessAndReceivesResponses(t *testing.T) {

	config := NewClientConfig("1.2.3.4", "", "", "")

	middleware := newCollectingMiddleware()
	startCompleted := sync.WaitGroup{}
	startCompleted.Add(1)
	middleware.afterStart = func() {
		startCompleted.Done()
	}

	client := NewClient("testdata/openvpn-client-process.sh", config, "testdataoutput", middleware)

	assert.NoError(t, client.Start())
	startCompleted.Wait()

	res, err := middleware.conn.SingleLineCommand("SINGLELINE_CMD")
	assert.NoError(t, err)
	assert.Equal(t, "SINGLELINE_CMD_OK", res)

	res, lines, err := middleware.conn.MultiLineCommand("MULTILINE_CMD")
	assert.NoError(t, err)
	assert.Equal(t, "MULTILINE_CMD_OK", res)
	assert.Equal(
		t,
		[]string{
			"LINE1",
			"LINE2",
		},
		lines,
	)

	res, err = middleware.conn.SingleLineCommand("BAD_COMMAND")
	assert.Equal(t, errors.New("command error: Unknown command BAD_COMMAND"), err)

	assert.NoError(t, client.Stop())
}

func newCollectingMiddleware() *collectingMiddleware {
	return &collectingMiddleware{
		conn:       nil, //will be set by Start callback
		lines:      nil,
		afterStart: func() {},
		afterStop:  func() {},
	}
}

type collectingMiddleware struct {
	conn       management.Connection
	lines      []string
	afterStart func()
	afterStop  func()
}

func (md *collectingMiddleware) Start(connection management.Connection) error {
	md.conn = connection
	md.afterStart()
	return nil
}

func (md *collectingMiddleware) Stop(connection management.Connection) error {
	md.afterStop()
	return nil
}

func (md *collectingMiddleware) ConsumeLine(line string) (bool, error) {
	md.lines = append(md.lines, line)
	return true, nil
}
