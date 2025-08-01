package chat

import (
	"log/slog"
)

// Option represents a configuration option for the Service.
type Option func(*serviceOptions)

// WithLogger sets the logger for the service.
func WithLogger(logger *slog.Logger) Option {
	return func(o *serviceOptions) {
		o.logger = logger
	}
}

// WithMaxRoomsPerUser sets the maximum number of rooms a user can create.
func WithMaxRoomsPerUser(max int) Option {
	return func(o *serviceOptions) {
		o.maxRoomsPerUser = max
	}
}

// WithMaxMessagesPerMin sets the maximum number of messages per minute per user.
func WithMaxMessagesPerMin(max int) Option {
	return func(o *serviceOptions) {
		o.maxMessagesPerMin = max
	}
}

// WithDefaultMaxUsers sets the default maximum users for new rooms.
func WithDefaultMaxUsers(max int) Option {
	return func(o *serviceOptions) {
		o.defaultMaxUsers = max
	}
}
