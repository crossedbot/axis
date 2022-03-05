package pinner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	ipfscid "github.com/ipfs/go-cid"
	clusterapi "github.com/ipfs/ipfs-cluster/api"
	cluster "github.com/ipfs/ipfs-cluster/api/rest/client"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/crossedbot/axis/pkg/pins/models"
)

const (
	DefaultReplicationFactorMin = 1
	DefaultReplicationFactorMax = 3
	DefaultPinMode              = clusterapi.PinModeRecursive
)

var (
	ErrInvalidMultiaddrFormat = errors.New("invalid multiaddr format")
)

type Pinner interface {
	Add(pin models.Pin) error
	Remove(cid string) error
	Delegates() ([]string, error)

	// Controls
	SetReplicationFactor(min, max int) error
	SetPinMode(mode clusterapi.PinMode)
}

type pinner struct {
	ctx        context.Context
	ipfsClient cluster.Client

	// Attributes
	ReplicationFactorMin int
	ReplicationFactorMax int
	PinMode              clusterapi.PinMode
}

func New(ctx context.Context, client cluster.Client) Pinner {
	p := &pinner{
		ctx:                  ctx,
		ipfsClient:           client,
		ReplicationFactorMin: DefaultReplicationFactorMin,
		ReplicationFactorMax: DefaultReplicationFactorMax,
		PinMode:              DefaultPinMode,
	}
	return p
}

func (p *pinner) Add(pin models.Pin) error {
	cid, err := ipfscid.Decode(pin.Cid)
	if err != nil {
		return err
	}
	origins := []clusterapi.Multiaddr{}
	for _, s := range pin.Origins {
		m, err := clusterapi.NewMultiaddr(s)
		if err != nil {
			return err
		}
		origins = append(origins, m)
	}
	_, err = p.ipfsClient.Pin(p.ctx, cid, clusterapi.PinOptions{
		ReplicationFactorMin: p.ReplicationFactorMin,
		ReplicationFactorMax: p.ReplicationFactorMax,
		Name:                 pin.Name,
		Mode:                 p.PinMode,
		ShardSize:            clusterapi.DefaultShardSize,
		Metadata:             pin.Meta,
		Origins:              origins,
	})
	return err
}

func (p *pinner) Remove(cid string) error {
	d, err := ipfscid.Decode(cid)
	if err != nil {
		return err
	}
	_, err = p.ipfsClient.Unpin(p.ctx, d)
	return err
}

func (p *pinner) Delegates() ([]string, error) {
	clientId, err := p.ipfsClient.ID(p.ctx)
	if err != nil {
		return []string{}, err
	}
	delegates := []string{}
	for _, addr := range clientId.IPFS.Addresses {
		v, err := ParseIPFromMultiaddr(addr.Multiaddr)
		if err == nil && !v.IsLoopback() {
			delegates = append(delegates, addr.String())
		}
	}
	return delegates, nil
}

func (p *pinner) SetReplicationFactor(min, max int) error {
	if min > max {
		return fmt.Errorf(
			"Minimum is greater than the maximum replication",
		)
	}
	p.ReplicationFactorMin = min
	p.ReplicationFactorMax = max
	return nil
}

func (p *pinner) SetPinMode(mode clusterapi.PinMode) {
	p.PinMode = mode
}

func ParseIPFromMultiaddr(addr ma.Multiaddr) (net.IP, error) {
	s := addr.String()
	parts := strings.Split(s, "/")
	if parts[0] != "" {
		return nil, ErrInvalidMultiaddrFormat
	}
	if len(parts) < 3 {
		return nil, ErrInvalidMultiaddrFormat
	}
	isip := parts[1] == "ip4" || parts[1] == "ip6"
	if !isip {
		return nil, ErrInvalidMultiaddrFormat
	}
	return net.ParseIP(parts[2]), nil
}
