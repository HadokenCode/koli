package spec

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList v1.ResourceList

// Plan defines how resources could be managed and distributed
type Plan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PlanSpec `json:"spec"`
}

// PlanList is a list of ServicePlans
type PlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Plan `json:"items"`
}

// PlanSpec holds specification parameters of an Plan
type PlanSpec struct {
	// Compute Resources required by containers.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Hard is the set of desired hard limits for each named resource.
	Hard  ResourceList   `json:"hard,omitempty"`
	Roles []PlatformRole `json:"roles,omitempty"`
}

const (
	// ResourceNamespace , number
	ResourceNamespace v1.ResourceName = "namespaces"
)

// PlatformRole is the name identifying various roles in a PlatformRoleList.
type PlatformRole string

const (
	// RoleExecAllow cluster role name
	RoleExecAllow PlatformRole = "exec-allow"
	// RolePortForwardAllow cluster role name
	RolePortForwardAllow PlatformRole = "portforward-allow"
	// RoleAutoScaleAllow cluster role name
	RoleAutoScaleAllow PlatformRole = "autoscale-allow"
	// RoleAttachAllow cluster role name
	RoleAttachAllow PlatformRole = "attach-allow"
	// RoleAddonManagement cluster role name
	RoleAddonManagement PlatformRole = "addon-management"
)

// ServicePlanPhase is the current lifecycle phase of the Service Plan.
type ServicePlanPhase string

const (
	// ServicePlanActive means the ServicePlan is available for use in the system
	ServicePlanActive ServicePlanPhase = "Active"
	// ServicePlanPending means the ServicePlan isn't associate with any global ServicePlan
	ServicePlanPending ServicePlanPhase = "Pending"
	// ServicePlanNotFound means the reference plan wasn't found
	ServicePlanNotFound ServicePlanPhase = "NotFound"
	// ServicePlanDisabled means the ServicePlan is disabled and cannot be associated with resources
	ServicePlanDisabled ServicePlanPhase = "Disabled"
)

// Addon defines integration with external resources
type Addon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AddonSpec `json:"spec"`
}

// AddonList is a list of Addons.
type AddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Addon `json:"items"`
}

// AddonSpec holds specification parameters of an addon
type AddonSpec struct {
	Type      string      `json:"type"`
	BaseImage string      `json:"baseImage"`
	Version   string      `json:"version"`
	Replicas  int32       `json:"replicas"`
	Port      int32       `json:"port"`
	Env       []v1.EnvVar `json:"env"`
	// More info: http://releases.k8s.io/HEAD/docs/user-guide/containers.md#containers-and-commands
	Args []string `json:"args,omitempty"`
}

// Release refers to compiled slug file versions
type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ReleaseSpec `json:"spec"`
}

// SourceType refers to the source of the build
type SourceType string

const (
	// GitHubSource means the build came from a webhook
	GitHubSource SourceType = "github"
	// GitLocalSource means the build came from the git local server
	GitLocalSource SourceType = "local"
)

// ReleaseSpec holds specification parameters of a release
type ReleaseSpec struct {
	// The URL of the git remote server to download the git revision tarball
	GitRemote     string     `json:"gitRemote"`
	GitRevision   string     `json:"gitRevision"`
	GitRepository string     `json:"gitRepository"`
	BuildRevision string     `json:"buildRevision"`
	AutoDeploy    bool       `json:"autoDeploy"`
	ExpireAfter   int32      `json:"expireAfter"`
	DeployName    string     `json:"deployName"`
	Build         bool       `json:"build"`
	AuthToken     string     `json:"authToken"` // expirable token
	Source        SourceType `json:"sourceType"`
}

// ReleaseList is a list of Release
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Release `json:"items"`
}

// User identifies an user on the platform
type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Customer     string `json:"customer"`
	Organization string `json:"org"`
	// Groups are a set of strings which associate users with as set of commonly grouped users.
	// A group name is unique in the cluster and it's formed by it's namespace, customer or the organization name:
	// [org] - Matches all the namespaces of the broker
	// [customer]-[org] - Matches all namespaces from the customer broker
	// [name]-[customer]-[org] - Matches a specific namespace
	// http://kubernetes.io/docs/admin/authentication/
	Groups []string `json:"groups"`
}

// Domain are a way for users to "claim" a domain and be able to create
// ingresses
type Domain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DomainSpec   `json:"spec,omitempty"`
	Status DomainStatus `json:"status,omitempty"`
}

// DomainList is a List of Domain
type DomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Domain `json:"items"`
}

// DomainStatus represents information about the status of a domain.
type DomainStatus struct {
	// The state of the domain, an empty state means it's a new resource
	// +optional
	Phase DomainPhase `json:"phase,omitempty"`
	// A human readable message indicating details about why the domain claim is in this state.
	// +optional
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the domain claim is in this state. e.g. 'AlreadyClaimed'
	// +optional
	Reason string `json:"reason,omitempty"`
	// The last time the resource was updated
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
	// DeletionTimestamp it's a temporary field to work around the issue:
	// https://github.com/kubernetes/kubernetes/issues/40715, once it's solved,
	// remove this field and use the DeletionTimestamp from metav1.ObjectMeta
	DeletionTimestamp *metav1.Time `json:"deletionTimestamp,omitempty"`
}

// DomainSpec represents information about a domain claim
type DomainSpec struct {
	// PrimaryDomain is the name of the primary domain, to set the resource as primary,
	// 'name' and 'primary' must have the same value.
	// +required
	PrimaryDomain string `json:"primary,omitempty"`
	// Sub is the label of the Primary Domain to form a subdomain
	// +optional
	Sub string `json:"sub,omitempty"`
	// Delegates contains a list of namespaces that are allowed to use this domain.
	// New domain resources could be referenced to primary ones using the 'parent' key.
	// A wildcard ("*") allows delegate access to all namespaces in the cluster.
	// +optional
	Delegates []string `json:"delegates,omitempty"`
	// Parent refers to the namespace where the primary domain is in.
	// It only makes sense when the type of the domain is set to 'shared',
	// +optional
	Parent string `json:"parent,omitempty"`
}

// DomainPhase is a label for the condition of a domain at the current time.
type DomainPhase string

const (
	// DomainStatusNew means it's a new resource and the phase it's not set
	DomainStatusNew DomainPhase = ""
	// DomainStatusOK means the domain doesn't have no pending operations or prohibitions,
	// and new ingresses could be created using the target domain.
	DomainStatusOK DomainPhase = "OK"
	// DomainStatusPending indicates that a request to create a new domain
	// has been received and is being processed.
	DomainStatusPending DomainPhase = "Pending"
	// DomainStatusFailed means the resource has failed on claiming the domain
	DomainStatusFailed DomainPhase = "Failed"
)
