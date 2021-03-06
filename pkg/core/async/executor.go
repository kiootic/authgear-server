package async

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
)

type Executor struct {
	tasks map[string]Task
	pool  db.Pool
}

func NewExecutor(dbPool db.Pool) *Executor {
	return &Executor{
		tasks: map[string]Task{},
		pool:  dbPool,
	}
}

func (e *Executor) Register(name string, task Task) {
	e.tasks[name] = task
}

func (e *Executor) Execute(ctx context.Context, spec TaskSpec) {
	ctx = db.InitDBContext(ctx, e.pool)
	task := e.tasks[spec.Name]

	tConfig := config.GetTenantConfig(ctx)

	logHook := logging.NewDefaultLogHook(tConfig.DefaultSensitiveLoggerValues())
	sentryHook := &sentry.LogHook{Hub: sentry.DefaultClient.Hub}
	loggerFactory := logging.NewFactory(logHook, sentryHook)
	logger := loggerFactory.NewLogger("async-executor")
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logger.WithFields(map[string]interface{}{
					"task_name": spec.Name,
					"error":     rec,
				}).Error("unexpected error occurred when running async task")
			}
		}()

		err := task.Run(ctx, spec.Param)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"task_name": spec.Name,
				"error":     err,
			}).Error("error occurred when running async task")
		}
	}()
}
