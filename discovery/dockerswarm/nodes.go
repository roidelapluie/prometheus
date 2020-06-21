// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dockerswarm

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/prometheus/common/model"

	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/strutil"
)

const (
	swarmLabelNodePrefix               = swarmLabel + "node_"
	swarmLabelNodeAddress              = swarmLabelNodePrefix + "address"
	swarmLabelNodeAvailability         = swarmLabelNodePrefix + "availability"
	swarmLabelNodeEngineVersion        = swarmLabelNodePrefix + "engine_version"
	swarmLabelNodeHostname             = swarmLabelNodePrefix + "hostname"
	swarmLabelNodeID                   = swarmLabelNodePrefix + "id"
	swarmLabelNodeLabelPrefix          = swarmLabelNodePrefix + "label_"
	swarmLabelNodeManagerAddr          = swarmLabelNodePrefix + "manager_address"
	swarmLabelNodeManagerLeader        = swarmLabelNodePrefix + "manager_leader"
	swarmLabelNodeManagerReachability  = swarmLabelNodePrefix + "manager_reachability"
	swarmLabelNodePlatformArchitecture = swarmLabelNodePrefix + "platform_architecture"
	swarmLabelNodePlatformOS           = swarmLabelNodePrefix + "platform_os"
	swarmLabelNodeRole                 = swarmLabelNodePrefix + "role"
	swarmLabelNodeStatus               = swarmLabelNodePrefix + "status"
)

func (d *Discovery) refreshNodes(ctx context.Context) ([]*targetgroup.Group, error) {
	tg := &targetgroup.Group{
		Source: "DockerSwarm",
	}

	nodes, err := d.client.NodeList(ctx, types.NodeListOptions{Filters: makeArgs(d.filters)})
	if err != nil {
		return nil, fmt.Errorf("error while listing swarm nodes: %w", err)
	}

	for _, n := range nodes {
		labels := model.LabelSet{
			swarmLabelNodeID:                   model.LabelValue(n.ID),
			swarmLabelNodeRole:                 model.LabelValue(n.Spec.Role),
			swarmLabelNodeAvailability:         model.LabelValue(n.Spec.Availability),
			swarmLabelNodeHostname:             model.LabelValue(n.Description.Hostname),
			swarmLabelNodePlatformArchitecture: model.LabelValue(n.Description.Platform.Architecture),
			swarmLabelNodePlatformOS:           model.LabelValue(n.Description.Platform.OS),
			swarmLabelNodeEngineVersion:        model.LabelValue(n.Description.Engine.EngineVersion),
			swarmLabelNodeStatus:               model.LabelValue(n.Status.State),
			swarmLabelNodeAddress:              model.LabelValue(n.Status.Addr),
		}
		if n.ManagerStatus != nil {
			labels[model.LabelName(swarmLabelNodeManagerLeader)] = model.LabelValue(fmt.Sprintf("%v", n.ManagerStatus.Leader))
			labels[model.LabelName(swarmLabelNodeManagerReachability)] = model.LabelValue(n.ManagerStatus.Reachability)
			labels[model.LabelName(swarmLabelNodeManagerAddr)] = model.LabelValue(n.ManagerStatus.Addr)
		}

		for k, v := range n.Spec.Labels {
			ln := strutil.SanitizeLabelName(k)
			labels[model.LabelName(swarmLabelNodeLabelPrefix+ln)] = model.LabelValue(v)
		}

		addr := net.JoinHostPort(n.Status.Addr, strconv.FormatUint(uint64(d.port), 10))
		labels[model.AddressLabel] = model.LabelValue(addr)

		tg.Targets = append(tg.Targets, labels)

	}
	return []*targetgroup.Group{tg}, nil
}

func (d *Discovery) getNodeLabels(ctx context.Context, nodeID string) (model.LabelSet, error) {
	n, _, err := d.client.NodeInspectWithRaw(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	labels := model.LabelSet{
		swarmLabelNodeID:                   model.LabelValue(n.ID),
		swarmLabelNodeRole:                 model.LabelValue(n.Spec.Role),
		swarmLabelNodeAddress:              model.LabelValue(n.Status.Addr),
		swarmLabelNodeAvailability:         model.LabelValue(n.Spec.Availability),
		swarmLabelNodeHostname:             model.LabelValue(n.Description.Hostname),
		swarmLabelNodePlatformArchitecture: model.LabelValue(n.Description.Platform.Architecture),
		swarmLabelNodePlatformOS:           model.LabelValue(n.Description.Platform.OS),
		swarmLabelNodeStatus:               model.LabelValue(n.Status.State),
	}
	for k, v := range n.Spec.Labels {
		ln := strutil.SanitizeLabelName(k)
		labels[model.LabelName(swarmLabelNodeLabelPrefix+ln)] = model.LabelValue(v)
	}
	return labels, nil
}
