/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodetasks

import (
	"k8s.io/kops/pkg/tasks"
	"reflect"
	"testing"
)

func TestServiceTask_Deps(t *testing.T) {
	s := &Service{}

	taskMap := make(map[string]tasks.Task)
	taskMap["LoadImageTask1"] = &LoadImageTask{}
	taskMap["FileTask1"] = &File{}

	deps := s.GetDependencies(taskMap)
	expected := []tasks.Task{taskMap["FileTask1"]}
	if !reflect.DeepEqual(expected, deps) {
		t.Fatalf("unexpected deps.  expected=%v, actual=%v", expected, deps)
	}
}

type FakeTask struct {
}

func (t *FakeTask) Run(tasks.Context) error {
	panic("not implemented")
}
func TestServiceTask_UnknownTypes(t *testing.T) {
	s := &Service{}

	taskMap := make(map[string]tasks.Task)
	taskMap["FakeTask1"] = &FakeTask{}

	deps := s.GetDependencies(taskMap)
	expected := []tasks.Task{taskMap["FakeTask1"]}
	if !reflect.DeepEqual(expected, deps) {
		t.Fatalf("unexpected deps.  expected=%v, actual=%v", expected, deps)
	}
}
