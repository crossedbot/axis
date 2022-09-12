package pinner

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	clusterapi "github.com/ipfs-cluster/ipfs-cluster/api"
	ipfscid "github.com/ipfs/go-cid"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/crossedbot/axis/pkg/mocks"
	"github.com/crossedbot/axis/pkg/pins/models"
)

func TestPinnerAdd(t *testing.T) {
	ctx := context.Background()
	pin := models.Pin{
		Cid:  "QmPAwR5un1YPJEF6iB7KvErDmAhiXxwL5J5qjA3Z9ceKqv",
		Name: "helloworld",
		Meta: models.Info{"uid": "myuserid"},
	}
	cid, err := ipfscid.Decode(pin.Cid)
	require.Nil(t, err)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIpfsClient := mocks.NewMockClient(mockCtrl)
	mockIpfsClient.EXPECT().
		Pin(ctx, clusterapi.NewCid(cid), clusterapi.PinOptions{
			ReplicationFactorMin: DefaultReplicationFactorMin,
			ReplicationFactorMax: DefaultReplicationFactorMax,
			Name:                 pin.Name,
			Mode:                 DefaultPinMode,
			ShardSize:            clusterapi.DefaultShardSize,
			Metadata:             pin.Meta,
			Origins:              []clusterapi.Multiaddr{},
		})
	p := &pinner{
		ctx:                  ctx,
		ipfsClient:           mockIpfsClient,
		ReplicationFactorMin: DefaultReplicationFactorMin,
		ReplicationFactorMax: DefaultReplicationFactorMax,
		PinMode:              DefaultPinMode,
	}
	err = p.Add(pin)
	require.Nil(t, err)
}

func TestPinnerRemove(t *testing.T) {
	ctx := context.Background()
	cidStr := "QmPAwR5un1YPJEF6iB7KvErDmAhiXxwL5J5qjA3Z9ceKqv"
	cid, err := ipfscid.Decode(cidStr)
	require.Nil(t, err)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIpfsClient := mocks.NewMockClient(mockCtrl)
	mockIpfsClient.EXPECT().
		Unpin(ctx, clusterapi.NewCid(cid))
	p := &pinner{
		ctx:        ctx,
		ipfsClient: mockIpfsClient,
	}
	err = p.Remove(cidStr)
	require.Nil(t, err)
}

func TestDelegates(t *testing.T) {
	ctx := context.Background()
	maStr := "/ip4/172.0.3.102/tcp/4001/p2p/QmPAwR5un1YPJEF6iB7KvErDmAhiXxwL5J5qjA3Z9ceKqv"
	multiaddr, err := ma.NewMultiaddr(maStr)
	require.Nil(t, err)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIpfsClient := mocks.NewMockClient(mockCtrl)
	mockIpfsClient.EXPECT().
		ID(ctx).
		Return(clusterapi.ID{
			IPFS: clusterapi.IPFSID{
				Addresses: []clusterapi.Multiaddr{
					{Multiaddr: multiaddr},
				},
			},
		}, nil)
	p := &pinner{
		ctx:        ctx,
		ipfsClient: mockIpfsClient,
	}
	delegates, err := p.Delegates()
	require.Nil(t, err)
	require.Equal(t, []string{maStr}, delegates)
}

func TestSetReplicationFactor(t *testing.T) {
	p := &pinner{
		ReplicationFactorMin: DefaultReplicationFactorMin,
		ReplicationFactorMax: DefaultReplicationFactorMax,
	}

	min, max := 3, 9
	err := p.SetReplicationFactor(min, max)
	require.Nil(t, err)
	require.Equal(t, min, p.ReplicationFactorMin)
	require.Equal(t, max, p.ReplicationFactorMax)

	min, max = 9, 3
	err = p.SetReplicationFactor(min, max)
	require.NotNil(t, err)
}

func TestSetPinMode(t *testing.T) {
	p := &pinner{PinMode: DefaultPinMode}
	expected := clusterapi.PinModeDirect
	p.SetPinMode(expected)
	require.Equal(t, expected, p.PinMode)
}

func TestParseIPFromMultiaddr(t *testing.T) {
	expected := "172.0.2.100"
	multiaddr := fmt.Sprintf(
		"/ip4/%s/tcp/4001/p2p/QmPAwR5un1YPJEF6iB7KvErDmAhiXxwL5J5qjA3Z9ceKqv",
		expected,
	)
	ip4Multiaddr, err := ma.NewMultiaddr(multiaddr)
	require.Nil(t, err)
	actual, err := ParseIPFromMultiaddr(ip4Multiaddr)
	require.Nil(t, err)
	require.Equal(t, expected, actual.String())

	expected = "2345:425:2ca1::567:5673:23b5"
	multiaddr = fmt.Sprintf(
		"/ip6/%s/tcp/4001/p2p/QmPAwR5un1YPJEF6iB7KvErDmAhiXxwL5J5qjA3Z9ceKqv",
		expected,
	)
	ip6Multiaddr, err := ma.NewMultiaddr(multiaddr)
	require.Nil(t, err)
	actual, err = ParseIPFromMultiaddr(ip6Multiaddr)
	require.Nil(t, err)
	require.Equal(t, expected, actual.String())
}
