// Copyright (c) 2021 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package peerup provides chain elements to 'up' peer
package peerup

import (
	"context"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	wireguardMech "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/wireguard"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/metadata"
)

type peerupClient struct {
	ctx     context.Context
	vppConn Connection
	sync.Once
}

// NewClient provides a NetworkServiceClient chain elements that 'up's the peer
func NewClient(ctx context.Context, vppConn Connection) networkservice.NetworkServiceClient {
	return &peerupClient{
		ctx:     ctx,
		vppConn: vppConn,
	}
}

func (u *peerupClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	conn, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}

	if mechanism := wireguardMech.ToMechanism(conn.GetMechanism()); mechanism != nil {
		if err := waitForPeerUp(ctx, u.vppConn, mechanism.DstPublicKey(), metadata.IsClient(u)); err != nil {
			_, _ = u.Close(ctx, conn, opts...)
			return nil, err
		}
	}
	return conn, nil
}

func (u *peerupClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	return next.Client(ctx).Close(ctx, conn, opts...)
}