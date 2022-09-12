package pinwatcher

import (
	"context"
	"sync"
	"time"

	clusterapi "github.com/ipfs-cluster/ipfs-cluster/api"
	cluster "github.com/ipfs-cluster/ipfs-cluster/api/rest/client"

	pinsdb "github.com/crossedbot/axis/pkg/pins/database"
	"github.com/crossedbot/axis/pkg/pins/models"
)

type PinWatcher interface {
	Register(
		p models.PinStatus,
		targetStatus clusterapi.TrackerStatus,
		checkFreq time.Duration,
	)
	Deregister(pid string)
}

type pinWatcher struct {
	*sync.Mutex
	ctx            context.Context
	client         cluster.Client
	db             pinsdb.Pins
	statusTrackers []StatusTracker
}

func New(
	ctx context.Context,
	client cluster.Client,
	db pinsdb.Pins,
) PinWatcher {
	return &pinWatcher{
		Mutex:  new(sync.Mutex),
		ctx:    ctx,
		client: client,
		db:     db,
	}
}

func (w *pinWatcher) Register(
	p models.PinStatus,
	target clusterapi.TrackerStatus,
	checkFreq time.Duration,
) {
	w.Lock()
	defer w.Unlock()
	statusTracker := NewStatusTracker(w.ctx, w.client, w.db, p)
	go statusTracker.Poll(target, checkFreq)
	w.statusTrackers = append(w.statusTrackers, statusTracker)
}

func (w *pinWatcher) Deregister(pid string) {
	w.Lock()
	defer w.Unlock()
	for i, st := range w.statusTrackers {
		if pid == st.GetId() {
			st.Stop()
			w.statusTrackers[i] =
				w.statusTrackers[len(w.statusTrackers)-1]
			w.statusTrackers =
				w.statusTrackers[:len(w.statusTrackers)-1]
		}
	}
}
