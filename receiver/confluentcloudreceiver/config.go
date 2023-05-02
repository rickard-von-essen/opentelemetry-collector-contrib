// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package confluentcloudreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/confluentcloudreceiver"

import (
	"errors"
	"fmt"
	"net/url"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

// Predefined error responses for configuration validation failures
var (
	errInvalidEndpoint = errors.New(`"endpoint" must be in the form of <scheme>://<hostname>:<port>`)
)

const defaultEndpoint = "https://api.telemetry.confluent.cloud/v2/metrics/cloud"

// Config defines the configuration for the various elements of the receiver agent.
type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
  ConfluentMetrics                        `mapstructure:",squash"`
	// Metrics                                 metadata.MetricsSettings `mapstructure:"metrics"` TODO
	// Create metadata from discorey doc. ?
	// Filter metrics include, exclude RE
}

type ConfluentMetrics struct {
  Resources Resources `mapstructure:"resources"`
  Filter MetricsFilter `mapstructure:"filter"`
}

type Resources struct {
  Kafka []string `mapstructure:"kafka"` // TODO validate there is exactly one of these!
}

type MetricsFilter struct {
  Include []string `mapstructure:"include"`
  Exclude []string `mapstructure:"exclude"`
}

// Validate validates the configuration by checking for missing or invalid fields
func (cfg *Config) Validate() error {
	var err error

	_, parseErr := url.Parse(cfg.Endpoint)
	if parseErr != nil {
		wrappedErr := fmt.Errorf("%s: %w", errInvalidEndpoint.Error(), parseErr)
		err = multierr.Append(err, wrappedErr)
	}

	return err
}
