package controller

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/crossedbot/common/golang/config"
	middleware "github.com/crossedbot/simplemiddleware"
	"github.com/google/uuid"
	clusterapi "github.com/ipfs-cluster/ipfs-cluster/api"
	cluster "github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/crossedbot/axis/pkg/auth"
	"github.com/crossedbot/axis/pkg/pinner"
	pinsdb "github.com/crossedbot/axis/pkg/pins/database"
	"github.com/crossedbot/axis/pkg/pins/models"
	"github.com/crossedbot/axis/pkg/pinwatcher"
)

var (
	// Pins controller errors
	ErrorPinNotFound = errors.New("pin not found")
)

// Controller represents an interface to the Pins manager
type Controller interface {
	// FindPins returns all pins that match the given parameters
	FindPins(
		uid string,
		cids []string,
		name, before, after string,
		match models.TextMatch,
		statuses []models.Status,
		limit int,
	) (models.Pins, error)
	// GetPin returns the Pin for the given ID
	GetPin(uid, id string) (models.PinStatus, error)
	// CreatePin creates a new Pin based on the given Pin object
	CreatePin(uid string, pin models.Pin) (models.PinStatus, error)
	// UpdatePin updates the Pin at the given ID via the given data
	UpdatePin(uid, id string, data models.Pin) (models.PinStatus, error)
	// PatchPin modifies the Pin at the given ID via the set fields of data
	PatchPin(uid, id string, data models.Pin) (models.PinStatus, error)
	// RemovePin deletes the Pin at the given ID
	RemovePin(uid, id string) error
}

// controller implements the Controller interface
type controller struct {
	ctx     context.Context
	db      pinsdb.Pins
	pinner  pinner.Pinner
	watcher pinwatcher.PinWatcher
}

type Config struct {
	DatabaseAddr        string `toml:"database_addr"`
	AuthenticatorAddr   string `toml:"authenticator_addr"`
	DropDatabaseOnStart bool   `toml:"drop_database_on_start"`

	// IPFS configuration
	IpfsClusterSsl          bool   `toml:"ipfs_cluster_ssl"`
	IpfsClusterNoVerifyCert bool   `toml:"ipfs_cluster_no_verify_cert"`
	IpfsClusterApiAddr      string `toml:"ipfs_cluster_api_addr"`
	IpfsClusterProtectorKey []byte `toml:"ipfs_cluster_protector_key"`
	IpfsClusterTimeout      int    `toml:"ipfs_cluster_timeout"` // in seconds
}

// control is a singleton of a Controller and can be accessed via the V1
// function
var control Controller
var controllerOnce sync.Once
var V1 = func() Controller {
	// initialize the controller only once
	controllerOnce.Do(func() {
		var cfg Config
		if err := config.Load(&cfg); err != nil {
			panic(err)
		}
		ctx := context.Background()
		db, err := pinsdb.New(ctx, cfg.DatabaseAddr, cfg.DropDatabaseOnStart)
		if err != nil {
			panic(fmt.Errorf(
				"Controller: failed to connect to database at "+
					"address ('%s') with error: %s",
				cfg.DatabaseAddr, err,
			))
		}
		apiAddr, err := ma.NewMultiaddr(cfg.IpfsClusterApiAddr)
		if err != nil {
			panic(fmt.Errorf(
				"Controller: failed to parse IPFS cluster API address ('%s')",
				cfg.IpfsClusterApiAddr,
			))
		}
		ipfsClient, err := cluster.NewDefaultClient(&cluster.Config{
			SSL:          cfg.IpfsClusterSsl,
			NoVerifyCert: cfg.IpfsClusterNoVerifyCert,
			APIAddr:      apiAddr,
			ProtectorKey: cfg.IpfsClusterProtectorKey,
			Timeout:      time.Duration(cfg.IpfsClusterTimeout) * time.Second,
		})
		if err != nil {
			panic(fmt.Errorf(
				"Controller: failed to create client for IPFS "+
					"cluster ('%s') with error %s",
				cfg.IpfsClusterApiAddr, err,
			))
		}
		middleware.SetKeyFunc(auth.KeyFunc(cfg.AuthenticatorAddr))
		middleware.SetErrFunc(auth.ErrFunc())
		control = New(ctx, ipfsClient, db)
	})
	return control
}

// New returns a new Controller for the given context and Pins database
// interface
func New(ctx context.Context, client cluster.Client, db pinsdb.Pins) Controller {
	return &controller{
		ctx:     ctx,
		db:      db,
		pinner:  pinner.New(ctx, client),
		watcher: pinwatcher.New(ctx, client, db),
	}
}

func (c *controller) FindPins(
	uid string,
	cids []string,
	name, before, after string,
	match models.TextMatch,
	statuses []models.Status,
	limit int,
) (models.Pins, error) {
	var err error
	ss := []string{}
	for _, s := range statuses {
		ss = append(ss, s.String())
	}
	b := int64(0)
	if before != "" {
		b, err = strconv.ParseInt(before, 10, 64)
		if err != nil {
			return models.Pins{}, fmt.Errorf(
				"'before' (%s) is not an integer", before)
		}
	}
	a := int64(0)
	if after != "" {
		a, err = strconv.ParseInt(after, 10, 64)
		if err != nil {
			return models.Pins{}, fmt.Errorf(
				"'after' (%s) is not an integer", after)
		}
	}
	// retrieve pins from datastore
	return c.db.Find(cids, ss, name, b, a, match.String(), limit)
}

func (c *controller) GetPin(uid, id string) (models.PinStatus, error) {
	// retrieve pin from datastore
	ps, err := c.db.Get(id)
	if err == pinsdb.ErrNotFound {
		return models.PinStatus{}, ErrorPinNotFound
	}
	return ps, err
}

func (c *controller) CreatePin(uid string, pin models.Pin) (models.PinStatus, error) {
	// Add pin status to datastore
	delegates, _ := c.pinner.Delegates()
	ps := models.PinStatus{
		Id:        uuid.New().String(),
		Status:    models.StatusInit.String(),
		Created:   strconv.FormatInt(time.Now().Unix(), 10),
		Pin:       pin,
		Delegates: delegates,
	}
	err := c.db.Set(ps)
	if err != nil {
		return models.PinStatus{}, err
	}
	ps, err = c.db.Get(ps.Id)
	if err != nil {
		return models.PinStatus{}, err
	}
	go func() {
		c.watcher.Register(
			ps, clusterapi.TrackerStatusPinned, 1*time.Minute)
		c.pinner.Add(pin)
		c.watcher.Deregister(ps.Id)
	}()
	return ps, nil
}

func (c *controller) UpdatePin(uid, id string, pin models.Pin) (models.PinStatus, error) {
	fields := map[string]interface{}{
		"name":    pin.Name,
		"origins": pin.Origins,
		"meta":    pin.Meta,
	}
	if err := c.db.Patch(id, fields); err != nil {
		return models.PinStatus{}, err
	}
	return c.db.Get(id)
}

func (c *controller) PatchPin(uid, id string, pin models.Pin) (models.PinStatus, error) {
	fields := make(map[string]interface{})
	if pin.Name != "" {
		fields["name"] = pin.Name
	}
	if len(pin.Origins) > 0 {
		fields["origins"] = pin.Origins
	}
	if len(pin.Meta) > 0 {
		fields["meta"] = pin.Meta
	}
	if err := c.db.Patch(id, fields); err != nil {
		return models.PinStatus{}, err
	}
	return c.db.Get(id)
}

func (c *controller) RemovePin(uid, id string) error {
	ps, err := c.GetPin(uid, id)
	if err != nil {
		return err
	}
	pins, err := c.FindPins(
		uid,
		[]string{ps.Pin.Cid},
		"", "", "",
		models.TextMatchExact,
		[]models.Status{}, 10,
	)
	if err != nil {
		return err
	}
	if pins.Count == 1 {
		// Remove pin from IPFS if no one else is tracking it
		if err := c.pinner.Remove(ps.Pin.Cid); err != nil {
			return err
		}
	}
	if err := c.db.Delete(id); err != nil {
		return err
	}
	return nil
}
