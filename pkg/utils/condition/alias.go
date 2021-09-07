package condition

var (
	// DefaultAccessor is an Accessor initialized with the default fields.
	// See NewAccessor for more.
	DefaultAccessor = NewAccessor(AccessorOptions{})

	// Update updates the condition with the given options.
	// See Accessor.Update for more.
	Update = DefaultAccessor.Update

	// MustUpdate updates the condition with the given options.
	// See Accessor.MustUpdate for more.
	MustUpdate = DefaultAccessor.MustUpdate

	// UpdateSlice updates the slice with the given options.
	// See Accessor.UpdateSlice for more.
	UpdateSlice = DefaultAccessor.UpdateSlice

	// MustUpdateSlice updates the slice with the given options.
	// See Accessor.MustUpdateSlice for more.
	MustUpdateSlice = DefaultAccessor.MustUpdateSlice

	// FindSlice finds the target condition in the given slice.
	// See Accessor.FindSlice for more.
	FindSlice = DefaultAccessor.FindSlice

	// MustFindSlice finds the target condition in the given slice.
	// See Accessor.MustFindSlice for more.
	MustFindSlice = DefaultAccessor.MustFindSlice

	// FindSliceStatus finds the condition status in the given slice.
	// See Accessor.FindSliceStatus for more.
	FindSliceStatus = DefaultAccessor.FindSliceStatus

	// MustFindSliceStatus finds the condition status in the given slice.
	// See Accessor.MustFindSliceStatus for more.
	MustFindSliceStatus = DefaultAccessor.MustFindSliceStatus
)
