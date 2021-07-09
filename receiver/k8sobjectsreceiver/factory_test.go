// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8sobjectsreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
)

func TestFactory(t *testing.T) {
	f := NewFactory()
	require.Equal(t, config.Type("k8s_cluster"), f.Type())

	cfg := f.CreateDefaultConfig()
	rCfg, ok := cfg.(*Config)
	require.True(t, ok)

	require.Equal(t, &Config{
		ReceiverSettings:           config.NewReceiverSettings(config.NewID(typeStr)),
		CollectionInterval:         10 * time.Second,
		NodeConditionTypesToReport: defaultNodeConditionsToReport,
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
	}, rCfg)

	r, err := f.CreateTracesReceiver(
		context.Background(), component.ReceiverCreateSettings{},
		&config.ReceiverSettings{}, consumertest.NewNop(),
	)
	require.Error(t, err)
	require.Nil(t, r)

	// Fails with bad K8s Config.
	r, err = f.CreateMetricsReceiver(
		context.Background(), component.ReceiverCreateSettings{},
		rCfg, consumertest.NewNop(),
	)
	require.Error(t, err)
	require.Nil(t, r)

	// Override for tests.
	rCfg.makeClient = func(apiConf k8sconfig.APIConfig) (kubernetes.Interface, error) {
		return nil, nil
	}
	r, err = f.CreateMetricsReceiver(
		context.Background(), component.ReceiverCreateSettings{Logger: zap.NewNop()},
		rCfg, consumertest.NewNop(),
	)
	require.NoError(t, err)
	require.NotNil(t, r)

	// Test metadata exporters setup.
	ctx := context.Background()
	require.NoError(t, r.Start(ctx, nopHostWithExporters{}))
	require.NoError(t, r.Shutdown(ctx))
	rCfg.MetadataExporters = []string{"nop/withoutmetadata"}
	require.Error(t, r.Start(context.Background(), nopHostWithExporters{}))
}

// nopHostWithExporters mocks a receiver.ReceiverHost for test purposes.
type nopHostWithExporters struct {
}

var _ component.Host = (*nopHostWithExporters)(nil)

func (n nopHostWithExporters) ReportFatalError(error) {
}

func (n nopHostWithExporters) GetFactory(component.Kind, config.Type) component.Factory {
	return nil
}

func (n nopHostWithExporters) GetExtensions() map[config.ComponentID]component.Extension {
	return nil
}

func (n nopHostWithExporters) GetExporters() map[config.DataType]map[config.ComponentID]component.Exporter {
	return map[config.DataType]map[config.ComponentID]component.Exporter{
		config.MetricsDataType: {
			config.NewIDWithName("nop", "withoutmetadata"): MockExporter{},
			config.NewIDWithName("nop", "withmetadata"):    mockExporterWithK8sMetadata{},
		},
	}
}
