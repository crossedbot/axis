package pinwatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/crossedbot/common/golang/logger"
	clusterapi "github.com/ipfs-cluster/ipfs-cluster/api"
	cluster "github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	ipfscid "github.com/ipfs/go-cid"

	pinsdb "github.com/crossedbot/axis/pkg/pins/database"
	"github.com/crossedbot/axis/pkg/pins/models"
)

type StatusTracker interface {
	Poll(target clusterapi.TrackerStatus, checkFreq time.Duration)
	CheckStatus(target clusterapi.TrackerStatus) bool
	Stop()
	GetId() string
}

type statusTracker struct {
	ctx    context.Context
	client cluster.Client
	pin    models.PinStatus
	db     pinsdb.Pins
	// TODO we should be closing this channel when quiting
	quit chan struct{}
}

func NewStatusTracker(
	ctx context.Context,
	client cluster.Client,
	db pinsdb.Pins,
	pin models.PinStatus,
) StatusTracker {
	return &statusTracker{
		ctx:    ctx,
		client: client,
		db:     db,
		pin:    pin,
		quit:   make(chan struct{}),
	}
}

func (t *statusTracker) Poll(
	target clusterapi.TrackerStatus,
	checkFreq time.Duration,
) {
	ticker := time.NewTicker(checkFreq)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			t.CheckStatus(target)
			return
		case <-t.quit:
			t.CheckStatus(target)
			return
		case <-ticker.C:
			if t.CheckStatus(target) {
				return
			}
		}
	}
}

func (t *statusTracker) CheckStatus(target clusterapi.TrackerStatus) bool {
	// get the currently known status of the pin
	var err error
	t.pin, err = t.db.Get(t.pin.Id)
	if err != nil {
		logger.Error(fmt.Sprintf(
			"Failed to get Pin: %s",
			err,
		))
		return false
	}
	prevStatus := t.pin.Status
	currStatus := prevStatus
	// poll the cluster peer for the pin's status
	cid, err := ipfscid.Decode(t.pin.Pin.Cid)
	if err != nil {
		logger.Error(fmt.Sprintf(
			"Failed to decode pin's CID: %s",
			err,
		))
		return false
	}
	gblPinInfo, err := t.client.Status(t.ctx, clusterapi.NewCid(cid), true)
	if err != nil {
		logger.Error(fmt.Sprintf(
			"Failed to get status: %s",
			err,
		))
		return false
	}
	targetReached := false
	for _, pinInfo := range gblPinInfo.PeerMap {
		// Assume there is one in the map due to local being true
		currStatus = pinInfo.Status.String()
		if pinInfo.Status == target {
			targetReached = true
		}
		break
	}
	// Patch the pins status if status has changed
	if prevStatus != currStatus {
		t.db.Patch(t.pin.Id, map[string]interface{}{
			"status": currStatus,
		})
	}
	return targetReached
}

func (t *statusTracker) Stop() {
	t.quit <- struct{}{}
}

func (t *statusTracker) GetId() string {
	return t.pin.Id
}
