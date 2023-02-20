// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package awscloudwatch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"

	"github.com/aws/private-amazon-cloudwatch-agent-staging/plugins/outputs/cloudwatch"
	"github.com/aws/private-amazon-cloudwatch-agent-staging/translator/translate/agent"
	"github.com/aws/private-amazon-cloudwatch-agent-staging/translator/translate/otel/common"
)

func TestTranslator(t *testing.T) {
	agent.Global_Config.Region = "us-east-1"
	agent.Global_Config.Role_arn = "global_arn"
	cwt := NewTranslator()
	require.EqualValues(t, "awscloudwatch", cwt.Type())
	testCases := map[string]struct {
		input       map[string]interface{}
		internal    bool
		credentials map[string]interface{}
		want        *cloudwatch.Config
		wantErr     error
	}{
		"WithMissingKey": {
			input: map[string]interface{}{"logs": map[string]interface{}{}},
			wantErr: &common.MissingKeyError{
				Type:    cwt.Type(),
				JsonKey: common.MetricsKey,
			},
		},
		"WithDefault": {
			input: map[string]interface{}{"metrics": map[string]interface{}{}},
			want: &cloudwatch.Config{
				Namespace:          "CWAgent",
				Region:             "us-east-1",
				ForceFlushInterval: time.Minute,
				MaxValuesPerDatum:  150,
				RoleARN:            "global_arn",
			},
		},
		"WithEndpointOverride": {
			input: map[string]interface{}{"metrics": map[string]interface{}{
				"endpoint_override": "https://monitoring-fips.us-east-1.amazonaws.com",
			}},
			want: &cloudwatch.Config{
				Namespace:          "CWAgent",
				Region:             "us-east-1",
				ForceFlushInterval: time.Minute,
				MaxValuesPerDatum:  150,
				EndpointOverride:   "https://monitoring-fips.us-east-1.amazonaws.com",
				RoleARN:            "global_arn",
			},
		},
		"WithInvalidCredentialFields": {
			input: map[string]interface{}{"metrics": map[string]interface{}{}},
			credentials: map[string]interface{}{
				"access_key":  "access_key",
				"secret_key":  "secret_key",
				"token":       "token",
				"prof":        "invalid field name",
				"shared_cred": "invalid field name",
			},
			want: &cloudwatch.Config{
				Namespace:          "CWAgent",
				Region:             "us-east-1",
				ForceFlushInterval: time.Minute,
				MaxValuesPerDatum:  150,
				RoleARN:            "global_arn",
				AccessKey:          "access_key",
				SecretKey:          "secret_key",
				Token:              "token",
			},
		},
		"WithValidCredentials": {
			input: map[string]interface{}{"metrics": map[string]interface{}{}},
			credentials: map[string]interface{}{
				"access_key":             "access_key",
				"secret_key":             "secret_key",
				"token":                  "token",
				"profile":                "profile",
				"shared_credential_file": "shared",
			},
			want: &cloudwatch.Config{
				Namespace:                "CWAgent",
				Region:                   "us-east-1",
				ForceFlushInterval:       time.Minute,
				MaxValuesPerDatum:        150,
				RoleARN:                  "global_arn",
				AccessKey:                "access_key",
				SecretKey:                "secret_key",
				Token:                    "token",
				Profile:                  "profile",
				SharedCredentialFilename: "shared",
			},
		},
		"WithInternal": {
			input:    getJson(t, filepath.Join("testdata", "config.json")),
			internal: true,
			want: &cloudwatch.Config{
				Namespace:          "namespace",
				Region:             "us-east-1",
				ForceFlushInterval: 30 * time.Second,
				MaxValuesPerDatum:  5000,
				EndpointOverride:   "https://monitoring-fips.us-west-2.amazonaws.com",
				RoleARN:            "metrics_role_arn_value_test",
				RollupDimensions:   [][]string{{"ImageId"}, {"InstanceId", "InstanceType"}, {"d1"}, {}},
				DropOriginConfigs: map[string][]string{
					"cpu":        {"cpu_usage_idle", "time_active"},
					"nvidia_smi": {"utilization_gpu", "temperature_gpu"},
				},
				MetricDecorations: []cloudwatch.MetricDecorationConfig{
					{Category: "cpu", Metric: "usage_idle", Rename: "CPU_USAGE_IDLE", Unit: "unit"},
					{Category: "cpu", Metric: "usage_nice", Unit: "unit"},
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			agent.Global_Config.Internal = testCase.internal
			agent.Global_Config.Credentials = testCase.credentials

			conf := confmap.NewFromStringMap(testCase.input)
			got, err := cwt.Translate(conf, common.TranslatorOptions{})
			require.Equal(t, testCase.wantErr, err)
			if testCase.want != nil {
				require.NoError(t, err)
				gotCfg, ok := got.(*cloudwatch.Config)
				require.True(t, ok)
				require.Equal(t, testCase.want.Namespace, gotCfg.Namespace)
				require.Equal(t, testCase.want.Region, gotCfg.Region)
				require.Equal(t, testCase.want.ForceFlushInterval, gotCfg.ForceFlushInterval)
				require.Equal(t, testCase.want.RoleARN, gotCfg.RoleARN)
				require.Equal(t, testCase.want.AccessKey, gotCfg.AccessKey)
				require.Equal(t, testCase.want.SecretKey, gotCfg.SecretKey)
				require.Equal(t, testCase.want.Token, gotCfg.Token)
				require.Equal(t, testCase.want.Profile, gotCfg.Profile)
				require.Equal(t, testCase.want.SharedCredentialFilename, gotCfg.SharedCredentialFilename)
				require.Equal(t, testCase.want.MaxValuesPerDatum, gotCfg.MaxValuesPerDatum)
				require.Equal(t, testCase.want.RollupDimensions, gotCfg.RollupDimensions)
				require.Equal(t, testCase.want.DropOriginConfigs, gotCfg.DropOriginConfigs)
				require.Equal(t, testCase.want.MetricDecorations, gotCfg.MetricDecorations)
			}
		})
	}
}

func getJson(t *testing.T, path string) map[string]interface{} {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(content, &result))
	return result
}
