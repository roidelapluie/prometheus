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

package digitalocean

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	conntrack "github.com/mwitkow/go-conntrack"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/version"

	"github.com/prometheus/prometheus/discovery/refresh"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

const (
	doLabel            = model.MetaLabelPrefix + "do_"
	doLabelID          = doLabel + "droplet_id"
	doLabelName        = doLabel + "droplet_name"
	doLabelImage       = doLabel + "image"
	doLabelPrivateIPv4 = doLabel + "private_ipv4"
	doLabelPublicIPv4  = doLabel + "public_ipv4"
	doLabelPublicIPv6  = doLabel + "public_ipv6"
	doLabelRegion      = doLabel + "region"
	doLabelSize        = doLabel + "size"
	doLabelStatus      = doLabel + "status"
	doLabelFeatures    = doLabel + "features"
	doLabelTags        = doLabel + "tags"
	separator          = ","
)

// DefaultSDConfig is the default DigitalOcean SD configuration.
var DefaultSDConfig = SDConfig{
	Port:            80,
	RefreshInterval: model.Duration(60 * time.Second),
}

// SDConfig is the configuration for DO based service discovery.
type SDConfig struct {
	BearerToken     config_util.Secret `yaml:"bearer_token,omitempty"`
	BearerTokenFile string             `yaml:"bearer_token_file,omitempty"`

	RefreshInterval model.Duration `yaml:"refresh_interval,omitempty"`
	Port            int            `yaml:"port"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *SDConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultSDConfig
	type plain SDConfig
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if c.BearerToken == "" && c.BearerTokenFile == "" {
		return fmt.Errorf("bearer_token or bearer_token_file is required")
	}
	if c.BearerToken != "" && c.BearerTokenFile != "" {
		return fmt.Errorf("bearer_token and bearer_token_file are mutually exclusive")
	}
	return nil
}

// Discovery periodically performs DigitalOcean requests. It implements
// the Discoverer interface.
type Discovery struct {
	*refresh.Discovery
	client *godo.Client
	tag    string
	port   int
}

// NewDiscovery returns a new Discovery which periodically refreshes its targets.
func NewDiscovery(conf *SDConfig, logger log.Logger) (*Discovery, error) {
	d := &Discovery{
		port: conf.Port,
	}

	var rt http.RoundTripper = &http.Transport{
		IdleConnTimeout: 5 * time.Duration(conf.RefreshInterval),
		DialContext: conntrack.NewDialContextFunc(
			conntrack.DialWithTracing(),
			conntrack.DialWithName("digitalocean_sd"),
		),
	}
	if conf.BearerToken != "" {
		rt = config_util.NewBearerAuthRoundTripper(conf.BearerToken, rt)
	}
	if conf.BearerTokenFile != "" {
		rt = config_util.NewBearerAuthFileRoundTripper(conf.BearerTokenFile, rt)
	}

	var err error
	d.client, err = godo.New(
		&http.Client{
			Transport: rt,
			Timeout:   5 * time.Duration(conf.RefreshInterval),
		},
		godo.SetUserAgent(fmt.Sprintf("Prometheus/%s", version.Version)),
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up digital ocean agent: %w", err)
	}

	d.Discovery = refresh.NewDiscovery(
		logger,
		"digitalocean",
		time.Duration(conf.RefreshInterval),
		d.refresh,
	)
	return d, nil
}

func (d *Discovery) refresh(ctx context.Context) ([]*targetgroup.Group, error) {
	source := "DO"
	if d.tag != "" {
		source += "_" + d.tag
	}
	tg := &targetgroup.Group{
		Source: source,
	}

	droplets, err := d.listDroplets()
	if err != nil {
		return nil, err
	}
	for _, droplet := range droplets {
		if droplet.Networks == nil || len(droplet.Networks.V4) == 0 {
			continue
		}

		privateIPv4, err := droplet.PrivateIPv4()
		if err != nil {
			return nil, fmt.Errorf("error while reading private IPv4 of droplet %d: %w", droplet.ID, err)
		}
		publicIPv4, err := droplet.PublicIPv4()
		if err != nil {
			return nil, fmt.Errorf("error while reading public IPv4 of droplet %d: %w", droplet.ID, err)
		}
		publicIPv6, err := droplet.PublicIPv6()
		if err != nil {
			return nil, fmt.Errorf("error while reading public IPv6 of droplet %d: %w", droplet.ID, err)
		}

		labels := model.LabelSet{
			doLabelID:          model.LabelValue(fmt.Sprintf("%d", droplet.ID)),
			doLabelName:        model.LabelValue(droplet.Name),
			doLabelImage:       model.LabelValue(droplet.Image.Slug),
			doLabelPrivateIPv4: model.LabelValue(privateIPv4),
			doLabelPublicIPv4:  model.LabelValue(publicIPv4),
			doLabelPublicIPv6:  model.LabelValue(publicIPv6),
			doLabelRegion:      model.LabelValue(droplet.Region.Slug),
			doLabelSize:        model.LabelValue(droplet.SizeSlug),
			doLabelStatus:      model.LabelValue(droplet.Status),
		}

		var addr string
		if publicIPv6 != "" {
			addr = fmt.Sprintf("[%s]", publicIPv6)
		} else if privateIPv4 != "" {
			addr = privateIPv4
		} else if publicIPv4 != "" {
			addr = publicIPv4
		} else {
			return nil, fmt.Errorf("no suitable address for droplet %d: %w", droplet.ID, err)
		}
		addr = fmt.Sprintf("%s:%d", addr, d.port)
		labels[model.AddressLabel] = model.LabelValue(addr)

		if len(droplet.Features) > 0 {
			// We surround the separated list with the separator as well. This way regular expressions
			// in relabeling rules don't have to consider features positions.
			features := separator + strings.Join(droplet.Features, separator) + separator
			labels[doLabelFeatures] = model.LabelValue(features)
		}

		if len(droplet.Tags) > 0 {
			// We surround the separated list with the separator as well. This way regular expressions
			// in relabeling rules don't have to consider tag positions.
			tags := separator + strings.Join(droplet.Tags, separator) + separator
			labels[doLabelTags] = model.LabelValue(tags)
		}

		tg.Targets = append(tg.Targets, labels)
	}
	return []*targetgroup.Group{tg}, nil
}

func (d *Discovery) listDroplets() ([]godo.Droplet, error) {
	var (
		droplets []godo.Droplet
		opts     = &godo.ListOptions{Page: 1}
	)
	for {
		paginatedDroplets, resp, err := d.client.Droplets.List(context.Background(), opts)
		if err != nil {
			return nil, fmt.Errorf("error while listing droplets page %d: %w", opts.Page, err)
		}
		droplets = append(droplets, paginatedDroplets...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opts.Page++
	}
	return droplets, nil
}
