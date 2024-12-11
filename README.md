go-log

This public go package configures `log/slog` stdlib for json formatted logs suitable for sending to Datadog.

APM `trace_id` is added from request context if available.

example usage:

```
import (
    "log/slog"
    "github.com/mediafly/go-log/log"
)

func init() {
    log.SetupLog(slog.LevelInfo)
    slog.Info(fmt.Sprintf("origin: %v", origin))
}
```