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

type SummaryInfo struct {
	*prometheus.SummaryVec
}

type CounterInfo struct {
	*prometheus.CounterVec
}

func NewCounterInfo() *CounterInfo {
	return &CounterInfo{
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "counter_info",
			Help: fmt.Sprintf("Information about the custom resources partitions."),
		}, []string{"namespace", "name"}),
	}
}
func NewSummaryInfo() *SummaryInfo {
	return &SummaryInfo{
		prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "summary_info",
			Help: fmt.Sprintf("Information about the custom resources size"),
		}, []string{"namespace", "name", "size"}),
	}
}

func NewCRInfoGauge() *CRInfoGauge {
	return &CRInfoGauge{
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "custom_resource_info",
			Help: fmt.Sprintf("Information about the custom resources"),
		}, []string{"namespace", "name", "group", "kind"}),
	}
}

func (vec *CRInfoGauge) Create(e event.CreateEvent) {
	name := e.Meta.GetName()
	namespace := e.Meta.GetNamespace()
	group := e.Object.GetObjectKind().GroupVersionKind().Group
	kind := e.Object.GetObjectKind().GroupVersionKind().Kind
	vec.set(name, namespace, group, kind)
}

func (vec *CRInfoGauge) Update(e event.UpdateEvent) {
	name := e.MetaNew.GetName()
	namespace := e.MetaNew.GetNamespace()
	group := e.ObjectNew.GetObjectKind().GroupVersionKind().Group
	kind := e.ObjectNew.GetObjectKind().GroupVersionKind().Kind
	vec.set(name, namespace, group, kind)
}

func (vec *CRInfoGauge) Delete(e event.DeleteEvent) {
	vec.GaugeVec.Delete(map[string]string{
		"name":      e.Meta.GetName(),
		"namespace": e.Meta.GetNamespace(),
		"group":     e.Object.GetObjectKind().GroupVersionKind().Group,
		"kind":      e.Object.GetObjectKind().GroupVersionKind().Kind,
	})
}

func (vec *CRInfoGauge) set(name, namespace, group, kind string) {
	labels := map[string]string{
		"name":      name,
		"namespace": namespace,
		"group":     group,
		"kind":      kind,
	}
	m, err := vec.GaugeVec.GetMetricWith(labels)
	if err != nil {
		panic(err)
	}
	m.Set(1)
}

func NewDefaultRegistry() RegistererGathererPredicater {
	crInfo := NewCRInfoGauge()
	summaryInfo := NewSummaryInfo()
	counterInfo := NewCounterInfo()
	r := NewRegistry()
	r.MustRegister(crInfo)
	r.MustRegister(summaryInfo)
	r.MustRegister(counterInfo)
	return r
}
