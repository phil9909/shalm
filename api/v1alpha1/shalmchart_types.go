/*

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

package v1alpha1

import (
	"bytes"
	"encoding/gob"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClonableMap -
type ClonableMap map[string]interface{}

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

// DeepCopy -
func (v *ClonableMap) DeepCopy() *ClonableMap {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		panic(err)
	}
	var copy map[string]interface{}
	err = dec.Decode(&copy)
	if err != nil {
		panic(err)
	}
	result := ClonableMap(copy)
	return &result
}

// ClonableArray -
type ClonableArray []interface{}

// DeepCopy -
func (v *ClonableArray) DeepCopy() *ClonableArray {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		panic(err)
	}
	var copy []interface{}
	err = dec.Decode(&copy)
	if err != nil {
		panic(err)
	}
	result := ClonableArray(copy)
	return &result
}

// ChartSpec defines the desired state of ShalmChart
type ChartSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Values     ClonableMap   `json:"values,omitempty"`
	Args       ClonableArray `json:"args,omitempty"`
	KwArgs     ClonableMap   `json:"kwargs,omitempty"`
	KubeConfig string        `json:"kubeconfig,omitempty"`
	Namespace  string        `json:"namespace,omitempty"`
	Suffix     string        `json:"suffix,omitempty"`
	ChartTgz   []byte        `json:"chart_tgz,omitempty"`
}

// Operation defines the progress of the last operation
type Operation struct {
	Type     string `json:"type,omitempty"`
	Progress int    `json:"progress,omitempty"`
}

// ChartStatus defines the observed state of ShalmChart
type ChartStatus struct {
	LastOp Operation `json:"lastOp,omitempty"`
}

// +kubebuilder:object:root=true

// ShalmChart is the Schema for the shalmcharts API
type ShalmChart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChartSpec   `json:"spec,omitempty"`
	Status ChartStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ShalmChartList contains a list of ShalmChart
type ShalmChartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShalmChart `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShalmChart{}, &ShalmChartList{})
}
