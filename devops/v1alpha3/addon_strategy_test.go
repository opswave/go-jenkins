/*
Copyright 2022 The KubeSphere Authors.

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

package v1alpha3

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddonInstallStrategy_IsValid(t *testing.T) {
	tests := []struct {
		name string
		a    AddonInstallStrategy
		want bool
	}{{
		name: "normal case - simple",
		a:    AddonInstallStrategySimple,
		want: true,
	}, {
		name: "normal case - helm",
		a:    AddonInstallStrategyHelm,
		want: true,
	}, {
		name: "normal case - operator",
		a:    AddonInstallStrategyOperator,
		want: true,
	}, {
		name: "normal case - simple-operator",
		a:    AddonInstallStrategySimpleOperator,
		want: true,
	}, {
		name: "a fake strategy",
		a:    AddonInstallStrategy("fake"),
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.a.IsValid(), "IsValid()")
		})
	}
}
