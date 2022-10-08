package controller

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	clusterapi "github.com/ipfs-cluster/ipfs-cluster/api"
	ipfscid "github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/crossedbot/axis/pkg/mocks"
	"github.com/crossedbot/axis/pkg/pins/match"
	"github.com/crossedbot/axis/pkg/pins/models"
)

func TestFindPins(t *testing.T) {
	cids := []string{"abc123", "def456", "ghi789"}
	statuses := []models.Status{models.StatusPinning, models.StatusPinned}
	name := "helloworld"
	before := "1621039121"
	after := "1620967121"
	match := models.TextMatchExact
	limit := 10
	expected := models.Pins{
		Count: 1,
		Results: []models.PinStatus{{
			Id:      "thisispinsid",
			Status:  "pinned",
			Created: "1621000000",
			Pin: models.Pin{
				Cid:     "abc123",
				Name:    "helloworld",
				Origins: []string{"somewherefaraway"},
				Meta:    models.Info{"uid": "myuserid"},
			},
		}},
	}
	statusStrings := []string{}
	for _, s := range statuses {
		statusStrings = append(statusStrings, s.String())
	}
	b64, err := strconv.ParseInt(before, 10, 64)
	require.Nil(t, err)
	a64, err := strconv.ParseInt(after, 10, 64)
	require.Nil(t, err)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockdb := mocks.NewMockPins(mockCtrl)
	mockdb.EXPECT().
		Find(
			cids, statusStrings, name,
			b64, a64, match.String(), limit,
		).
		Return(expected, nil)
	ctrl := &controller{db: mockdb}
	actual, err := ctrl.FindPins(
		"", cids, name, before,
		after, match, statuses,
		limit,
	)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}

func TestGetPin(t *testing.T) {
	expected := models.PinStatus{
		Id:      "thispinsid",
		Status:  "pinned",
		Created: "1621000000",
		Pin: models.Pin{
			Cid:     "abc123",
			Name:    "helloworld",
			Origins: []string{"somewherefaraway"},
			Meta:    models.Info{"uid": "myuserid"},
		},
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockdb := mocks.NewMockPins(mockCtrl)
	mockdb.EXPECT().
		Get(expected.Id).
		Return(expected, nil)
	mockdb.EXPECT().
		Get(expected.Id).
		Return(expected, nil)
	ctrl := &controller{db: mockdb}
	actual, err := ctrl.GetPin("", expected.Id)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}

func TestCreatePin(t *testing.T) {
	pin := models.Pin{
		Cid:     "abc123",
		Name:    "helloworld",
		Origins: []string{"somewherefaraway"},
		Meta:    models.Info{"uid": "myuserid"},
	}
	expected := models.PinStatus{
		Id:      "thispinsid",
		Status:  "pinned",
		Created: "1621000000",
		Pin:     pin,
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockdb := mocks.NewMockPins(mockCtrl)
	mockdb.EXPECT().
		Set(match.PinStatus(pin)).
		Return(nil)
	mockdb.EXPECT().
		Get(gomock.Any()).
		Return(expected, nil)
	mockPinner := mocks.NewMockPinner(mockCtrl)
	mockPinner.EXPECT().
		Delegates().
		Return([]string{}, nil)
	mockPinner.EXPECT().
		Add(pin).
		Return(nil)
	mockWatcher := mocks.NewMockPinWatcher(mockCtrl)
	mockWatcher.EXPECT().
		Register(
			expected,
			clusterapi.TrackerStatusPinned,
			1*time.Minute,
		)
	mockWatcher.EXPECT().Deregister(expected.Id)
	ctrl := &controller{
		db:      mockdb,
		pinner:  mockPinner,
		watcher: mockWatcher,
	}
	actual, err := ctrl.CreatePin("", pin)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
	time.Sleep(1 * time.Millisecond)
}

func TestUpdatePin(t *testing.T) {
	pin := models.Pin{
		Cid:     "abc123",
		Name:    "helloworld",
		Origins: []string{"somewherefaraway"},
		Meta:    models.Info{"uid": "myuserid"},
	}
	expected := models.PinStatus{
		Id:      "thispinsid",
		Status:  "pinned",
		Created: "1621000000",
		Pin:     pin,
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockdb := mocks.NewMockPins(mockCtrl)
	mockdb.EXPECT().
		Patch(expected.Id, map[string]interface{}{
			"name":    pin.Name,
			"origins": pin.Origins,
			"meta":    pin.Meta,
		}).
		Return(nil)
	mockdb.EXPECT().
		Get(expected.Id).
		Return(expected, nil)
	ctrl := &controller{db: mockdb}
	actual, err := ctrl.UpdatePin("", expected.Id, pin)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}

func TestPatchPin(t *testing.T) {
	pin := models.Pin{
		Cid:     "abc123",
		Name:    "helloworld",
		Origins: []string{"somewherefaraway"},
		Meta:    models.Info{"uid": "myuserid"},
	}
	expected := models.PinStatus{
		Id:      "thispinsid",
		Status:  "pinned",
		Created: "1621000000",
		Pin:     pin,
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockdb := mocks.NewMockPins(mockCtrl)
	mockdb.EXPECT().
		Patch(expected.Id, map[string]interface{}{
			"name":    pin.Name,
			"origins": pin.Origins,
			"meta":    pin.Meta,
		}).
		Return(nil)
	mockdb.EXPECT().
		Get(expected.Id).
		Return(expected, nil)
	ctrl := &controller{db: mockdb}
	actual, err := ctrl.PatchPin("", expected.Id, pin)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}

func TestRemovePin(t *testing.T) {
	expected := models.PinStatus{
		Id:      "thispinsid",
		Status:  "pinned",
		Created: "1621000000",
		Pin: models.Pin{
			Cid:     "abc123",
			Name:    "helloworld",
			Origins: []string{"somewherefaraway"},
			Meta:    models.Info{"uid": "myuserid"},
		},
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockdb := mocks.NewMockPins(mockCtrl)
	mockdb.EXPECT().
		Get(expected.Id).
		Return(expected, nil)
	mockdb.EXPECT().
		Delete(expected.Id).
		Return(nil)
	mockPinner := mocks.NewMockPinner(mockCtrl)
	mockPinner.EXPECT().
		Remove(expected.Pin.Cid).
		Return(nil)
	ctrl := &controller{db: mockdb, pinner: mockPinner}
	err := ctrl.RemovePin("", expected.Id)
	require.Nil(t, err)
}

func TestUpdatePinStatus(t *testing.T) {
	ctx := context.Background()
	ps := models.PinStatus{
		Id:      "thispinsid",
		Status:  "pinned",
		Created: "1621000000",
		Pin: models.Pin{
			Cid:     "QmPAwR5un1YPJEF6iB7KvErDmAhiXxwL5J5qjA3Z9ceKqv",
			Name:    "helloworld",
			Origins: []string{"somewherefaraway"},
			Meta:    models.Info{"uid": "myuserid"},
		},
	}
	cid, err := ipfscid.Decode(ps.Pin.Cid)
	require.Nil(t, err)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockdb := mocks.NewMockPins(mockCtrl)
	mockdb.EXPECT().
		Get(ps.Id).
		Return(ps, nil)
	mockIpfsClient := mocks.NewMockClient(mockCtrl)
	mockIpfsClient.EXPECT().
		Status(ctx, clusterapi.NewCid(cid), true).
		Return(clusterapi.GlobalPinInfo{
			PeerMap: map[string]clusterapi.PinInfoShort{
				"peer": clusterapi.PinInfoShort{
					Status: clusterapi.TrackerStatusPinned,
				},
			},
		}, nil)
	ctrl := &controller{ctx: ctx, db: mockdb, client: mockIpfsClient}
	require.Nil(t, ctrl.UpdatePinStatus("", ps.Id))
}
