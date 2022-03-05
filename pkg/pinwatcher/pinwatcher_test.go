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
	mockDb.EXPECT().
		Get(pinStatus.Id).
		Return(pinStatus, nil)
	mockDb.EXPECT().
		Patch(pinStatus.Id, map[string]interface{}{"status": "pinned"}).
		Return(nil)
	mockIpfsClient := mocks.NewMockClient(mockCtrl)
	mockIpfsClient.EXPECT().
		Status(ctx, cid, true).
		Return(&clusterapi.GlobalPinInfo{
			PeerMap: map[string]*clusterapi.PinInfoShort{
				"peer": &clusterapi.PinInfoShort{
					Status: clusterapi.TrackerStatusPinned,
				},
			},
		}, nil)
	pw := &pinWatcher{
		Mutex:  new(sync.Mutex),
		ctx:    ctx,
		client: mockIpfsClient,
		db:     mockDb,
	}
	pw.Register(pinStatus, clusterapi.TrackerStatusPinned, 1*time.Second)
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
