/*
Copyright 2018 The Knative Authors.
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
package metrics

import (
	"reflect"
	"testing"

	. "github.com/knative/pkg/logging/testing"
)

const (
	testProj      = "test-project"
	servingDomain = "serving.knative.dev"
	badDomain     = "test.domain"
	testComponent = "testComponent"
)

func TestGetMetricsConfig(t *testing.T) {
	errorTests := []struct {
		name        string
		cm          map[string]string
		domain      string
		component   string
		expectedErr string
	}{{
		name:        "backendKeyMissing",
		cm:          map[string]string{"": ""},
		domain:      servingDomain,
		component:   testComponent,
		expectedErr: "metrics.backend-destination key is missing",
	}, {
		name:        "stackdriverProjectIDMissing",
		cm:          map[string]string{"metrics.backend-destination": "stackdriver"},
		domain:      servingDomain,
		component:   testComponent,
		expectedErr: "For backend stackdriver, metrics.stackdriver-project-id field must exist and cannot be empty",
	}, {
		name: "stackdriverProjectIDEmpty",
		cm: map[string]string{
			"metrics.backend-destination":    "stackdriver",
			"metrics.stackdriver-project-id": "",
		},
		domain:      servingDomain,
		component:   testComponent,
		expectedErr: "For backend stackdriver, metrics.stackdriver-project-id field must exist and cannot be empty",
	}, {
		name: "unsupportedBackend",
		cm: map[string]string{
			"metrics.backend-destination":    "unsupported",
			"metrics.stackdriver-project-id": testProj,
		},
		domain:      servingDomain,
		component:   testComponent,
		expectedErr: "Unsupported metrics backend value \"unsupported\"",
	}, {
		name: "invalidDomain",
		cm: map[string]string{
			"metrics.backend-destination": "prometheus",
		},
		domain:      "abc.knative.dev",
		component:   testComponent,
		expectedErr: "Invalid metrics domain name \"abc.knative.dev\"",
	}, {
		name: "invalidComponent",
		cm: map[string]string{
			"metrics.backend-destination": "prometheus",
		},
		domain:      servingDomain,
		component:   "",
		expectedErr: "Metrics component name cannot be empty",
	}}

	for _, test := range errorTests {
		_, err := getMetricsConfig(test.cm, test.domain, test.component, TestLogger(t))
		if err.Error() != test.expectedErr {
			t.Errorf("In test %v, wanted err: %v, got: %v", test.name, test.expectedErr, err)
		}
	}

	successTests := []struct {
		name           string
		cm             map[string]string
		domain         string
		component      string
		expectedConfig metricsConfig
	}{{
		name:      "validPrometheus",
		cm:        map[string]string{"metrics.backend-destination": "prometheus"},
		domain:    servingDomain,
		component: testComponent,
		expectedConfig: metricsConfig{
			domain:             servingDomain,
			component:          testComponent,
			backendDestination: Prometheus},
	}, {
		name: "validStackdriver",
		cm: map[string]string{"metrics.backend-destination": "stackdriver",
			"metrics.stackdriver-project-id": testProj},
		domain:    servingDomain,
		component: testComponent,
		expectedConfig: metricsConfig{
			domain:               servingDomain,
			component:            testComponent,
			backendDestination:   Stackdriver,
			stackdriverProjectID: testProj},
	}, {
		name: "validCapitalStackdriver",
		cm: map[string]string{"metrics.backend-destination": "STACKDRIVER",
			"metrics.stackdriver-project-id": testProj},
		domain:    servingDomain,
		component: testComponent,
		expectedConfig: metricsConfig{
			domain:               servingDomain,
			component:            testComponent,
			backendDestination:   Stackdriver,
			stackdriverProjectID: testProj},
	}}

	for _, test := range successTests {
		mc, err := getMetricsConfig(test.cm, test.domain, test.component, TestLogger(t))
		if err != nil {
			t.Errorf("In test %v, wanted valid config %v, got error %v", test.name, test.expectedConfig, err)
		}
		if !reflect.DeepEqual(*mc, test.expectedConfig) {
			t.Errorf("In test %v, wanted config %v, got config %v", test.name, test.expectedConfig, *mc)
		}
	}
}
