package pipeline

import (
	"context"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type ThresholdOutputConfig struct {
	FieldName  string
	Thresholds []data.Threshold
	Channel    string
}

type ThresholdOutput struct {
	frameStorage  *FrameStorage
	ruleProcessor *RuleProcessor
	config        ThresholdOutputConfig
}

func NewThresholdOutput(frameStorage *FrameStorage, ruleProcessor *RuleProcessor, config ThresholdOutputConfig) *ThresholdOutput {
	return &ThresholdOutput{frameStorage: frameStorage, ruleProcessor: ruleProcessor, config: config}
}

func (l ThresholdOutput) Output(_ context.Context, vars OutputVars, frame *data.Frame) error {
	previousFrame, previousFrameOk, err := l.frameStorage.Get(vars.OrgID, l.config.Channel)
	if err != nil {
		return err
	}
	fieldName := l.config.FieldName

	currentFrameFieldIndex := -1
	for i, f := range frame.Fields {
		if f.Name == fieldName {
			currentFrameFieldIndex = i
		}
	}
	if currentFrameFieldIndex < 0 {
		return nil
	}

	// TODO: support other numeric types.
	value, ok := frame.Fields[currentFrameFieldIndex].At(0).(*float64)
	if !ok {
		return nil
	}

	if value == nil {
		// TODO: what should we do here?
		return nil
	}

	var currentThreshold data.Threshold
	for _, threshold := range l.config.Thresholds {
		if *value >= float64(threshold.Value) {
			currentThreshold = threshold
			continue
		}
		break
	}

	if !previousFrameOk {
		fTime := data.NewFieldFromFieldType(data.FieldTypeTime, 1)
		fTime.Name = "time"
		fTime.Set(0, time.Now())
		f1 := data.NewFieldFromFieldType(data.FieldTypeFloat64, 1)
		f1.Set(0, *value)
		f1.Name = "value"
		f2 := data.NewFieldFromFieldType(data.FieldTypeString, 1)
		f2.Set(0, currentThreshold.State)
		f2.Name = "state"
		stateFrame := data.NewFrame("state", fTime, f1, f2)
		_ = l.frameStorage.Put(vars.OrgID, l.config.Channel, stateFrame)
		return l.ruleProcessor.ProcessFrame(context.Background(), vars.OrgID, l.config.Channel, stateFrame)
	}

	// TODO: support other numeric types.
	previousState, ok := previousFrame.Fields[2].At(0).(string)
	if !ok {
		return nil
	}

	if currentThreshold.State != previousState {
		fTime := data.NewFieldFromFieldType(data.FieldTypeTime, 1)
		fTime.Name = "time"
		fTime.Set(0, time.Now())
		f1 := data.NewFieldFromFieldType(data.FieldTypeFloat64, 1)
		f1.Set(0, *value)
		f1.Name = "value"
		f2 := data.NewFieldFromFieldType(data.FieldTypeString, 1)
		f2.Set(0, currentThreshold.State)
		f2.Name = "state"
		stateFrame := data.NewFrame("state", fTime, f1, f2)
		_ = l.frameStorage.Put(vars.OrgID, l.config.Channel, stateFrame)
		return l.ruleProcessor.ProcessFrame(context.Background(), vars.OrgID, l.config.Channel, stateFrame)
	}

	return nil
}
