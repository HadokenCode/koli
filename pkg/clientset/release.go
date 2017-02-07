package clientset

import (
	"encoding/json"
	"errors"

	"kolihub.io/koli/pkg/spec"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"
)

// ReleaseGetter has a method to return an ReleaseInterface.
// A group's client should implement this interface.
type ReleaseGetter interface {
	Release(namespace string) ReleaseInterface
}

// ReleaseInterface has methods to work with Release resources.
type ReleaseInterface interface {
	List(opts *api.ListOptions) (*spec.ReleaseList, error)
	Get(name string) (*spec.Release, error)
	Delete(name string, options *api.DeleteOptions) error
	Create(data *spec.Release) (*spec.Release, error)
	Update(data *spec.Release) (*spec.Release, error)
	Watch(opts *api.ListOptions) (watch.Interface, error)
	Patch(name string, pt api.PatchType, data []byte, subresources ...string) (*spec.Release, error)
}

// release implements ReleaseInterface
type release struct {
	client    restclient.Interface
	namespace string
	resource  *unversioned.APIResource
}

// Get gets the resource with the specified name.
func (r *release) Get(name string) (*spec.Release, error) {
	release := &spec.Release{}
	err := r.client.Get().
		NamespaceIfScoped(r.namespace, r.resource.Namespaced).
		Resource(r.resource.Name).
		Name(name).
		Do().
		Into(release)
	return release, err
}

// List returns a list of objects for this resource.
func (r *release) List(opts *api.ListOptions) (*spec.ReleaseList, error) {
	if opts == nil {
		opts = &api.ListOptions{}
	}
	releaseList := &spec.ReleaseList{}
	err := r.client.Get().
		NamespaceIfScoped(r.namespace, r.resource.Namespaced).
		Resource(r.resource.Name).
		FieldsSelectorParam(nil).
		VersionedParams(opts, api.ParameterCodec). // TODO: test this option
		Do().
		Into(releaseList)
	return releaseList, err
}

// Delete deletes the resource with the specified name.
func (r *release) Delete(name string, opts *api.DeleteOptions) error {
	if opts == nil {
		opts = &api.DeleteOptions{}
	}
	return r.client.Delete().
		NamespaceIfScoped(r.namespace, r.resource.Namespaced).
		Resource(r.resource.Name).
		Name(name).
		Body(opts).
		Do().
		Error()
}

// Create creates the provided resource.
func (r *release) Create(data *spec.Release) (*spec.Release, error) {
	release := &spec.Release{}
	err := r.client.Post().
		NamespaceIfScoped(r.namespace, r.resource.Namespaced).
		Resource(r.resource.Name).
		Body(data).
		Do().
		Into(release)
	return release, err
}

// Update updates the provided resource.
func (r *release) Update(data *spec.Release) (*spec.Release, error) {
	release := &spec.Release{}
	if len(data.GetName()) == 0 {
		return data, errors.New("object missing name")
	}
	err := r.client.Put().
		NamespaceIfScoped(r.namespace, r.resource.Namespaced).
		Resource(r.resource.Name).
		Name(data.GetName()).
		Body(data).
		Do().
		Into(release)
	return release, err
}

// Watch returns a watch.Interface that watches the resource.
func (r *release) Watch(opts *api.ListOptions) (watch.Interface, error) {
	// TODO: Using Watch method gives the following error on creation and deletion of resources:
	// expected type X, but watch event object had type *runtime.Unstructured
	stream, err := r.client.Get().
		Prefix("watch").
		NamespaceIfScoped(r.namespace, r.resource.Namespaced).
		Resource(r.resource.Name).
		// VersionedParams(opts, spec.DefaultParameterEncoder).
		VersionedParams(opts, api.ParameterCodec).
		Stream()
	if err != nil {
		return nil, err
	}

	return watch.NewStreamWatcher(&releaseDecoder{
		dec:   json.NewDecoder(stream),
		close: stream.Close,
	}), nil
}

// Patch applies the patch and returns the patched release.
func (r *release) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (*spec.Release, error) {
	release := &spec.Release{}
	err := r.client.Patch(pt).
		Namespace(r.namespace).
		Resource(r.resource.Name).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(release)
	return release, err
}

// releaseDecoder provides a decoder for watching release resources
type releaseDecoder struct {
	dec   *json.Decoder
	close func() error
}

// Close decoder
func (d *releaseDecoder) Close() {
	d.close()
}

// Decode data
func (d *releaseDecoder) Decode() (watch.EventType, runtime.Object, error) {
	var e struct {
		Type   watch.EventType
		Object spec.Release
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
