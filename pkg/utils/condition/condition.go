// Package condition simplifies condition handling with any structurally compatible condition
// (comparable to a sort of duck-typing) via go reflection.
package condition

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/util/clock"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"reflect"
)

const (
	// DefaultTypeField is the default name for a condition's type field.
	DefaultTypeField = "Type"
	// DefaultStatusField is the default name for a condition's status field.
	DefaultStatusField = "Status"
	// DefaultLastUpdateTimeField field is the default name for a condition's last update time field.
	DefaultLastUpdateTimeField = "LastUpdateTime"
	// DefaultLastTransitionTimeField field is the default name for a condition's last transition time field.
	DefaultLastTransitionTimeField = "LastTransitionTime"
	// DefaultReasonField field is the default name for a condition's reason field.
	DefaultReasonField = "Reason"
	// DefaultMessageField field is the default name for a condition's message field.
	DefaultMessageField = "Message"
	// DefaultObservedGenerationField field is the default name for a condition's observed generation field.
	DefaultObservedGenerationField = "ObservedGeneration"
)

func enforceStruct(cond interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(cond)
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("type %T is not a struct", cond)
	}
	return v, nil
}

func enforcePtrToStruct(cond interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(cond)
	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("type %T is not a pointer to a struct", cond)
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("type %T is not a pointer to a struct", cond)
	}
	return v, nil
}

func enforceStructSlice(condSlice interface{}) (sliceV reflect.Value, structType reflect.Type, err error) {
	v := reflect.ValueOf(condSlice)
	if v.Kind() != reflect.Slice {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a slice of structs", condSlice)
	}

	structType = v.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a slice of structs", condSlice)
	}

	return v, structType, nil
}

func enforcePtrToStructSlice(condSlicePtr interface{}) (sliceV reflect.Value, structType reflect.Type, err error) {
	v := reflect.ValueOf(condSlicePtr)
	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a pointer to a slice of structs", condSlicePtr)
	}

	v = v.Elem()

	if v.Kind() != reflect.Slice {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a pointer to a slice of structs", condSlicePtr)
	}

	structType = v.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a pointer to a slice of structs", condSlicePtr)
	}

	return v, structType, nil
}

func getAndConvertField(v reflect.Value, name string, into interface{}) error {
	f := v.FieldByName(name)
	if !v.IsValid() {
		return fmt.Errorf("type %T has no field %q", v.Interface(), name)
	}

	intoV, err := conversion.EnforcePtr(into)
	if err != nil {
		return err
	}

	fType := f.Type()
	if fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}

	if !fType.ConvertibleTo(intoV.Type()) {
		return fmt.Errorf("type %T field %q type %s cannot be converted into %T", v.Interface(), fType, name, into)
	}
	intoV.Set(reflect.Indirect(f).Convert(intoV.Type()))
	return nil
}

// direct is the inverse to reflect.Indirect.
//
// It takes in a value and returns nil if it is a zero value.
// Otherwise, it returns a pointer to that value.
func direct(v reflect.Value) reflect.Value {
	if v.IsZero() {
		return reflect.New(reflect.PtrTo(v.Type())).Elem()
	}

	res := reflect.New(v.Type())
	res.Elem().Set(v)
	return res
}

// setFieldConverted sets the specified field to the given value, potentially converting it before.
func setFieldConverted(v reflect.Value, name string, newValue interface{}) error {
	f := v.FieldByName(name)
	if f == (reflect.Value{}) {
		return fmt.Errorf("type %T has no field %q", v.Interface(), name)
	}

	fType := f.Type()
	var isPtr bool
	if fType.Kind() == reflect.Ptr {
		isPtr = true
		fType = fType.Elem()
	}

	newV := reflect.ValueOf(newValue)
	if !newV.CanConvert(fType) {
		return fmt.Errorf("value %T cannot be converted into type %s of field %q of type %T", newValue, fType, name, v.Interface())
	}

	newV = newV.Convert(fType)
	if isPtr {
		newV = direct(newV)
	}

	f.Set(newV)
	return nil
}

func valueHasField(v reflect.Value, name string) bool {
	return v.FieldByName(name) != (reflect.Value{})
}

type Accessor struct {
	typeField               string
	statusField             string
	lastUpdateTimeField     string
	lastTransitionTimeField string
	reasonField             string
	messageField            string
	observedGenerationField string
	clock                   clock.Clock
}

// Type extracts the type of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) Type(cond interface{}) (string, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return "", err
	}

	var typeValue string
	if err := getAndConvertField(v, a.typeField, &typeValue); err != nil {
		return "", err
	}
	return typeValue, nil
}

// MustType extracts the type of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustType(cond interface{}) string {
	typ, err := a.Type(cond)
	utilruntime.Must(err)
	return typ
}

// SetType sets the type of the given condition to the given value.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetType(condPtr interface{}, typ string) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.typeField, typ)
}

// MustSetType sets the type of the given condition to the given value.
//
// It panics if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetType(condPtr interface{}, typ string) {
	utilruntime.Must(a.SetType(condPtr, typ))
}

// Status extracts the status of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) Status(cond interface{}) (corev1.ConditionStatus, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return "", err
	}

	var status corev1.ConditionStatus
	if err := getAndConvertField(v, a.statusField, &status); err != nil {
		return "", err
	}
	return status, nil
}

// MustStatus extracts the status of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustStatus(cond interface{}) corev1.ConditionStatus {
	status, err := a.Status(cond)
	utilruntime.Must(err)
	return status
}

// SetStatus sets the status of the given condition.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetStatus(condPtr interface{}, status corev1.ConditionStatus) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.statusField, status)
}

// MustSetStatus sets the status of the given condition.
//
// It panics if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetStatus(condPtr interface{}, status corev1.ConditionStatus) {
	utilruntime.Must(a.SetStatus(condPtr, status))
}

// HasLastUpdateTime checks if the given condition has a 'LastUpdateTime' field.
//
// It errors if the given value is not a struct.
func (a *Accessor) HasLastUpdateTime(cond interface{}) (bool, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return false, err
	}

	return valueHasField(v, a.lastUpdateTimeField), nil
}

// MustHasLastUpdateTime checks if the given condition has a 'LastUpdateTime' field.
//
// It panics if the given value is not a struct.
func (a *Accessor) MustHasLastUpdateTime(cond interface{}) bool {
	ok, err := a.HasLastUpdateTime(cond)
	utilruntime.Must(err)
	return ok
}

// LastUpdateTime extracts the last update time of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) LastUpdateTime(cond interface{}) (metav1.Time, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return metav1.Time{}, err
	}

	var lastUpdateTime metav1.Time
	if err := getAndConvertField(v, a.lastUpdateTimeField, &lastUpdateTime); err != nil {
		return metav1.Time{}, err
	}
	return lastUpdateTime, nil
}

// MustLastUpdateTime extracts the last update time of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustLastUpdateTime(cond interface{}) metav1.Time {
	t, err := a.LastUpdateTime(cond)
	utilruntime.Must(err)
	return t
}

// SetLastUpdateTime sets the last update time of the given condition.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetLastUpdateTime(condPtr interface{}, lastUpdateTime metav1.Time) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.lastUpdateTimeField, lastUpdateTime)
}

// MustSetLastUpdateTime sets the last update time of the given condition.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetLastUpdateTime(condPtr interface{}, lastUpdateTime metav1.Time) {
	utilruntime.Must(a.SetLastUpdateTime(condPtr, lastUpdateTime))
}

// SetLastUpdateTimeIfExists sets the last update time of the given condition if the field exists.
//
// It errors if the given value is not a pointer to a struct or the field value cannot be converted
// to the given format.
func (a *Accessor) SetLastUpdateTimeIfExists(condPtr interface{}, lastUpdateTime metav1.Time) error {
	condV, err := conversion.EnforcePtr(condPtr)
	if err != nil {
		return err
	}

	cond := condV.Interface()
	if ok, err := a.HasLastUpdateTime(cond); err != nil || !ok {
		return err
	}

	return a.SetLastUpdateTime(condPtr, lastUpdateTime)
}

// MustSetLastUpdateTimeIfExists sets the last update time of the given condition if the field exists.
//
// It panics if the given value is not a pointer to a struct or the field value cannot be converted
// to the given format.
func (a *Accessor) MustSetLastUpdateTimeIfExists(condPtr interface{}, lastUpdateTime metav1.Time) {
	utilruntime.Must(a.SetLastUpdateTimeIfExists(condPtr, lastUpdateTime))
}

// HasLastTransitionTime checks if the given condition has a 'LastTransitionTime' field.
//
// It errors if the given value is not a struct.
func (a *Accessor) HasLastTransitionTime(cond interface{}) (bool, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return false, err
	}

	return valueHasField(v, a.lastTransitionTimeField), nil
}

// MustHasLastTransitionTime checks if the given condition has a 'LastTransitionTime' field.
//
// It panics if the given value is not a struct.
func (a *Accessor) MustHasLastTransitionTime(cond interface{}) bool {
	ok, err := a.HasLastTransitionTime(cond)
	utilruntime.Must(err)
	return ok
}

// LastTransitionTime extracts the last transition time of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) LastTransitionTime(cond interface{}) (metav1.Time, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return metav1.Time{}, err
	}

	var lastTransitionTime metav1.Time
	if err := getAndConvertField(v, a.lastTransitionTimeField, &lastTransitionTime); err != nil {
		return metav1.Time{}, err
	}
	return lastTransitionTime, nil
}

// MustLastTransitionTime extracts the last transition time of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustLastTransitionTime(cond interface{}) metav1.Time {
	t, err := a.LastTransitionTime(cond)
	utilruntime.Must(err)
	return t
}

// SetLastTransitionTime sets the last transition time of the given condition if the field exists.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) SetLastTransitionTime(condPtr interface{}, lastTransitionTime metav1.Time) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.lastTransitionTimeField, lastTransitionTime)
}

// MustSetLastTransitionTime sets the last transition time of the given condition.
//
// It panics if the given value is not a pointer to a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustSetLastTransitionTime(condPtr interface{}, lastTransitionTime metav1.Time) {
	utilruntime.Must(a.SetLastTransitionTime(condPtr, lastTransitionTime))
}

// SetLastTransitionTimeIfExists sets the last transition time of the given condition.
//
// It errors if the given value is not a pointer to a struct or the field value cannot be converted
// to the given format.
func (a *Accessor) SetLastTransitionTimeIfExists(condPtr interface{}, lastTransitionTime metav1.Time) error {
	condV, err := conversion.EnforcePtr(condPtr)
	if err != nil {
		return err
	}

	cond := condV.Interface()
	if ok, err := a.HasLastTransitionTime(cond); err != nil || !ok {
		return err
	}

	return a.SetLastTransitionTime(condPtr, lastTransitionTime)
}

// MustSetLastTransitionTimeIfExists sets the last transition time of the given condition.
//
// It panics if the given value is not a pointer to a struct or the field value cannot be converted
// to the given format.
func (a *Accessor) MustSetLastTransitionTimeIfExists(condPtr interface{}, lastTransitionTime metav1.Time) {
	utilruntime.Must(a.SetLastTransitionTimeIfExists(condPtr, lastTransitionTime))
}

// Reason extracts the reason of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) Reason(cond interface{}) (string, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return "", err
	}

	var reason string
	if err := getAndConvertField(v, a.reasonField, &reason); err != nil {
		return "", err
	}
	return reason, nil
}

// MustReason extracts the reason of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustReason(cond interface{}) string {
	s, err := a.Reason(cond)
	utilruntime.Must(err)
	return s
}

// SetReason sets the reason of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetReason(condPtr interface{}, reason string) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.reasonField, reason)
}

// MustSetReason sets the reason of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetReason(condPtr interface{}, reason string) {
	utilruntime.Must(a.SetReason(condPtr, reason))
}

// Message gets the message of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the input format.
func (a *Accessor) Message(cond interface{}) (string, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return "", err
	}

	var message string
	if err := getAndConvertField(v, a.messageField, &message); err != nil {
		return "", err
	}
	return message, nil
}

// MustMessage gets the message of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the input format.
func (a *Accessor) MustMessage(cond interface{}) string {
	s, err := a.Message(cond)
	utilruntime.Must(err)
	return s
}

// SetMessage sets the message of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetMessage(condPtr interface{}, message string) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.messageField, message)
}

// MustSetMessage sets the message of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetMessage(condPtr interface{}, message string) {
	utilruntime.Must(a.SetMessage(condPtr, message))
}

// HasObservedGeneration checks if the given condition has a observed generation field.
//
// It errors if the given value is not a struct.
func (a *Accessor) HasObservedGeneration(cond interface{}) (bool, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return false, err
	}

	return valueHasField(v, a.observedGenerationField), nil
}

// MustHasObservedGeneration checks if the given condition has a observed generation field.
//
// It panics if the given value is not a struct.
func (a *Accessor) MustHasObservedGeneration(cond interface{}) bool {
	ok, err := a.HasObservedGeneration(cond)
	utilruntime.Must(err)
	return ok
}

// ObservedGeneration gets the observed generation of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the input format.
func (a *Accessor) ObservedGeneration(cond interface{}) (int64, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return 0, err
	}

	var gen int64
	if err := getAndConvertField(v, a.observedGenerationField, &gen); err != nil {
		return 0, err
	}

	return gen, nil
}

// MustObservedGeneration gets the observed generation of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the input format.
func (a *Accessor) MustObservedGeneration(cond interface{}) int64 {
	gen, err := a.ObservedGeneration(cond)
	utilruntime.Must(err)
	return gen
}

// SetObservedGeneration sets the observed generation of the given condition.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetObservedGeneration(condPtr interface{}, gen int64) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.observedGenerationField, gen)
}

// MustSetObservedGeneration sets the observed generation of the given condition.
//
// It panics if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetObservedGeneration(condPtr interface{}, gen int64) {
	utilruntime.Must(a.SetObservedGeneration(condPtr, gen))
}

// MustSetMessage sets the message of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) findTypeIndex(condSliceV reflect.Value, typ string) (int, error) {
	for i, n := 0, condSliceV.Len(); i < n; i++ {
		it := condSliceV.Index(i)
		itType, err := a.Type(it.Interface())
		if err != nil {
			return -1, fmt.Errorf("[index %d]: %w", i, err)
		}

		if itType == typ {
			return i, nil
		}
	}
	return -1, nil
}

// FindSlice finds the condition with the given type from the given slice and updates the target value with it.
//
// If the target type is not found, false is returned and the target value is not updated.
// FindSlice errors if condSlice is not a slice, intoPtr is not a pointer to a struct and if intoPtr's target
// value is not settable with an element of condSlice.
func (a *Accessor) FindSlice(condSlice interface{}, typ string, intoPtr interface{}) (ok bool, err error) {
	v, elemType, err := enforceStructSlice(condSlice)
	if err != nil {
		return false, err
	}

	intoV, err := enforcePtrToStruct(intoPtr)
	if err != nil {
		return false, err
	}

	if intoV.Type() != elemType {
		return false, fmt.Errorf("into type %T cannot accept slice type %T", intoPtr, condSlice)
	}

	idx, err := a.findTypeIndex(v, typ)
	if err != nil {
		return false, err
	}

	if idx == -1 {
		return false, nil
	}

	intoV.Set(v.Index(idx))
	return true, nil
}

// MustFindSlice finds the condition with the given type from the given slice and updates the target value with it.
//
// If the target type is not found, false is returned and the target value is not updated.
// FindSlice panics if condSlice is not a slice, intoPtr is not a pointer to a struct and if intoPtr's target
// value is not settable with an element of condSlice.
func (a *Accessor) MustFindSlice(condSlice interface{}, typ string, intoPtr interface{}) bool {
	ok, err := a.FindSlice(condSlice, typ, intoPtr)
	utilruntime.Must(err)
	return ok
}

// FindSliceStatus finds the status of the condition with the given type.
// If the condition cannot be found, corev1.ConditionUnknown is returned.
//
// FindSliceStatus errors if the given condSlice is not a slice of structs or if any
// of the conditions does not support access.
func (a *Accessor) FindSliceStatus(condSlice interface{}, typ string) (corev1.ConditionStatus, error) {
	v, _, err := enforceStructSlice(condSlice)
	if err != nil {
		return "", err
	}

	idx, err := a.findTypeIndex(v, typ)
	if err != nil {
		return "", err
	}

	if idx == -1 {
		return corev1.ConditionUnknown, nil
	}

	condV := v.Index(idx)
	return a.Status(condV.Interface())
}

// MustFindSliceStatus finds the status of the condition with the given type.
// If the condition cannot be found, corev1.ConditionUnknown is returned.
//
// MustFindSliceStatus errors if the given condSlice is not a slice of structs or if any
// of the conditions does not support access.
func (a *Accessor) MustFindSliceStatus(condSlice interface{}, typ string) corev1.ConditionStatus {
	status, err := a.FindSliceStatus(condSlice, typ)
	utilruntime.Must(err)
	return status
}

// UpdateOption is an option given to Accessor.UpdateSlice that updates an individual condition.
type UpdateOption interface {
	// ApplyUpdate applies the update, given a condition pointer.
	ApplyUpdate(a *Accessor, condPtr interface{}) error
}

// Update updates the condition with the given options, setting transition- and update time accordingly.
//
// Update errors if the given condPtr is not a pointer to a struct supporting the required condition fields.
func (a *Accessor) Update(condPtr interface{}, opts ...UpdateOption) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	// Record status before updates to be able to infer LastTransitionTime.
	statusBefore, err := a.Status(v.Interface())
	if err != nil {
		return err
	}

	for _, opt := range opts {
		if err := opt.ApplyUpdate(a, condPtr); err != nil {
			return err
		}
	}

	statusAfter, err := a.Status(v.Interface())
	if err != nil {
		return err
	}

	now := metav1.NewTime(a.clock.Now())
	if statusBefore != statusAfter {
		if err := a.SetLastTransitionTimeIfExists(condPtr, now); err != nil {
			return err
		}
	}

	if err := a.SetLastUpdateTimeIfExists(condPtr, now); err != nil {
		return err
	}

	return nil
}

// MustUpdate updates the condition with the given options, setting transition- and update time accordingly.
//
// MustUpdate panics if the given condPtr is not a pointer to a struct supporting the required condition fields.
func (a *Accessor) MustUpdate(condPtr interface{}, opts ...UpdateOption) {
	utilruntime.Must(a.Update(condPtr, opts...))
}

// UpdateSlice finds and updates the condition with the given target type.
//
// UpdateSlice errors if condSlicePtr is not a pointer to a slice of structs that can be accessed with
// this Accessor.
// If no condition with the given type can be found, a new one is appended with the given type and updates
// applied.
// The last update time and last transition time of the new / existing condition is correctly updated:
// For new conditions, it's always set to the current time while for existing conditions, it's checked
// whether the status changed and then updated.
func (a *Accessor) UpdateSlice(condSlicePtr interface{}, typ string, opts ...UpdateOption) error {
	sliceV, elemType, err := enforcePtrToStructSlice(condSlicePtr)
	if err != nil {
		return err
	}

	idx, err := a.findTypeIndex(sliceV, typ)
	if err != nil {
		return err
	}

	var v reflect.Value
	if idx != -1 {
		v = sliceV.Index(idx).Addr()
	} else {
		v = reflect.New(elemType)
	}

	condPtr := v.Interface()

	// Ensure both type and initial transition time (if applicable) are set
	// for new conditions.
	if idx == -1 {
		if err := a.SetType(condPtr, typ); err != nil {
			return err
		}

		now := metav1.NewTime(a.clock.Now())
		if err := a.SetLastTransitionTimeIfExists(condPtr, now); err != nil {
			return err
		}
	}

	if err := a.Update(condPtr, opts...); err != nil {
		return err
	}

	// Make sure to append to the slice in case the condition is new, otherwise
	// it was already updated in-place.
	if idx == -1 {
		sliceV.Set(reflect.Append(sliceV, v.Elem()))
	}
	return nil
}

// MustUpdateSlice finds and updates the condition with the given target type.
//
// MustUpdateSlice panics if condSlicePtr is not a pointer to a slice of structs that can be accessed with
// this Accessor.
// If no condition with the given type can be found, a new one is appended with the given type and updates
// applied.
// The last update time and last transition time of the new / existing condition is correctly updated:
// For new conditions, it's always set to the current time while for existing conditions, it's checked
// whether the status changed and then updated.
func (a *Accessor) MustUpdateSlice(condSlicePtr interface{}, typ string, opts ...UpdateOption) {
	utilruntime.Must(a.UpdateSlice(condSlicePtr, typ, opts...))
}

// UpdateStatus implements UpdateOption to set a corev1.ConditionStatus.
type UpdateStatus corev1.ConditionStatus

// ApplyUpdate implements UpdateOption.
func (u UpdateStatus) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	return a.SetStatus(condPtr, corev1.ConditionStatus(u))
}

// UpdateMessage implements UpdateOption to set the message.
type UpdateMessage string

// ApplyUpdate implements UpdateOption.
func (u UpdateMessage) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	return a.SetMessage(condPtr, string(u))
}

// UpdateReason implements UpdateOption to set the reason.
type UpdateReason string

// ApplyUpdate implements UpdateOption.
func (u UpdateReason) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	return a.SetReason(condPtr, string(u))
}

// UpdateObservedGeneration implements UpdateOption to set the observed generation.
type UpdateObservedGeneration int64

// ApplyUpdate implements UpdateOption.
func (u UpdateObservedGeneration) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	return a.SetObservedGeneration(condPtr, int64(u))
}

// UpdateObserved is a shorthand for updating the observed generation from a metav1.Object's generation.
func UpdateObserved(obj metav1.Object) UpdateObservedGeneration {
	return UpdateObservedGeneration(obj.GetGeneration())
}

// AccessorOptions are options to create an Accessor.
//
// If left blank, defaults are being used via AccessorOptions.SetDefaults.
type AccessorOptions struct {
	TypeField               string
	StatusField             string
	LastUpdateTimeField     string
	LastTransitionTimeField string
	ReasonField             string
	MessageField            string
	ObservedGenerationField string

	Clock clock.Clock
}

// SetDefaults sets default values for AccessorOptions.
func (o *AccessorOptions) SetDefaults() {
	if o.TypeField == "" {
		o.TypeField = DefaultTypeField
	}
	if o.StatusField == "" {
		o.StatusField = DefaultStatusField
	}
	if o.LastUpdateTimeField == "" {
		o.LastUpdateTimeField = DefaultLastUpdateTimeField
	}
	if o.LastTransitionTimeField == "" {
		o.LastTransitionTimeField = DefaultLastTransitionTimeField
	}
	if o.ReasonField == "" {
		o.ReasonField = DefaultReasonField
	}
	if o.MessageField == "" {
		o.MessageField = DefaultMessageField
	}
	if o.ObservedGenerationField == "" {
		o.ObservedGenerationField = DefaultObservedGenerationField
	}
	if o.Clock == nil {
		o.Clock = clock.RealClock{}
	}
}

// NewAccessor creates a new Accessor with the given AccessorOptions.
func NewAccessor(opts AccessorOptions) *Accessor {
	opts.SetDefaults()
	return &Accessor{
		typeField:               opts.TypeField,
		statusField:             opts.StatusField,
		lastUpdateTimeField:     opts.LastUpdateTimeField,
		lastTransitionTimeField: opts.LastTransitionTimeField,
		reasonField:             opts.ReasonField,
		messageField:            opts.MessageField,
		observedGenerationField: opts.ObservedGenerationField,
		clock:                   opts.Clock,
	}
}
