package client

import (
	"github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type EDPComponentInterface interface {
	Create(*v1alpha1.EDPComponent) (*v1alpha1.EDPComponent, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.EDPComponent, error)
}

type edpComponents struct {
	client rest.Interface
	ns     string
}

func newEdpComponents(c *EDPComponentV1Client, namespace string) *edpComponents {
	return &edpComponents{
		client: c.restClient,
		ns:     namespace,
	}
}

func (e edpComponents) Create(c *v1alpha1.EDPComponent) (res *v1alpha1.EDPComponent, err error) {
	res = &v1alpha1.EDPComponent{}
	err = e.client.Post().
		Namespace(e.ns).
		Resource("edpcomponents").
		Body(c).
		Do().
		Into(res)
	return
}

func (e edpComponents) Get(name string, options metav1.GetOptions) (res *v1alpha1.EDPComponent, err error) {
	res = &v1alpha1.EDPComponent{}
	err = e.client.Get().
		Namespace(e.ns).
		Resource("edpcomponents").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(res)
	return
}
