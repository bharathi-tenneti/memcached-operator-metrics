/*
Copyright 2020 The Operator-SDK Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type CRInfoGauge struct {
	*prometheus.GaugeVec
}

type SizeInfo struct {
	*prometheus.GaugeVec
}

type SummaryInfo struct {
	*prometheus.GaugeVec
}

func NewSummaryInfo() *SummaryInfo {
	return &SummaryInfo{
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "summary_info",
			Help: fmt.Sprintf("Information about the custom resource summary."),
		}, []string{"namespace", "name", "apiversion", "kind"}),
	}
}
func NewSizeInfo() *SizeInfo {
	return &SizeInfo{
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "size_info",
			Help: fmt.Sprintf("Information about the custom resources size"),
		}, []string{"namespace", "name"}),
	}
}

func NewCRInfoGauge() *CRInfoGauge {
	return &CRInfoGauge{
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "custom_resource_info",
			Help: fmt.Sprintf("Information about the custom resources"),
		}, []string{"namespace", "name", "created"}),
	}
}

func (vec *CRInfoGauge) Create(e event.CreateEvent) {
	name := e.Meta.GetName()
	namespace := e.Meta.GetNamespace()
	created := e.Meta.GetCreationTimestamp().String()
	vec.set(name, namespace, created)
}

func (vec *CRInfoGauge) Update(e event.UpdateEvent) {
	name := e.MetaNew.GetName()
	namespace := e.MetaNew.GetNamespace()
	created := e.MetaNew.GetCreationTimestamp().String()
	vec.set(name, namespace, created)
}

func (vec *CRInfoGauge) Delete(e event.DeleteEvent) {
	vec.GaugeVec.Delete(map[string]string{
		"name":      e.Meta.GetName(),
		"namespace": e.Meta.GetNamespace(),
		"created":   e.Meta.GetCreationTimestamp().String(),
	})
}

func (vec *CRInfoGauge) set(name, namespace, created string) {
	labels := map[string]string{
		"name":      name,
		"namespace": namespace,
		"created":   created,
	}
	m, err := vec.GaugeVec.GetMetricWith(labels)
	if err != nil {
		panic(err)
	}
	m.Set(1)
}

func NewDefaultRegistry() RegistererGathererPredicater {
	r := NewRegistry()
	return r
}
