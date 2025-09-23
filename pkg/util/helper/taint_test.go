/*
Copyright 2022 The Karmada Authors.

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

package helper

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1alpha1 "github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"
	policyv1alpha1 "github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
)

var (
	unreachableTaintTemplate = &corev1.Taint{
		Key:    clusterv1alpha1.TaintClusterUnreachable,
		Effect: corev1.TaintEffectNoExecute,
	}

	notReadyTaintTemplate = &corev1.Taint{
		Key:    clusterv1alpha1.TaintClusterNotReady,
		Effect: corev1.TaintEffectNoExecute,
	}
)

func TestSetCurrentClusterTaints(t *testing.T) {
	type args struct {
		taints         []corev1.Taint
		taintsToAdd    []*corev1.Taint
		taintsToRemove []*corev1.Taint
	}
	tests := []struct {
		name       string
		args       args
		wantTaints []corev1.Taint
	}{
		{
			name: "ready condition from true to false",
			args: args{
				taints:         nil,
				taintsToAdd:    []*corev1.Taint{notReadyTaintTemplate.DeepCopy()},
				taintsToRemove: []*corev1.Taint{unreachableTaintTemplate.DeepCopy()},
			},
			wantTaints: []corev1.Taint{*notReadyTaintTemplate},
		},
		{
			name: "ready condition from true to unknown",
			args: args{
				taints:         nil,
				taintsToAdd:    []*corev1.Taint{unreachableTaintTemplate.DeepCopy()},
				taintsToRemove: []*corev1.Taint{notReadyTaintTemplate.DeepCopy()},
			},
			wantTaints: []corev1.Taint{*unreachableTaintTemplate},
		},
		{
			name: "ready condition from false to unknown",
			args: args{
				taints:         []corev1.Taint{*notReadyTaintTemplate},
				taintsToAdd:    []*corev1.Taint{unreachableTaintTemplate.DeepCopy()},
				taintsToRemove: []*corev1.Taint{notReadyTaintTemplate.DeepCopy()},
			},
			wantTaints: []corev1.Taint{*unreachableTaintTemplate},
		},
		{
			name: "ready condition from false to true",
			args: args{
				taints:         []corev1.Taint{*notReadyTaintTemplate},
				taintsToAdd:    []*corev1.Taint{},
				taintsToRemove: []*corev1.Taint{notReadyTaintTemplate.DeepCopy(), unreachableTaintTemplate.DeepCopy()},
			},
			wantTaints: nil,
		},
		{
			name: "ready condition from unknown to true",
			args: args{
				taints:         []corev1.Taint{*unreachableTaintTemplate},
				taintsToAdd:    []*corev1.Taint{},
				taintsToRemove: []*corev1.Taint{notReadyTaintTemplate.DeepCopy(), unreachableTaintTemplate.DeepCopy()},
			},
			wantTaints: nil,
		},
		{
			name: "ready condition from unknown to false",
			args: args{
				taints:         []corev1.Taint{*unreachableTaintTemplate},
				taintsToAdd:    []*corev1.Taint{notReadyTaintTemplate.DeepCopy()},
				taintsToRemove: []*corev1.Taint{unreachableTaintTemplate.DeepCopy()},
			},
			wantTaints: []corev1.Taint{*notReadyTaintTemplate},
		},
		{
			name: "clusterTaintsToAdd is nil and clusterTaintsToRemove is nil",
			args: args{
				taints:         []corev1.Taint{*unreachableTaintTemplate},
				taintsToAdd:    []*corev1.Taint{unreachableTaintTemplate.DeepCopy()},
				taintsToRemove: []*corev1.Taint{notReadyTaintTemplate.DeepCopy()},
			},
			wantTaints: []corev1.Taint{*unreachableTaintTemplate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &clusterv1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{Name: "member"},
				Spec: clusterv1alpha1.ClusterSpec{
					Taints: tt.args.taints,
				},
			}

			taints := SetCurrentClusterTaints(tt.args.taintsToAdd, tt.args.taintsToRemove, cluster)
			if len(taints) != len(tt.wantTaints) {
				t.Errorf("Cluster gotTaints = %v, want %v", taints, tt.wantTaints)
			}
			for i := range taints {
				if taints[i].Key != tt.wantTaints[i].Key ||
					taints[i].Value != tt.wantTaints[i].Value ||
					taints[i].Effect != tt.wantTaints[i].Effect {
					t.Errorf("Cluster gotTaints = %v, want %v", taints, tt.wantTaints)
				}
			}
		})
	}
}

func TestTaintExists(t *testing.T) {
	type args struct {
		taints      []corev1.Taint
		taintToFind *corev1.Taint
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "exist",
			args: args{
				taints: []corev1.Taint{
					{
						Key:    clusterv1alpha1.TaintClusterUnreachable,
						Effect: corev1.TaintEffectNoExecute,
					},
					{
						Key:    clusterv1alpha1.TaintClusterNotReady,
						Effect: corev1.TaintEffectNoExecute,
					},
				},
				taintToFind: &corev1.Taint{
					Key:    clusterv1alpha1.TaintClusterUnreachable,
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			want: true,
		},
		{
			name: "not exist",
			args: args{
				taints: []corev1.Taint{
					{
						Key:    clusterv1alpha1.TaintClusterUnreachable,
						Effect: corev1.TaintEffectNoExecute,
					},
					{
						Key:    clusterv1alpha1.TaintClusterNotReady,
						Effect: corev1.TaintEffectNoExecute,
					},
				},
				taintToFind: &corev1.Taint{
					Key:    clusterv1alpha1.TaintClusterNotReady,
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TaintExists(tt.args.taints, tt.args.taintToFind); got != tt.want {
				t.Errorf("TaintExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTolerationExists(t *testing.T) {
	type args struct {
		tolerations      []corev1.Toleration
		tolerationToFind *corev1.Toleration
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not exist",
			args: args{
				tolerations: []corev1.Toleration{
					{
						Key:      clusterv1alpha1.TaintClusterUnreachable,
						Effect:   corev1.TaintEffectNoExecute,
						Operator: corev1.TolerationOpEqual,
						Value:    "foo",
					},
					{
						Key:      clusterv1alpha1.TaintClusterNotReady,
						Effect:   corev1.TaintEffectNoExecute,
						Operator: corev1.TolerationOpEqual,
						Value:    "foo",
					},
				},
				tolerationToFind: &corev1.Toleration{
					Key:      clusterv1alpha1.TaintClusterNotReady,
					Effect:   corev1.TaintEffectNoSchedule,
					Operator: corev1.TolerationOpEqual,
					Value:    "foo",
				},
			},
			want: false,
		},
		{
			name: "exist",
			args: args{
				tolerations: []corev1.Toleration{
					{
						Key:      clusterv1alpha1.TaintClusterUnreachable,
						Effect:   corev1.TaintEffectNoExecute,
						Operator: corev1.TolerationOpEqual,
						Value:    "foo",
					},
					{
						Key:      clusterv1alpha1.TaintClusterNotReady,
						Effect:   corev1.TaintEffectNoExecute,
						Operator: corev1.TolerationOpEqual,
						Value:    "foo",
					},
				},
				tolerationToFind: &corev1.Toleration{
					Key:      clusterv1alpha1.TaintClusterNotReady,
					Effect:   corev1.TaintEffectNoExecute,
					Operator: corev1.TolerationOpEqual,
					Value:    "foo",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TolerationExists(tt.args.tolerations, tt.args.tolerationToFind); got != tt.want {
				t.Errorf("TolerationExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddTolerations(t *testing.T) {
	placement := &policyv1alpha1.Placement{
		ClusterTolerations: []corev1.Toleration{},
	}

	toleration1 := &corev1.Toleration{
		Key:      "key1",
		Operator: corev1.TolerationOpEqual,
		Value:    "value1",
		Effect:   corev1.TaintEffectNoSchedule,
	}
	toleration2 := &corev1.Toleration{
		Key:      "key2",
		Operator: corev1.TolerationOpEqual,
		Value:    "value2",
		Effect:   corev1.TaintEffectNoSchedule,
	}

	AddTolerations(placement, toleration1, toleration2)

	assert.Equal(t, 2, len(placement.ClusterTolerations))
	assert.Equal(t, *toleration1, placement.ClusterTolerations[0])
	assert.Equal(t, *toleration2, placement.ClusterTolerations[1])
}

func TestHasNoExecuteTaints(t *testing.T) {
	tests := []struct {
		name   string
		taints []corev1.Taint
		want   bool
	}{
		{
			name: "has NoExecute taints",
			taints: []corev1.Taint{
				{
					Key:    clusterv1alpha1.TaintClusterUnreachable,
					Effect: corev1.TaintEffectNoExecute,
				},
				{
					Key:    clusterv1alpha1.TaintClusterNotReady,
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			want: true,
		},
		{
			name: "no NoExecute taints",
			taints: []corev1.Taint{
				{
					Key:    clusterv1alpha1.TaintClusterUnreachable,
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
				{
					Key:    clusterv1alpha1.TaintClusterNotReady,
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasNoExecuteTaints(tt.taints); got != tt.want {
				t.Errorf("HasNoExecuteTaints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNoExecuteTaints(t *testing.T) {
	tests := []struct {
		name   string
		taints []corev1.Taint
		want   []corev1.Taint
	}{
		{
			name: "has NoExecute taints",
			taints: []corev1.Taint{
				{
					Key:    clusterv1alpha1.TaintClusterUnreachable,
					Effect: corev1.TaintEffectNoExecute,
				},
				{
					Key:    clusterv1alpha1.TaintClusterNotReady,
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			want: []corev1.Taint{
				{
					Key:    clusterv1alpha1.TaintClusterUnreachable,
					Effect: corev1.TaintEffectNoExecute,
				},
			},
		},
		{
			name: "no NoExecute taints",
			taints: []corev1.Taint{
				{
					Key:    clusterv1alpha1.TaintClusterUnreachable,
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
				{
					Key:    clusterv1alpha1.TaintClusterNotReady,
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNoExecuteTaints(tt.taints); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNoExecuteTaints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMinTolerationTimeWithCurrentTime(t *testing.T) {
	// Use fixed time for deterministic testing
	fixedTime := time.Date(2025, 9, 23, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		noExecuteTaints []corev1.Taint
		usedTolerantion []corev1.Toleration
		wantResult      time.Duration
	}{
		{
			name:            "no noExecuteTaints",
			noExecuteTaints: []corev1.Taint{},
			usedTolerantion: []corev1.Toleration{
				{
					Key:               "key",
					Operator:          corev1.TolerationOpExists,
					Effect:            corev1.TaintEffectNoExecute,
					TolerationSeconds: &[]int64{60}[0],
				},
			},
			wantResult: -1,
		},
		{
			name: "no usedTolerations",
			noExecuteTaints: []corev1.Taint{
				{
					Key:       "key",
					Value:     "value",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &metav1.Time{Time: fixedTime},
				},
			},
			usedTolerantion: []corev1.Toleration{},
			wantResult:      0,
		},
		{
			name: "with noExecuteTaints and usedTolerations",
			noExecuteTaints: []corev1.Taint{
				{
					Key:       "key",
					Value:     "value",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &metav1.Time{Time: fixedTime},
				},
			},
			usedTolerantion: []corev1.Toleration{
				{
					Key:               "key",
					Operator:          corev1.TolerationOpExists,
					Effect:            corev1.TaintEffectNoExecute,
					TolerationSeconds: &[]int64{60}[0],
				},
			},
			wantResult: 60,
		},
		{
			name: "usedTolerantion.TolerationSeconds is nil",
			noExecuteTaints: []corev1.Taint{
				{
					Key:       "key",
					Value:     "value",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &metav1.Time{Time: fixedTime},
				},
			},
			usedTolerantion: []corev1.Toleration{
				{
					Key:               "key",
					Operator:          corev1.TolerationOpExists,
					Effect:            corev1.TaintEffectNoExecute,
					TolerationSeconds: nil,
				},
			},
			wantResult: -1,
		},
		{
			name: "noExecuteTaints.TimeAdded is nil",
			noExecuteTaints: []corev1.Taint{
				{
					Key:       "key",
					Value:     "value",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: nil,
				},
			},
			usedTolerantion: []corev1.Toleration{
				{
					Key:               "key",
					Operator:          corev1.TolerationOpExists,
					Effect:            corev1.TaintEffectNoExecute,
					TolerationSeconds: &[]int64{60}[0],
				},
			},
			wantResult: -1,
		},
		{
			name: "find the latest trigger time",
			noExecuteTaints: []corev1.Taint{
				{
					Key:       "key1",
					Value:     "value1",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &metav1.Time{Time: fixedTime},
				},
				{
					Key:       "key2",
					Value:     "value2",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &metav1.Time{Time: fixedTime},
				},
			},
			usedTolerantion: []corev1.Toleration{
				{
					Key:               "key1",
					Operator:          corev1.TolerationOpExists,
					Effect:            corev1.TaintEffectNoExecute,
					TolerationSeconds: &[]int64{120}[0],
				},
				{
					Key:               "key2",
					Operator:          corev1.TolerationOpExists,
					Effect:            corev1.TaintEffectNoExecute,
					TolerationSeconds: &[]int64{60}[0],
				},
			},
			wantResult: 60,
		},
		{
			name: "trigger time is up",
			noExecuteTaints: []corev1.Taint{
				{
					Key:    "key",
					Value:  "value",
					Effect: corev1.TaintEffectNoExecute,
					TimeAdded: &metav1.Time{
						Time: time.Date(2025, 9, 23, 11, 58, 0, 0, time.UTC), // 2 minutes before fixedTime
					},
				},
			},
			usedTolerantion: []corev1.Toleration{
				{
					Key:               "key",
					Operator:          corev1.TolerationOpExists,
					Effect:            corev1.TaintEffectNoExecute,
					TolerationSeconds: &[]int64{60}[0], // 60 seconds tolerance
				},
			},
			wantResult: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMinTolerationTimeWithCurrentTime(tt.noExecuteTaints, tt.usedTolerantion, fixedTime)

			// Convert expected duration from seconds to time.Duration for comparison
			var expectedDuration time.Duration
			if tt.wantResult > 0 {
				expectedDuration = tt.wantResult * time.Second
			} else {
				expectedDuration = tt.wantResult
			}

			if result != expectedDuration {
				t.Errorf("GetMinTolerationTimeWithCurrentTime() = %v, want %v", result, expectedDuration)
			}
		})
	}
}

func TestGetMatchingTolerations(t *testing.T) {
	tests := []struct {
		name                  string
		taints                []corev1.Taint
		tolerations           []corev1.Toleration
		wantActual            bool
		wantActualTolerations []corev1.Toleration
	}{
		{
			name:   "taints is nil",
			taints: []corev1.Taint{},
			tolerations: []corev1.Toleration{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			wantActual:            true,
			wantActualTolerations: []corev1.Toleration{},
		},
		{
			name: "tolerations is nil",
			taints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			tolerations:           []corev1.Toleration{},
			wantActual:            false,
			wantActualTolerations: []corev1.Toleration{},
		},
		{
			name: "tolerated is true",
			taints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			tolerations: []corev1.Toleration{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			wantActual: true,
			wantActualTolerations: []corev1.Toleration{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
		{
			name: "tolerated is false",
			taints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			tolerations: []corev1.Toleration{
				{
					Key:    "key1",
					Value:  "value_1",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "key2",
					Value:  "value_2",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			wantActual:            false,
			wantActualTolerations: []corev1.Toleration{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, actualTolerations := GetMatchingTolerations(tt.taints, tt.tolerations)
			if actual != tt.wantActual || !reflect.DeepEqual(actualTolerations, tt.wantActualTolerations) {
				t.Errorf("GetMatchingTolerations(%v, %v) = (%v, %v), expected (%v, %v)", tt.taints, tt.tolerations, actual, actualTolerations, tt.wantActual, tt.wantActualTolerations)
			}
		})
	}
}

func TestGenerateTaintsMessage(t *testing.T) {
	tests := []struct {
		name     string
		taints   []corev1.Taint
		expected string
	}{
		{
			name:     "no taint",
			taints:   []corev1.Taint{},
			expected: "cluster now does not have taints",
		},
		{
			name: "single taint with value",
			taints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expected: "cluster now has taints([{Key:key1,Value:value1,Effect:NoSchedule}])",
		},
		{
			name: "multiple taints with no value",
			taints: []corev1.Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "key2",
					Value:  "",
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
			},
			expected: "cluster now has taints([{Key:key1,Value:value1,Effect:NoSchedule},{Key:key2,Effect:PreferNoSchedule}])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateTaintsMessage(tt.taints)
			if result != tt.expected {
				t.Errorf("test failed: %s, expected: %s, actual: %s", tt.name, tt.expected, result)
			}
		})
	}
}
