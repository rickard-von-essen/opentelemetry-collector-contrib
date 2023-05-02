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

import ()

type Page struct {
  Data []MetricDescriptor `json:"data"`
  Meta PageMeta `json:"meta"`
  Links map[string]string `json:"links"` // Ignore
}

type PageMeta struct {
  Pagination Pagination `json:"pagination"`
}

type Pagination struct {
  NextPageToken string `json:"next_page_token"`
  PageSize int `json:"page_size"`
  TotalSize int `json:"total_size"`
}

type MetricDescriptor struct {
  Description string `json:"description"`
  Exportable bool `json:"exportable"`
  Labels []MetricLabel `json:"labels"`
  LifecycleStage string `json:"lifecycle_stage"`
  Name string `json:"name"`
  Resources []string `json:"resources"`
  Type string `json:"type"`
  Unit string `json:"unit"`
}

type MetricLabel struct {
  Description string `json:"description"`
  Exportable bool `json:"exportable"`
  Key string `json:"key"`
}

// Is there more pages to get?
func (p *Page) HasMorePages() bool {
  return p.Meta.Pagination.NextPageToken != ""
}
