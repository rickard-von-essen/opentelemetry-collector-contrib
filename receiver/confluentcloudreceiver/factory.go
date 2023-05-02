// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package confluentcloudreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/confluentcloudreceiver"

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const (
	typeStr   = "confluentcloud"
	stability = component.StabilityLevelAlpha
	transport = "http"
)

var errConfigNotConfluentCloudReceveiver = errors.New("config was not a Confluent Cloud receiver config")

// NewFactory creates a new receiver factory
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, stability))
}

func createDefaultConfig() component.Config {
	return &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 60 * time.Second,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: defaultEndpoint,
			Timeout:  60 * time.Second,
		},
		// Metrics: metadata.DefaultMetricsSettings(),
	}
}

func createMetricsReceiver(ctx context.Context, params receiver.CreateSettings, rConf component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	params.Logger.Warn("Starting Confluent Cloud receiver!")
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotConfluentCloudReceveiver
	}

	ccScraper := newCcScraper(params, cfg, params.Logger.Sugar())
	scraper, err := scraperhelper.NewScraper(typeStr, ccScraper.scrape, scraperhelper.WithStart(ccScraper.start))
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewScraperControllerReceiver(&cfg.ScraperControllerSettings, params, consumer, scraperhelper.AddScraper(scraper))
}
