/*
Copyright 2019 The Kanister Authors.

Copyright 2017 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Some of the code below came from https://github.com/coreos/etcd-operator
which also has the apache 2.0 license.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	sp "k8s.io/apimachinery/pkg/util/strategicpatch"
)

const (
	// SchemeVersion is the API version of objects in this package.
	SchemeVersion = "v1alpha1"
	// ResourceGroup is the API group of resources in this package.
	ResourceGroup = "cr.kanister.io"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: ResourceGroup, Version: SchemeVersion}

// JSONMap contains PodOverride specs.
type JSONMap sp.JSONMap

var _ runtime.Object = (*ActionSet)(nil)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActionSet describes kanister actions.
type ActionSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *ActionSetSpec   `json:"spec,omitempty"`
	Status            *ActionSetStatus `json:"status,omitempty"`
}

// ObjectReference refers to a kubernetes object.
type ObjectReference struct {
	// API version of the referent.
	APIVersion string `json:"apiVersion"`
	// API Group of the referent.
	Group string `json:"group"`
	// Resource name of the referent.
	Resource string `json:"resource"`
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	Kind string `json:"kind"`
	// Name of the referent.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name"`
	// Namespace of the referent.
	// More info: http://kubernetes.io/docs/user-guide/namespaces
	Namespace string `json:"namespace,omitempty"`
}

// ActionSetSpec is the specification for the actionset.
type ActionSetSpec struct {
	Actions []ActionSpec `json:"actions,omitempty"`
}

// ActionSpec is the specification for a single Action.
type ActionSpec struct {
	// Name is the action we'll perform. For example: `backup` or `restore`.
	Name string `json:"name"`
	// Object refers to the thing we'll perform this action on.
	Object ObjectReference `json:"object"`
	// Blueprint with instructions on how to execute this action.
	Blueprint string `json:"blueprint,omitempty"`
	// Artifacts will be passed as inputs into this phase.
	Artifacts map[string]Artifact `json:"artifacts,omitempty"`
	// ConfigMaps that we'll get and pass into the blueprint.
	ConfigMaps map[string]ObjectReference `json:"configMaps,omitempty"`
	// Secrets that we'll get and pass into the blueprint.
	Secrets map[string]ObjectReference `json:"secrets,omitempty"`
	// Profile is use to specify the location where store artifacts and the
	// credentials authorized to access them.
	Profile *ObjectReference `json:"profile,omitempty"`
	// PodOverride is used to specify pod specs that will override the
	// default pod specs
	PodOverride JSONMap `json:"podOverride,omitempty"`
	// Options will be used to specify additional values
	// to be used in the Blueprint.
	Options map[string]string `json:"options,omitempty"`
	// PreferredVersion will be used to select the preferred version of Kanister functions
	// to be executed for this action
	PreferredVersion string `json:"preferredVersion"`
}

// ActionSetStatus is the status for the actionset. This should only be updated by the controller.
type ActionSetStatus struct {
	State    State          `json:"state"`
	Actions  []ActionStatus `json:"actions,omitempty"`
	Error    Error          `json:"error,omitempty"`
	Progress ActionProgress `json:"progress,omitempty"`
}

// ActionStatus is updated as we execute phases.
type ActionStatus struct {
	// Name is the action we'll perform. For example: `backup` or `restore`.
	Name string `json:"name"`
	// Object refers to the thing we'll perform this action on.
	Object ObjectReference `json:"object"`
	// Blueprint with instructions on how to execute this action.
	Blueprint string `json:"blueprint"`
	// Phases are sub-actions an are executed sequentially.
	Phases []Phase `json:"phases,omitempty"`
	// Artifacts created by this phase.
	Artifacts map[string]Artifact `json:"artifacts,omitempty"`
	// DeferPhase is the phase that is executed at the end of an action
	// irrespective of the status of other phases in the action
	DeferPhase Phase `json:"deferPhase,omitempty"`
}

// ActionProgress provides information on the progress of an action.
type ActionProgress struct {
	// RunningPhase represents which phase of the action is being run
	RunningPhase string `json:"runningPhase,omitempty"`
	// PercentCompleted is computed by assessing the number of completed phases
	// against the the total number of phases.
	PercentCompleted string `json:"percentCompleted,omitempty"`
	// LastTransitionTime represents the last date time when the progress status
	// was received.
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// State is the current state of a phase of execution.
type State string

const (
	// StatePending mean this action or phase has yet to be executed.
	StatePending State = "pending"
	// StateRunning means this action or phase is currently executing.
	StateRunning State = "running"
	// StateFailed means this action or phase was unsuccessful.
	StateFailed State = "failed"
	// StateComplete means this action or phase finished successfully.
	StateComplete State = "complete"
)

type Error struct {
	Message string `json:"message"`
}

// Phase is subcomponent of an action.
type Phase struct {
	Name   string                 `json:"name"`
	State  State                  `json:"state"`
	Output map[string]interface{} `json:"output,omitempty"`
}

// k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Artifact tracks objects produced by an action.
type Artifact struct {
	KeyValue map[string]string `json:"keyValue,omitempty"`
	// KopiaSnapshot captures the kopia snapshot information
	// produced as a JSON string by kando command in phases of an action.
	KopiaSnapshot string `json:"kopiaSnapshot,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActionSetList is the definition of a list of ActionSets
type ActionSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []*ActionSet `json:"items"`
}

var _ runtime.Object = (*Blueprint)(nil)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Blueprint describes kanister actions.
type Blueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Actions           map[string]*BlueprintAction `json:"actions,omitempty"`
}

// BlueprintAction describes the set of phases that constitute an action.
type BlueprintAction struct {
	Name               string              `json:"name"`
	Kind               string              `json:"kind"`
	ConfigMapNames     []string            `json:"configMapNames,omitempty"`
	SecretNames        []string            `json:"secretNames,omitempty"`
	InputArtifactNames []string            `json:"inputArtifactNames,omitempty"`
	OutputArtifacts    map[string]Artifact `json:"outputArtifacts,omitempty"`
	Phases             []BlueprintPhase    `json:"phases,omitempty"`
	DeferPhase         *BlueprintPhase     `json:"deferPhase,omitempty"`
}

// BlueprintPhase is a an individual unit of execution.
type BlueprintPhase struct {
	Func       string                     `json:"func"`
	Name       string                     `json:"name"`
	ObjectRefs map[string]ObjectReference `json:"objects,omitempty"`
	Args       map[string]interface{}     `json:"args"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BlueprintList is the definition of a list of Blueprints
type BlueprintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []*Blueprint `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Profile
type Profile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Location          Location   `json:"location"`
	Credential        Credential `json:"credential"`
	SkipSSLVerify     bool       `json:"skipSSLVerify"`
}

// LocationType
type LocationType string

const (
	LocationTypeGCS         LocationType = "gcs"
	LocationTypeS3Compliant LocationType = "s3Compliant"
	LocationTypeAzure       LocationType = "azure"
	LocationTypeKopia       LocationType = "kopia"
)

// Location
type Location struct {
	Type     LocationType `json:"type"`
	Bucket   string       `json:"bucket"`
	Endpoint string       `json:"endpoint"`
	Prefix   string       `json:"prefix"`
	Region   string       `json:"region"`
}

// CredentialType
type CredentialType string

const (
	CredentialTypeKeyPair CredentialType = "keyPair"
	CredentialTypeSecret  CredentialType = "secret"
	CredentialTypeKopia   CredentialType = "kopia"
)

// Credential
type Credential struct {
	Type              CredentialType     `json:"type"`
	KeyPair           *KeyPair           `json:"keyPair,omitempty"`
	Secret            *ObjectReference   `json:"secret,omitempty"`
	KopiaServerSecret *KopiaServerSecret `json:"kopiaServerSecret,omitempty"`
}

// KeyPair
type KeyPair struct {
	IDField     string          `json:"idField"`
	SecretField string          `json:"secretField"`
	Secret      ObjectReference `json:"secret"`
}

// KopiaServerSecret contains credentials to connect to Kopia server
type KopiaServerSecret struct {
	Username       string                `json:"username,omitempty"`
	Hostname       string                `json:"hostname,omitempty"`
	UserPassphrase *KopiaServerSecretRef `json:"userPassphrase,omitempty"`
	TLSCert        *KopiaServerSecretRef `json:"tlsCert,omitempty"`
	ConnectOptions map[string]int        `json:"connectOptions,omitempty"`
}

// KopiaServerSecretRef refers to K8s secrets containing Kopia creds
type KopiaServerSecretRef struct {
	Key    string           `json:"key"`
	Secret *ObjectReference `json:"secret"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProfileList is the definition of a list of Profiles
type ProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []*Profile `json:"items"`
}
