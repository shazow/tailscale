// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tailcfg

import (
	"encoding"
	"reflect"
	"strings"
	"testing"
	"time"

	"inet.af/netaddr"
	"tailscale.com/types/wgkey"
)

func fieldsOf(t reflect.Type) (fields []string) {
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}
	return
}

func TestHostinfoEqual(t *testing.T) {
	hiHandles := []string{
		"IPNVersion", "FrontendLogID", "BackendLogID",
		"OS", "OSVersion", "DeviceModel", "Hostname",
		"ShieldsUp", "ShareeNode",
		"GoArch",
		"RoutableIPs", "RequestTags",
		"Services", "NetInfo",
	}
	if have := fieldsOf(reflect.TypeOf(Hostinfo{})); !reflect.DeepEqual(have, hiHandles) {
		t.Errorf("Hostinfo.Equal check might be out of sync\nfields: %q\nhandled: %q\n",
			have, hiHandles)
	}

	nets := func(strs ...string) (ns []netaddr.IPPrefix) {
		for _, s := range strs {
			n, err := netaddr.ParseIPPrefix(s)
			if err != nil {
				panic(err)
			}
			ns = append(ns, n)
		}
		return ns
	}
	tests := []struct {
		a, b *Hostinfo
		want bool
	}{
		{
			nil,
			nil,
			true,
		},
		{
			&Hostinfo{},
			nil,
			false,
		},
		{
			nil,
			&Hostinfo{},
			false,
		},
		{
			&Hostinfo{},
			&Hostinfo{},
			true,
		},

		{
			&Hostinfo{IPNVersion: "1"},
			&Hostinfo{IPNVersion: "2"},
			false,
		},
		{
			&Hostinfo{IPNVersion: "2"},
			&Hostinfo{IPNVersion: "2"},
			true,
		},

		{
			&Hostinfo{FrontendLogID: "1"},
			&Hostinfo{FrontendLogID: "2"},
			false,
		},
		{
			&Hostinfo{FrontendLogID: "2"},
			&Hostinfo{FrontendLogID: "2"},
			true,
		},

		{
			&Hostinfo{BackendLogID: "1"},
			&Hostinfo{BackendLogID: "2"},
			false,
		},
		{
			&Hostinfo{BackendLogID: "2"},
			&Hostinfo{BackendLogID: "2"},
			true,
		},

		{
			&Hostinfo{OS: "windows"},
			&Hostinfo{OS: "linux"},
			false,
		},
		{
			&Hostinfo{OS: "windows"},
			&Hostinfo{OS: "windows"},
			true,
		},

		{
			&Hostinfo{Hostname: "vega"},
			&Hostinfo{Hostname: "iris"},
			false,
		},
		{
			&Hostinfo{Hostname: "vega"},
			&Hostinfo{Hostname: "vega"},
			true,
		},

		{
			&Hostinfo{RoutableIPs: nil},
			&Hostinfo{RoutableIPs: nets("10.0.0.0/16")},
			false,
		},
		{
			&Hostinfo{RoutableIPs: nets("10.1.0.0/16", "192.168.1.0/24")},
			&Hostinfo{RoutableIPs: nets("10.2.0.0/16", "192.168.2.0/24")},
			false,
		},
		{
			&Hostinfo{RoutableIPs: nets("10.1.0.0/16", "192.168.1.0/24")},
			&Hostinfo{RoutableIPs: nets("10.1.0.0/16", "192.168.2.0/24")},
			false,
		},
		{
			&Hostinfo{RoutableIPs: nets("10.1.0.0/16", "192.168.1.0/24")},
			&Hostinfo{RoutableIPs: nets("10.1.0.0/16", "192.168.1.0/24")},
			true,
		},

		{
			&Hostinfo{RequestTags: []string{"abc", "def"}},
			&Hostinfo{RequestTags: []string{"abc", "def"}},
			true,
		},
		{
			&Hostinfo{RequestTags: []string{"abc", "def"}},
			&Hostinfo{RequestTags: []string{"abc", "123"}},
			false,
		},
		{
			&Hostinfo{RequestTags: []string{}},
			&Hostinfo{RequestTags: []string{"abc"}},
			false,
		},

		{
			&Hostinfo{Services: []Service{Service{Proto: TCP, Port: 1234, Description: "foo"}}},
			&Hostinfo{Services: []Service{Service{Proto: UDP, Port: 2345, Description: "bar"}}},
			false,
		},
		{
			&Hostinfo{Services: []Service{Service{Proto: TCP, Port: 1234, Description: "foo"}}},
			&Hostinfo{Services: []Service{Service{Proto: TCP, Port: 1234, Description: "foo"}}},
			true,
		},
		{
			&Hostinfo{ShareeNode: true},
			&Hostinfo{},
			false,
		},
	}
	for i, tt := range tests {
		got := tt.a.Equal(tt.b)
		if got != tt.want {
			t.Errorf("%d. Equal = %v; want %v", i, got, tt.want)
		}
	}
}

func TestNodeEqual(t *testing.T) {
	nodeHandles := []string{
		"ID", "StableID", "Name", "DisplayName", "User", "Sharer",
		"Key", "KeyExpiry", "Machine", "DiscoKey",
		"Addresses", "AllowedIPs", "Endpoints", "DERP", "Hostinfo",
		"Created", "LastSeen", "KeepAlive", "MachineAuthorized",
	}
	if have := fieldsOf(reflect.TypeOf(Node{})); !reflect.DeepEqual(have, nodeHandles) {
		t.Errorf("Node.Equal check might be out of sync\nfields: %q\nhandled: %q\n",
			have, nodeHandles)
	}

	newPublicKey := func(t *testing.T) wgkey.Key {
		t.Helper()
		k, err := wgkey.NewPrivate()
		if err != nil {
			t.Fatal(err)
		}
		return k.Public()
	}
	n1 := newPublicKey(t)
	now := time.Now()

	tests := []struct {
		a, b *Node
		want bool
	}{
		{
			&Node{},
			nil,
			false,
		},
		{
			nil,
			&Node{},
			false,
		},
		{
			&Node{},
			&Node{},
			true,
		},
		{
			&Node{},
			&Node{},
			true,
		},
		{
			&Node{ID: 1},
			&Node{},
			false,
		},
		{
			&Node{ID: 1},
			&Node{ID: 1},
			true,
		},
		{
			&Node{StableID: "node-abcd"},
			&Node{},
			false,
		},
		{
			&Node{StableID: "node-abcd"},
			&Node{StableID: "node-abcd"},
			true,
		},
		{
			&Node{User: 0},
			&Node{User: 1},
			false,
		},
		{
			&Node{User: 1},
			&Node{User: 1},
			true,
		},
		{
			&Node{Key: NodeKey(n1)},
			&Node{Key: NodeKey(newPublicKey(t))},
			false,
		},
		{
			&Node{Key: NodeKey(n1)},
			&Node{Key: NodeKey(n1)},
			true,
		},
		{
			&Node{KeyExpiry: now},
			&Node{KeyExpiry: now.Add(60 * time.Second)},
			false,
		},
		{
			&Node{KeyExpiry: now},
			&Node{KeyExpiry: now},
			true,
		},
		{
			&Node{Machine: MachineKey(n1)},
			&Node{Machine: MachineKey(newPublicKey(t))},
			false,
		},
		{
			&Node{Machine: MachineKey(n1)},
			&Node{Machine: MachineKey(n1)},
			true,
		},
		{
			&Node{Addresses: []netaddr.IPPrefix{}},
			&Node{Addresses: nil},
			false,
		},
		{
			&Node{Addresses: []netaddr.IPPrefix{}},
			&Node{Addresses: []netaddr.IPPrefix{}},
			true,
		},
		{
			&Node{AllowedIPs: []netaddr.IPPrefix{}},
			&Node{AllowedIPs: nil},
			false,
		},
		{
			&Node{Addresses: []netaddr.IPPrefix{}},
			&Node{Addresses: []netaddr.IPPrefix{}},
			true,
		},
		{
			&Node{Endpoints: []string{}},
			&Node{Endpoints: nil},
			false,
		},
		{
			&Node{Endpoints: []string{}},
			&Node{Endpoints: []string{}},
			true,
		},
		{
			&Node{Hostinfo: Hostinfo{Hostname: "alice"}},
			&Node{Hostinfo: Hostinfo{Hostname: "bob"}},
			false,
		},
		{
			&Node{Hostinfo: Hostinfo{}},
			&Node{Hostinfo: Hostinfo{}},
			true,
		},
		{
			&Node{Created: now},
			&Node{Created: now.Add(60 * time.Second)},
			false,
		},
		{
			&Node{Created: now},
			&Node{Created: now},
			true,
		},
		{
			&Node{LastSeen: &now},
			&Node{LastSeen: nil},
			false,
		},
		{
			&Node{LastSeen: &now},
			&Node{LastSeen: &now},
			true,
		},
		{
			&Node{DERP: "foo"},
			&Node{DERP: "bar"},
			false,
		},
	}
	for i, tt := range tests {
		got := tt.a.Equal(tt.b)
		if got != tt.want {
			t.Errorf("%d. Equal = %v; want %v", i, got, tt.want)
		}
	}
}

func TestNetInfoFields(t *testing.T) {
	handled := []string{
		"MappingVariesByDestIP",
		"HairPinning",
		"WorkingIPv6",
		"WorkingUDP",
		"UPnP",
		"PMP",
		"PCP",
		"PreferredDERP",
		"LinkType",
		"DERPLatency",
	}
	if have := fieldsOf(reflect.TypeOf(NetInfo{})); !reflect.DeepEqual(have, handled) {
		t.Errorf("NetInfo.Clone/BasicallyEqually check might be out of sync\nfields: %q\nhandled: %q\n",
			have, handled)
	}
}

func TestMachineKeyMarshal(t *testing.T) {
	var k1, k2 MachineKey
	for i := range k1 {
		k1[i] = byte(i)
	}
	testKey(t, "mkey:", k1, &k2)
}

func TestNodeKeyMarshal(t *testing.T) {
	var k1, k2 NodeKey
	for i := range k1 {
		k1[i] = byte(i)
	}
	testKey(t, "nodekey:", k1, &k2)
}

func TestDiscoKeyMarshal(t *testing.T) {
	var k1, k2 DiscoKey
	for i := range k1 {
		k1[i] = byte(i)
	}
	testKey(t, "discokey:", k1, &k2)
}

type keyIn interface {
	String() string
	MarshalText() ([]byte, error)
}

func testKey(t *testing.T, prefix string, in keyIn, out encoding.TextUnmarshaler) {
	got, err := in.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if err := out.UnmarshalText(got); err != nil {
		t.Fatal(err)
	}
	if s := in.String(); string(got) != s {
		t.Errorf("MarshalText = %q != String %q", got, s)
	}
	if !strings.HasPrefix(string(got), prefix) {
		t.Errorf("%q didn't start with prefix %q", got, prefix)
	}
	if reflect.ValueOf(out).Elem().Interface() != in {
		t.Errorf("mismatch after unmarshal")
	}
}

func TestCloneUser(t *testing.T) {
	tests := []struct {
		name string
		u    *User
	}{
		{"nil_logins", &User{}},
		{"zero_logins", &User{Logins: make([]LoginID, 0)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u2 := tt.u.Clone()
			if !reflect.DeepEqual(tt.u, u2) {
				t.Errorf("not equal")
			}
		})
	}
}

func TestCloneNode(t *testing.T) {
	tests := []struct {
		name string
		v    *Node
	}{
		{"nil_fields", &Node{}},
		{"zero_fields", &Node{
			Addresses:  make([]netaddr.IPPrefix, 0),
			AllowedIPs: make([]netaddr.IPPrefix, 0),
			Endpoints:  make([]string, 0),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v2 := tt.v.Clone()
			if !reflect.DeepEqual(tt.v, v2) {
				t.Errorf("not equal")
			}
		})
	}
}
