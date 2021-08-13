package pipeline

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type AutoJsonConverterConfig struct {
	FieldTips map[string]Field
}

type AutoJsonConverter struct {
	config AutoJsonConverterConfig
}

func NewAutoJsonConverter(c AutoJsonConverterConfig) *AutoJsonConverter {
	return &AutoJsonConverter{config: c}
}

// Automatic conversion works this way:
// * Time added automatically
// * Nulls dropped
// To preserve nulls we need FieldTips from a user.
// Custom time can be injected on Processor stage theoretically.
// Custom labels can be injected on Processor stage theoretically.
func (c *AutoJsonConverter) Convert(_ context.Context, vars Vars, body []byte) ([]*data.Frame, error) {
	frame, err := JSONDocToFrame(vars.Path, body, c.config.FieldTips)
	return []*data.Frame{frame}, err
}
