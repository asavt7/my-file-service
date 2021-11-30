package health

import "context"

type Healthchecker interface {
	ReadinessProbe(ctx context.Context) error
	LivenessProbe(ctx context.Context) error
}
