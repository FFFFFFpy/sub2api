package schema

import (
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/require"
)

func TestUserPlatformQuotaAllowsExternalOpenAICompatiblePlatform(t *testing.T) {
	validator := requireStringFieldValidators(t, UserPlatformQuota{}.Fields(), "platform")

	require.NoError(t, validator("external_openai_compatible"))
	require.Error(t, validator("volcengine_coding"))
	require.Error(t, validator("xunfei_coding"))
	require.Error(t, validator("unknown"))
}

func requireStringFieldValidators(t *testing.T, fields []ent.Field, name string) func(string) error {
	t.Helper()

	for _, entField := range fields {
		descriptor := entField.Descriptor()
		if descriptor.Name != name {
			continue
		}
		require.NotEmpty(t, descriptor.Validators, "field %s should include validators", name)
		return func(value string) error {
			for _, rawValidator := range descriptor.Validators {
				validator, ok := rawValidator.(func(string) error)
				require.True(t, ok, "field %s validator should be func(string) error", name)
				if err := validator(value); err != nil {
					return err
				}
			}
			return nil
		}
	}

	require.Failf(t, "missing field validators", "schema should include field %s", name)
	return nil
}
