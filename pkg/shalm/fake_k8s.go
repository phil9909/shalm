// Code generated by counterfeiter. DO NOT EDIT.
package shalm

import (
	"io"
	"sync"
)

type FakeK8s struct {
	ApplyStub        func(func(io.Writer) error, *K8sOptions) error
	applyMutex       sync.RWMutex
	applyArgsForCall []struct {
		arg1 func(io.Writer) error
		arg2 *K8sOptions
	}
	applyReturns struct {
		result1 error
	}
	applyReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteStub        func(func(io.Writer) error, *K8sOptions) error
	deleteMutex       sync.RWMutex
	deleteArgsForCall []struct {
		arg1 func(io.Writer) error
		arg2 *K8sOptions
	}
	deleteReturns struct {
		result1 error
	}
	deleteReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteObjectStub        func(string, string, *K8sOptions) error
	deleteObjectMutex       sync.RWMutex
	deleteObjectArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 *K8sOptions
	}
	deleteObjectReturns struct {
		result1 error
	}
	deleteObjectReturnsOnCall map[int]struct {
		result1 error
	}
	ForNamespaceStub        func(string) K8s
	forNamespaceMutex       sync.RWMutex
	forNamespaceArgsForCall []struct {
		arg1 string
	}
	forNamespaceReturns struct {
		result1 K8s
	}
	forNamespaceReturnsOnCall map[int]struct {
		result1 K8s
	}
	GetStub        func(string, string, io.Writer, *K8sOptions) error
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 io.Writer
		arg4 *K8sOptions
	}
	getReturns struct {
		result1 error
	}
	getReturnsOnCall map[int]struct {
		result1 error
	}
	IsNotExistStub        func(error) bool
	isNotExistMutex       sync.RWMutex
	isNotExistArgsForCall []struct {
		arg1 error
	}
	isNotExistReturns struct {
		result1 bool
	}
	isNotExistReturnsOnCall map[int]struct {
		result1 bool
	}
	RolloutStatusStub        func(string, string, *K8sOptions) error
	rolloutStatusMutex       sync.RWMutex
	rolloutStatusArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 *K8sOptions
	}
	rolloutStatusReturns struct {
		result1 error
	}
	rolloutStatusReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeK8s) Apply(arg1 func(io.Writer) error, arg2 *K8sOptions) error {
	fake.applyMutex.Lock()
	ret, specificReturn := fake.applyReturnsOnCall[len(fake.applyArgsForCall)]
	fake.applyArgsForCall = append(fake.applyArgsForCall, struct {
		arg1 func(io.Writer) error
		arg2 *K8sOptions
	}{arg1, arg2})
	fake.recordInvocation("Apply", []interface{}{arg1, arg2})
	fake.applyMutex.Unlock()
	if fake.ApplyStub != nil {
		return fake.ApplyStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.applyReturns
	return fakeReturns.result1
}

func (fake *FakeK8s) ApplyCallCount() int {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	return len(fake.applyArgsForCall)
}

func (fake *FakeK8s) ApplyCalls(stub func(func(io.Writer) error, *K8sOptions) error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = stub
}

func (fake *FakeK8s) ApplyArgsForCall(i int) (func(io.Writer) error, *K8sOptions) {
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	argsForCall := fake.applyArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeK8s) ApplyReturns(result1 error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = nil
	fake.applyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) ApplyReturnsOnCall(i int, result1 error) {
	fake.applyMutex.Lock()
	defer fake.applyMutex.Unlock()
	fake.ApplyStub = nil
	if fake.applyReturnsOnCall == nil {
		fake.applyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.applyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) Delete(arg1 func(io.Writer) error, arg2 *K8sOptions) error {
	fake.deleteMutex.Lock()
	ret, specificReturn := fake.deleteReturnsOnCall[len(fake.deleteArgsForCall)]
	fake.deleteArgsForCall = append(fake.deleteArgsForCall, struct {
		arg1 func(io.Writer) error
		arg2 *K8sOptions
	}{arg1, arg2})
	fake.recordInvocation("Delete", []interface{}{arg1, arg2})
	fake.deleteMutex.Unlock()
	if fake.DeleteStub != nil {
		return fake.DeleteStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.deleteReturns
	return fakeReturns.result1
}

func (fake *FakeK8s) DeleteCallCount() int {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	return len(fake.deleteArgsForCall)
}

func (fake *FakeK8s) DeleteCalls(stub func(func(io.Writer) error, *K8sOptions) error) {
	fake.deleteMutex.Lock()
	defer fake.deleteMutex.Unlock()
	fake.DeleteStub = stub
}

func (fake *FakeK8s) DeleteArgsForCall(i int) (func(io.Writer) error, *K8sOptions) {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	argsForCall := fake.deleteArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeK8s) DeleteReturns(result1 error) {
	fake.deleteMutex.Lock()
	defer fake.deleteMutex.Unlock()
	fake.DeleteStub = nil
	fake.deleteReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) DeleteReturnsOnCall(i int, result1 error) {
	fake.deleteMutex.Lock()
	defer fake.deleteMutex.Unlock()
	fake.DeleteStub = nil
	if fake.deleteReturnsOnCall == nil {
		fake.deleteReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) DeleteObject(arg1 string, arg2 string, arg3 *K8sOptions) error {
	fake.deleteObjectMutex.Lock()
	ret, specificReturn := fake.deleteObjectReturnsOnCall[len(fake.deleteObjectArgsForCall)]
	fake.deleteObjectArgsForCall = append(fake.deleteObjectArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 *K8sOptions
	}{arg1, arg2, arg3})
	fake.recordInvocation("DeleteObject", []interface{}{arg1, arg2, arg3})
	fake.deleteObjectMutex.Unlock()
	if fake.DeleteObjectStub != nil {
		return fake.DeleteObjectStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.deleteObjectReturns
	return fakeReturns.result1
}

func (fake *FakeK8s) DeleteObjectCallCount() int {
	fake.deleteObjectMutex.RLock()
	defer fake.deleteObjectMutex.RUnlock()
	return len(fake.deleteObjectArgsForCall)
}

func (fake *FakeK8s) DeleteObjectCalls(stub func(string, string, *K8sOptions) error) {
	fake.deleteObjectMutex.Lock()
	defer fake.deleteObjectMutex.Unlock()
	fake.DeleteObjectStub = stub
}

func (fake *FakeK8s) DeleteObjectArgsForCall(i int) (string, string, *K8sOptions) {
	fake.deleteObjectMutex.RLock()
	defer fake.deleteObjectMutex.RUnlock()
	argsForCall := fake.deleteObjectArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeK8s) DeleteObjectReturns(result1 error) {
	fake.deleteObjectMutex.Lock()
	defer fake.deleteObjectMutex.Unlock()
	fake.DeleteObjectStub = nil
	fake.deleteObjectReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) DeleteObjectReturnsOnCall(i int, result1 error) {
	fake.deleteObjectMutex.Lock()
	defer fake.deleteObjectMutex.Unlock()
	fake.DeleteObjectStub = nil
	if fake.deleteObjectReturnsOnCall == nil {
		fake.deleteObjectReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteObjectReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) ForNamespace(arg1 string) K8s {
	fake.forNamespaceMutex.Lock()
	ret, specificReturn := fake.forNamespaceReturnsOnCall[len(fake.forNamespaceArgsForCall)]
	fake.forNamespaceArgsForCall = append(fake.forNamespaceArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("ForNamespace", []interface{}{arg1})
	fake.forNamespaceMutex.Unlock()
	if fake.ForNamespaceStub != nil {
		return fake.ForNamespaceStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.forNamespaceReturns
	return fakeReturns.result1
}

func (fake *FakeK8s) ForNamespaceCallCount() int {
	fake.forNamespaceMutex.RLock()
	defer fake.forNamespaceMutex.RUnlock()
	return len(fake.forNamespaceArgsForCall)
}

func (fake *FakeK8s) ForNamespaceCalls(stub func(string) K8s) {
	fake.forNamespaceMutex.Lock()
	defer fake.forNamespaceMutex.Unlock()
	fake.ForNamespaceStub = stub
}

func (fake *FakeK8s) ForNamespaceArgsForCall(i int) string {
	fake.forNamespaceMutex.RLock()
	defer fake.forNamespaceMutex.RUnlock()
	argsForCall := fake.forNamespaceArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeK8s) ForNamespaceReturns(result1 K8s) {
	fake.forNamespaceMutex.Lock()
	defer fake.forNamespaceMutex.Unlock()
	fake.ForNamespaceStub = nil
	fake.forNamespaceReturns = struct {
		result1 K8s
	}{result1}
}

func (fake *FakeK8s) ForNamespaceReturnsOnCall(i int, result1 K8s) {
	fake.forNamespaceMutex.Lock()
	defer fake.forNamespaceMutex.Unlock()
	fake.ForNamespaceStub = nil
	if fake.forNamespaceReturnsOnCall == nil {
		fake.forNamespaceReturnsOnCall = make(map[int]struct {
			result1 K8s
		})
	}
	fake.forNamespaceReturnsOnCall[i] = struct {
		result1 K8s
	}{result1}
}

func (fake *FakeK8s) Get(arg1 string, arg2 string, arg3 io.Writer, arg4 *K8sOptions) error {
	fake.getMutex.Lock()
	ret, specificReturn := fake.getReturnsOnCall[len(fake.getArgsForCall)]
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 io.Writer
		arg4 *K8sOptions
	}{arg1, arg2, arg3, arg4})
	fake.recordInvocation("Get", []interface{}{arg1, arg2, arg3, arg4})
	fake.getMutex.Unlock()
	if fake.GetStub != nil {
		return fake.GetStub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getReturns
	return fakeReturns.result1
}

func (fake *FakeK8s) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *FakeK8s) GetCalls(stub func(string, string, io.Writer, *K8sOptions) error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = stub
}

func (fake *FakeK8s) GetArgsForCall(i int) (string, string, io.Writer, *K8sOptions) {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	argsForCall := fake.getArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeK8s) GetReturns(result1 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) GetReturnsOnCall(i int, result1 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	if fake.getReturnsOnCall == nil {
		fake.getReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.getReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) IsNotExist(arg1 error) bool {
	fake.isNotExistMutex.Lock()
	ret, specificReturn := fake.isNotExistReturnsOnCall[len(fake.isNotExistArgsForCall)]
	fake.isNotExistArgsForCall = append(fake.isNotExistArgsForCall, struct {
		arg1 error
	}{arg1})
	fake.recordInvocation("IsNotExist", []interface{}{arg1})
	fake.isNotExistMutex.Unlock()
	if fake.IsNotExistStub != nil {
		return fake.IsNotExistStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.isNotExistReturns
	return fakeReturns.result1
}

func (fake *FakeK8s) IsNotExistCallCount() int {
	fake.isNotExistMutex.RLock()
	defer fake.isNotExistMutex.RUnlock()
	return len(fake.isNotExistArgsForCall)
}

func (fake *FakeK8s) IsNotExistCalls(stub func(error) bool) {
	fake.isNotExistMutex.Lock()
	defer fake.isNotExistMutex.Unlock()
	fake.IsNotExistStub = stub
}

func (fake *FakeK8s) IsNotExistArgsForCall(i int) error {
	fake.isNotExistMutex.RLock()
	defer fake.isNotExistMutex.RUnlock()
	argsForCall := fake.isNotExistArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeK8s) IsNotExistReturns(result1 bool) {
	fake.isNotExistMutex.Lock()
	defer fake.isNotExistMutex.Unlock()
	fake.IsNotExistStub = nil
	fake.isNotExistReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeK8s) IsNotExistReturnsOnCall(i int, result1 bool) {
	fake.isNotExistMutex.Lock()
	defer fake.isNotExistMutex.Unlock()
	fake.IsNotExistStub = nil
	if fake.isNotExistReturnsOnCall == nil {
		fake.isNotExistReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.isNotExistReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeK8s) RolloutStatus(arg1 string, arg2 string, arg3 *K8sOptions) error {
	fake.rolloutStatusMutex.Lock()
	ret, specificReturn := fake.rolloutStatusReturnsOnCall[len(fake.rolloutStatusArgsForCall)]
	fake.rolloutStatusArgsForCall = append(fake.rolloutStatusArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 *K8sOptions
	}{arg1, arg2, arg3})
	fake.recordInvocation("RolloutStatus", []interface{}{arg1, arg2, arg3})
	fake.rolloutStatusMutex.Unlock()
	if fake.RolloutStatusStub != nil {
		return fake.RolloutStatusStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.rolloutStatusReturns
	return fakeReturns.result1
}

func (fake *FakeK8s) RolloutStatusCallCount() int {
	fake.rolloutStatusMutex.RLock()
	defer fake.rolloutStatusMutex.RUnlock()
	return len(fake.rolloutStatusArgsForCall)
}

func (fake *FakeK8s) RolloutStatusCalls(stub func(string, string, *K8sOptions) error) {
	fake.rolloutStatusMutex.Lock()
	defer fake.rolloutStatusMutex.Unlock()
	fake.RolloutStatusStub = stub
}

func (fake *FakeK8s) RolloutStatusArgsForCall(i int) (string, string, *K8sOptions) {
	fake.rolloutStatusMutex.RLock()
	defer fake.rolloutStatusMutex.RUnlock()
	argsForCall := fake.rolloutStatusArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeK8s) RolloutStatusReturns(result1 error) {
	fake.rolloutStatusMutex.Lock()
	defer fake.rolloutStatusMutex.Unlock()
	fake.RolloutStatusStub = nil
	fake.rolloutStatusReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) RolloutStatusReturnsOnCall(i int, result1 error) {
	fake.rolloutStatusMutex.Lock()
	defer fake.rolloutStatusMutex.Unlock()
	fake.RolloutStatusStub = nil
	if fake.rolloutStatusReturnsOnCall == nil {
		fake.rolloutStatusReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.rolloutStatusReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeK8s) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.applyMutex.RLock()
	defer fake.applyMutex.RUnlock()
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	fake.deleteObjectMutex.RLock()
	defer fake.deleteObjectMutex.RUnlock()
	fake.forNamespaceMutex.RLock()
	defer fake.forNamespaceMutex.RUnlock()
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	fake.isNotExistMutex.RLock()
	defer fake.isNotExistMutex.RUnlock()
	fake.rolloutStatusMutex.RLock()
	defer fake.rolloutStatusMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeK8s) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ K8s = new(FakeK8s)