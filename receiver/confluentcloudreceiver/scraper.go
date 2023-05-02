// Copyright 2020, OpenTelemetry Authors
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type ccScraper struct {
	httpClient *http.Client

	settings component.TelemetrySettings
	cfg      *Config
	logger   *zap.SugaredLogger

  descriptors *[]MetricDescriptor
}

func newCcScraper(settings receiver.CreateSettings, cfg *Config, logger *zap.SugaredLogger) *ccScraper {
	return &ccScraper{
		settings: settings.TelemetrySettings,
		cfg:      cfg,
		logger:   logger,
    descriptors: nil,
	}
}

func (r *ccScraper) start(_ context.Context, host component.Host) error {

	r.logger.Warnf("Starting Confluent Cloud Receiver")

	httpClient, err := r.cfg.ToClient(host, r.settings)
	if err != nil {
		return err
	}
	r.httpClient = httpClient

	resp, err := r.httpClient.Get(r.cfg.Endpoint + "/descriptors/metrics?resource_type=kafka&page_size=1000") // TODO resource_type in config and pagination
	if err != nil {
		r.logger.Errorf("Error getting descriptors %s - %s", resp.Status, err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		r.logger.Errorf("Error getting descriptors %s - %s", resp.Status, err.Error())
    return fmt.Errorf("Error getting descriptors from %s got: %s", resp.Request.URL, resp.Status)
	}
  
  defer resp.Body.Close()
  descriptorsPage := Page{}

  err = json.NewDecoder(resp.Body).Decode(&descriptorsPage)
	if err != nil {
    r.logger.Errorf("Error parsing descriptors: %s", err.Error())
		return err
	}
  r.descriptors = &descriptorsPage.Data

  for _, desc := range descriptorsPage.Data {
    r.logger.Debugf("Metric: %s", desc.Name)    
  }
  r.descriptors = &descriptorsPage.Data
	return nil
}

func (r *ccScraper) scrape(context.Context) (pmetric.Metrics, error) {

  metrics, err := r.getSingleMetric("io.confluent.kafka.server/sent_bytes") // TODO use include / exclude from discovery

  return metrics, err
}

func (r *ccScraper) getSingleMetric(metricName string) (pmetric.Metrics, error) {

  query := MetricQuery{
    Aggregations: &[]Aggregation{{Metric: metricName}}, 
    GroupBy: []string{"metric.topic", "metric.partition"}, // TODO get these from discovery doc
    Filter: &Filter{
      Field: "resource.kafka.id",
      Op: "EQ",
      Value: r.cfg.Resources.Kafka[0],
    },
    Granularity: "ALL",
    Intervals: []string{"now-3m|m/now-1m|m"},
  }

  body, err := json.Marshal(query)
  if err != nil {
    r.logger.Fatalf("impossible to marshall query: %s", err)
  }
  r.logger.Debugf("Body: %s", body)

  req, err := http.NewRequest("POST", r.cfg.Endpoint + "/query", bytes.NewReader(body)) // TODO ?page_token= pagination

  if err != nil {
    r.logger.Fatalf("impossible to build request: %s", err)
  }

  req.Header.Set("Content-Type", "application/json")

  resp, err := r.httpClient.Do(req)
  if err != nil {
    r.logger.Fatalf("impossible to send request: %s", err  )
  }
  r.logger.Debugf("status Code: %d", resp.StatusCode)

  defer resp.Body.Close()
  /*
  buf := new(strings.Builder)
  _, err = io.Copy(buf, resp.Body)
  r.logger.Debugf("Response: %s", buf.String())
  */

  metrics := pmetric.NewMetrics() // TODO use the right type 
  m := pmetric.NewMetric() // TODO move some parts of this outside func
	m.SetName(metricName)
	m.SetDescription("The delta count of bytes of the customer's data sent over the network. Each sample is the number of bytes sent since the previous data point. The count is sampled every 60 seconds.")
	m.SetUnit("By")
	m.SetEmptySum()
	m.Sum().SetIsMonotonic(false)
	m.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

  err = r.parseMetrics(&m, r.cfg.Resources.Kafka[0], &resp.Body)
  if err != nil {
    r.logger.Fatalf("error parsing: %s", err  )
  }

	rm := pmetric.NewResourceMetrics()
	// rm.Resource().Attributes().EnsureCapacity(mb.resourceCapacity)
	ils := rm.ScopeMetrics().AppendEmpty()
	ils.Scope().SetName("otelcol/confluentcloudreceiver")
	// ils.Scope().SetVersion(mb.buildInfo.Version) // TODO

 	m.MoveTo(ils.Metrics().AppendEmpty())
	rm.MoveTo(metrics.ResourceMetrics().AppendEmpty())

  return metrics, nil
}

/*
{
  "data": [
    {
      "timestamp": "2023-04-01T14:58:00Z",
      "value": 139118,
      "metric.topic": "event.something.log"
    },
    {
      "timestamp": "2023-04-01T14:58:00Z",
      "value": 105152,
      "metric.topic": "event.somethingelse.log"
    }
  ]
}
*/

type MetricsResponse struct {
	Data []map[string]interface{} `json:"data"`
}

func (r *ccScraper) parseMetrics(m *pmetric.Metric, kafkaId string, body *io.ReadCloser) error {

  var metrics MetricsResponse

  decoder := json.NewDecoder(*body)
  err := decoder.Decode(&metrics)
  if err != nil && err != io.EOF {
    return err
  }

  r.logger.Infof("Resp: %v", metrics)

  for _, v := range metrics.Data {

	  dp := m.Sum().DataPoints().AppendEmpty()
	  // dp.SetStartTimestamp(start)
    ts, _ := time.Parse(time.RFC3339, v["timestamp"].(string)) // TODO error handeling
	  dp.SetTimestamp(pcommon.NewTimestampFromTime(ts))
    attr := dp.Attributes()
    attr.PutStr("kafka.id", kafkaId)

    for label, value := range v {
      if label != "timestamp" && label != "value" {
        attr.PutStr(label, value.(string))
      }  
    }
	  dp.SetDoubleValue(v["value"].(float64)) // TODO pick right type from desc doc
  }

  return nil
}
