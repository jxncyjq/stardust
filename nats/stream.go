package nats

import (
	"errors"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func (s *NatsConnection) ensureStreamSubjects(streamName string, subjects []string) error {
	if !s.useStream {
		return nats.ErrNoStreamResponse
	}
	if streamName == "" {
		return nats.ErrStreamNameRequired
	}

	js := s.getJS()
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		if !errors.Is(err, nats.ErrStreamNotFound) {
			return err
		}

		s.streamInfo, err = js.AddStream(&nats.StreamConfig{
			Name:      streamName,
			Subjects:  subjects,
			Retention: nats.WorkQueuePolicy,
		})
		if err != nil {
			s.logger.Error("Failed to add stream",
				zap.String("stream", streamName),
				zap.Strings("subjects", subjects),
				zap.Error(err))
			return err
		}
		s.logger.Info("Stream added",
			zap.String("stream", streamName),
			zap.Strings("subjects", subjects))
		return nil
	}

	merged, changed := mergeStreamSubjects(stream.Config.Subjects, subjects)
	if !changed {
		s.streamInfo = stream
		return nil
	}

	cfg := stream.Config
	cfg.Subjects = merged
	s.streamInfo, err = js.UpdateStream(&cfg)
	if err != nil {
		s.logger.Error("Failed to update stream subjects",
			zap.String("stream", streamName),
			zap.Strings("current_subjects", stream.Config.Subjects),
			zap.Strings("required_subjects", subjects),
			zap.Strings("merged_subjects", merged),
			zap.Error(err))
		return err
	}
	s.logger.Info("Stream subjects updated",
		zap.String("stream", streamName),
		zap.Strings("subjects", merged))
	return nil
}
