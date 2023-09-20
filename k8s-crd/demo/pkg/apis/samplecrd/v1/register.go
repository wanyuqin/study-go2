package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"study-go/k8s-crd/pkg/apis/samplecrd"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   samplecrd.GroupName,
	Version: samplecrd.Version,
}

func addKnownTypes(scheme *runtime.Scheme) error {
	//scheme.AddKnownTypes(
	//	SchemeGroupVersion,
	//	&Network{},
	//	&NetworkList{})
	//
	//metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
