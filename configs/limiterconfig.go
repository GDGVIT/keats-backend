package configs

import(
	"time"

	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/spf13/viper"

	"github.com/Krishap-s/keats-backend/errors"
)

var LimiterConfig = limiter.Config{
	Max: viper.GetInt("MAX_REQUESTS"),
	Duration: time.Duration(viper.GetInt("TIME_PERIOD_IN_MINUTES")) * time.Minute,
	LimitReached:errors.TooManyRequestsError,
}

