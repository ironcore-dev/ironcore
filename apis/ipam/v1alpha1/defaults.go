package v1alpha1

func SetDefaults_PrefixReference(ref *PrefixReference) {
	if ref.Kind == "" {
		ref.Kind = PrefixKind
	}
}

func SetDefaults_PrefixSelector(sel *PrefixSelector) {
	if sel.Kind == "" {
		sel.Kind = PrefixKind
	}
}
