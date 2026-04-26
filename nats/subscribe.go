package nats

import (
	"errors"
	"strings"
	"time"

	"github.com/jxncyjq/stardust/logs"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

//func (s *NatsConnection) Subscribe(subject, durableName string, handler func(msg *nats.Msg)) (err error) {
//	if subject == "" || durableName == "" || handler == nil {
//		s.logger.Error("Invalid subscribe parameters")
//		return errors.New("invalid subscribe parameters")
//	}
//
//	if s.useStream {
//		// 确保 Stream 已正确创建
//		if err := s.EnsureStream(); err != nil {
//			return err
//		}
//
//		// 对于 JetStream，使用 Subscribe 而不是 QueueSubscribe
//		subjectObj, err := s.js.Subscribe(subject, handler, nats.Durable(durableName))
//		if err != nil {
//			s.logger.Error("Error subscribing to subject", logs.ErrorInfo(err))
//			return err
//		}
//		s.logger.Info("Subscribed to subject", logs.String("subject", subject), logs.String("durable", durableName))
//		s.subject = append(s.subject, subjectObj)
//		return nil
//	} else {
//		// 对于非 JetStream，使用普通的 QueueSubscribe
//		subjectObj, err := s.conn.QueueSubscribe(subject, durableName, handler)
//		if err != nil {
//			s.logger.Error("Error subscribing to subject", logs.ErrorInfo(err))
//			return err
//		}
//		s.logger.Info("Subscribed to subject", logs.String("subject", subject))
//		s.subject = append(s.subject, subjectObj)
//		return nil
//	}
//}

func (s *NatsConnection) AddStream(streamName string, subjects []string) error {
	if !s.useStream {
		return nats.ErrNoStreamResponse
	}
	js := s.getJS()
	stream, err := js.StreamInfo(streamName)
	if err != nil && stream != nil {
		return err
	}
	if stream == nil {
		s.streamInfo, err = js.AddStream(&nats.StreamConfig{
			Name:      streamName,
			Subjects:  subjects,
			Retention: nats.WorkQueuePolicy, // 使用工作队列策略，确保每条消息只能被消费一次
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
	}
	return nil
}

func (s *NatsConnection) AddConsumer(streamName, durableName string, subjects ...string) error {
	if !s.useStream {
		return nats.ErrNoStreamResponse
	}

	js := s.getJS()
	// 首先检查 Consumer 是否已存在
	consumer, err := js.ConsumerInfo(streamName, durableName)
	if err == nil && consumer != nil {
		// Consumer 已存在
		s.logger.Info("Consumer already exists",
			zap.String("stream", streamName),
			zap.String("durable", durableName),
			zap.Any("subject", subjects))
		return nil
	}

	if err != nil && !strings.Contains(err.Error(), "consumer not found") {
		// 只有不是 "not found" 的错误才返回
		s.logger.Error("Failed to get consumer info",
			zap.String("stream", streamName),
			zap.String("durable", durableName),
			zap.Error(err))
		return err
	}
	consumerCfg := &nats.ConsumerConfig{
		Durable:       durableName,
		AckPolicy:     nats.AckExplicitPolicy,
		DeliverPolicy: nats.DeliverAllPolicy,
	}

	// Consumer 不存在，创建新的
	if len(subjects) > 0 {
		if len(subjects) > 1 {
			consumerCfg.FilterSubjects = subjects
		} else {
			consumerCfg.FilterSubject = subjects[0]
		}
		s.logger.Info("Creating consumer with filter subjects",
			zap.String("stream", streamName),
			zap.String("durable", durableName),
			zap.Any("subjects", subjects))
	} else {
		s.logger.Info("Creating consumer without filter subjects (will consume all subjects)",
			zap.String("stream", streamName),
			zap.String("durable", durableName))
	}

	_, err = js.AddConsumer(streamName, consumerCfg)
	if err != nil {
		s.logger.Error("Failed to add consumer",
			zap.String("stream", streamName),
			zap.String("durable", durableName),
			zap.Any("subject", subjects),
			zap.Error(err))
		return err
	}
	s.logger.Info("Consumer added",
		zap.String("stream", streamName),
		zap.String("durable", durableName),
		zap.Any("subject", subjects))
	return nil
}

func (s *NatsConnection) GetJetStream() nats.JetStreamContext {
	return s.getJS()
}

func (s *NatsConnection) GetNativeConn() *nats.Conn {
	return s.conn
}

func (s *NatsConnection) Start() {
	if s.useStream {
		// 检查并仅在不存在时添加 Stream
		stream, err := s.getJS().StreamInfo(s.config.StreamName)
		if err != nil && stream == nil {
			if err := s.AddStream(s.config.StreamName, s.config.Subject); err != nil {
				s.logger.Error("create stream failed, aborting start", zap.Error(err))
				return
			}
		}
	}

	// 创建连接状态监听通道
	reconnectChan := make(chan bool, 1)
	disconnectChan := make(chan bool, 1)

	// 设置连接状态回调
	s.conn.SetReconnectHandler(func(nc *nats.Conn) {
		s.logger.Info("NATS reconnected",
			zap.String("url", nc.ConnectedUrl()),
			zap.String("name", s.name))
		select {
		case reconnectChan <- true:
		default:
		}
	})

	s.conn.SetDisconnectErrHandler(func(nc *nats.Conn, err error) {
		s.logger.Warn("NATS disconnected",
			zap.String("url", nc.ConnectedUrl()),
			zap.String("name", s.name),
			zap.Error(err))
		select {
		case disconnectChan <- true:
		default:
		}
	})

	s.conn.SetClosedHandler(func(nc *nats.Conn) {
		s.logger.Error("NATS connection closed",
			zap.String("name", s.name))
	})

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("NATS connection context done, stopping...")
			return
		case <-reconnectChan:
			s.logger.Info("Handling NATS reconnection", zap.String("name", s.name))
			// 重连后重新初始化 JetStream 相关资源
			if s.useStream {
				// 重新获取 JetStream 上下文
				js, err := s.conn.JetStream(nats.PublishAsyncMaxPending(100))
				if err != nil {
					s.logger.Error("Failed to recreate JetStream context after reconnect", zap.Error(err))
					continue
				}
				s.jsMu.Lock()
				s.js = js
				s.jsMu.Unlock()

				// 确保 Stream 存在
				if err := s.EnsureStream(); err != nil {
					s.logger.Error("Failed to ensure stream after reconnect", zap.Error(err))
				}
			}
		case <-disconnectChan:
			// 连接断开时的处理逻辑（如果需要的话）
			// 这里可以添加一些清理或通知逻辑
			s.logger.Info("Handling NATS disconnection", zap.String("name", s.name))
			//case msg := <-s.messageChan:
			//	if msg != nil {
			//		handler, ok := s.handlers[msg.Subject]
			//		if ok && handler != nil {
			//			handler(msg)
			//		} else {
			//			s.logger.Warn("No handler for subject", logs.String("subject", msg.Subject))
			//		}
			//		// JetStream 消息需要手动 ACK
			//		if s.useStream {
			//			if err := msg.Ack(); err != nil {
			//				s.logger.Error("Ack error in Start", logs.ErrorInfo(err))
			//			}
			//		}
			//	}
		}
	}
}

func (s *NatsConnection) Stop() {
	s.cancel()
	if s.conn != nil {
		s.conn.Close()
		s.logger.Info("NATS connection closed", logs.String("name", s.name), logs.String("url", s.url))
	}
}

func (s *NatsConnection) StartSubscription(subject, durableName string, handler func(msg *nats.Msg)) error {
	if handler == nil || subject == "" {
		s.logger.Error("Invalid subscribe parameters")
		return errors.New("invalid subscribe parameters")
	}

	// 注册 handler 到映射表
	s.handlersMu.Lock()
	s.handlers[subject] = handler
	s.handlersMu.Unlock()

	if s.useStream {
		// JetStream 场景
		if err := s.AddConsumer(s.config.StreamName, durableName, subject); err != nil {
			s.logger.Error("AddConsumer error", logs.ErrorInfo(err))
			return err
		}

		sub, err := s.getJS().PullSubscribe(subject, durableName, nats.BindStream(s.config.StreamName))
		if err != nil {
			s.logger.Error("JetStream PullSubscribe error", logs.ErrorInfo(err))
			return err
		}
		s.logger.Info("JetStream PullSubscribe started", logs.String("subject", subject), logs.String("durable", durableName))
		s.subject = append(s.subject, sub)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("Fetch goroutine panic recovered", zap.Any("panic", r), logs.String("subject", subject))
				}
			}()

			for {
				select {
				case <-s.ctx.Done():
					s.logger.Info("Fetch goroutine stopping due to context cancellation", logs.String("subject", subject))
					return
				default:
				}

				msgs, err := sub.Fetch(10)
				if err != nil {
					if errors.Is(err, nats.ErrTimeout) {
						time.Sleep(50 * time.Millisecond)
						continue
					}
					s.logger.Error("Fetch error, retrying", logs.ErrorInfo(err), logs.String("subject", subject))
					time.Sleep(1 * time.Second)
					continue
				}

				for _, msg := range msgs {
					// 确保只有匹配的 subject 才处理消息
					if msg.Subject == subject {
						s.logger.Debug("Processing message", logs.String("subject", msg.Subject), logs.String("data", string(msg.Data)))
						handler(msg)
						if err := msg.Ack(); err != nil {
							s.logger.Error("Ack error", logs.ErrorInfo(err), logs.String("subject", subject))
						}
					} else {
						s.logger.Warn("Received message for wrong subject",
							logs.String("expected", subject),
							logs.String("actual", msg.Subject))
						// 如果 subject 不匹配，仍然需要 Ack 以避免消息重新投递
						//if err := msg.Ack(); err != nil {
						//	s.logger.Error("Ack error for wrong subject", logs.ErrorInfo(err))
						//}
					}
				}
			}
		}()
		return nil
	} else {
		// 非 JetStream 场景
		ch := make(chan *nats.Msg, 64)
		sub, err := s.conn.QueueSubscribe(subject, durableName, func(msg *nats.Msg) {
			ch <- msg
		})
		if err != nil {
			s.logger.Error("Error subscribing to subject", logs.ErrorInfo(err))
			return err
		}
		s.logger.Info("Subscribed to subject", logs.String("subject", subject), logs.String("durable", durableName))
		s.subject = append(s.subject, sub)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("ChanSubscribe goroutine panic recovered", zap.Any("panic", r), logs.String("subject", subject))
				}
			}()

			for {
				select {
				case <-s.ctx.Done():
					s.logger.Info("ChanSubscribe goroutine stopping due to context cancellation", logs.String("subject", subject))
					return
				case msg := <-ch:
					if msg != nil {
						handler(msg)
					}
				}
			}
		}()
		return nil
	}
}

func (s *NatsConnection) StopSubscription(subject string) error {
	for i, sub := range s.subject {
		if sub.Subject == subject {
			err := sub.Unsubscribe()
			if err != nil {
				return err
			}
			s.logger.Info("Unsubscribed from subject", logs.String("subject", subject))
			s.subject = append(s.subject[:i], s.subject[i+1:]...)
			return nil
		}
	}
	s.logger.Warn("Subject not found to unsubscribe", logs.String("subject", subject))
	return errors.New("subject not found")
}

func (s *NatsConnection) StopAllSubscriptions() error {
	for _, sub := range s.subject {
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Error("Failed to unsubscribe from subject", logs.String("subject", sub.Subject), logs.ErrorInfo(err))
			return err
		}
		s.logger.Info("Unsubscribed from subject", logs.String("subject", sub.Subject))
	}
	s.subject = nil
	return nil
}
