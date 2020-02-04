package envtest

import apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

// mergePaths merges two string slices containing paths.
// This function makes no guarantees about order of the merged slice.
func mergePaths(s1, s2 []string) []string {
	m := make(map[string]struct{})
	for _, s := range s1 {
		m[s] = struct{}{}
	}
	for _, s := range s2 {
		m[s] = struct{}{}
	}
	merged := make([]string, len(m))
	i := 0
	for key := range m {
		merged[i] = key
		i++
	}
	return merged
}

// mergeCRDs merges two CRD slices using their names.
// This function makes no guarantees about order of the merged slice.
func mergeCRDs(s1, s2 []*apiextensionsv1beta1.CustomResourceDefinition) []*apiextensionsv1beta1.CustomResourceDefinition {
	m := make(map[string]*apiextensionsv1beta1.CustomResourceDefinition)
	for _, crd := range s1 {
		m[crd.Name] = crd
	}
	for _, crd := range s2 {
		m[crd.Name] = crd
	}
	merged := make([]*apiextensionsv1beta1.CustomResourceDefinition, len(m))
	i := 0
	for _, crd := range m {
		merged[i] = crd
		i++
	}
	return merged
}
