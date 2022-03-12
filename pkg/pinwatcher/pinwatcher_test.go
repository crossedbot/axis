package pinwatcher

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	ipfscid "github.com/ipfs/go-cid"
	clusterapi "github.com/ipfs/ipfs-cluster/api"
	"github.com/stretchr/testify/require"

	"github.com/crossedbot/axis/pkg/mocks"
	"github.com/crossedbot/axis/pkg/pins/models"
)

func TestPinWatcherRegister(t *testing.T) {
	ctx := context.Background()
	pinStatus := models.PinStatus{
		Id:      "thispinsid",
		Status:  "pinning",
		Created: "1621000000",
		Pin: models.Pin{
			Cid:     "QmPAwR5un1YPJEF6iB7KvErDmAhiXxwL5J5qjA3Z9ceKqv",
			Name:    "helloworld",
			Origins: []string{"somewherefaraway"},
			Meta:    models.Info{"uid": "myuserid"},
		},
	}
	cid, err := ipfscid.Decode(pinStatus.Pin.Cid)
	require.Nil(t, err)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDb := mocks.NewMockPins(mockCtrl)
	mockIpfsClient := mocks.NewMockClient(mockCtrl)
	mockDb.EXPECT().
		Get(pinStatus.Id).
		Return(pinStatus, nil)
	mockIpfsClient.EXPECT().
		Status(ctx, cid, true).
		Return(&clusterapi.GlobalPinInfo{
			PeerMap: map[string]*clusterapi.PinInfoShort{
				"peer": &clusterapi.PinInfoShort{
					Status: clusterapi.TrackerStatusPinned,
				},
			},
		}, nil)
	mockDb.EXPECT().
		Patch(pinStatus.Id, map[string]interface{}{"status": "pinned"}).
		Return(nil)
	pw := &pinWatcher{
		Mutex:  new(sync.Mutex),
		ctx:    ctx,
		client: mockIpfsClient,
		db:     mockDb,
	}
	pw.Register(
		pinStatus,
		clusterapi.TrackerStatusPinned,
		10*time.Millisecond,
	)
	// Sleep since the Register call will create a go routine to poll the
	// pin's status
	time.Sleep(100 * time.Millisecond)
}

func TestPinWatcherDeregister(t *testing.T) {
	pid := "thispinsid"
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStatusTracker := mocks.NewMockStatusTracker(mockCtrl)
	mockStatusTracker.EXPECT().
		GetId().
		Return(pid)
	mockStatusTracker.EXPECT().
		Stop()
	pw := &pinWatcher{
		Mutex:          new(sync.Mutex),
		statusTrackers: []StatusTracker{mockStatusTracker},
	}
	pw.Deregister(pid)
}
