package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trace/trace/config"
	"github.com/trace/trace/internal/application/agents/coach"
	"github.com/trace/trace/internal/application/agents/growth"
	"github.com/trace/trace/internal/application/agents/injury"
	"github.com/trace/trace/internal/application/agents/interview"
	"github.com/trace/trace/internal/application/agents/logagent"
	"github.com/trace/trace/internal/application/agents/metabolic"
	moegate "github.com/trace/trace/internal/application/agents/moe_gate"
	"github.com/trace/trace/internal/application/agents/planner"
	"github.com/trace/trace/internal/application/agents/progression"
	"github.com/trace/trace/internal/application/agents/recovery"
	"github.com/trace/trace/internal/application/agents/twin"
	"github.com/trace/trace/internal/application/orchestrator"
	deliveryhttp "github.com/trace/trace/internal/delivery/http"
	"github.com/trace/trace/internal/delivery/http/handlers"
	"github.com/trace/trace/internal/delivery/sse"
	"github.com/trace/trace/internal/infrastructure/groq"
	"github.com/trace/trace/internal/infrastructure/mcp"
	"github.com/trace/trace/internal/infrastructure/messaging"
	redisinfra "github.com/trace/trace/internal/infrastructure/redis"
	"github.com/trace/trace/pkg/auth"
	"github.com/trace/trace/pkg/crypto"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// ── Infrastructure ────────────────────────────────────────────────────────

	redisClient, err := redisinfra.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("redis connect: %v", err)
	}
	defer redisClient.Close()

	cipher, err := crypto.NewAESCipher(cfg.Crypto.AESKey)
	if err != nil {
		log.Fatalf("aes cipher: %v", err)
	}

	llmClient := groq.NewClient(cfg.Groq)

	// ── Repositories ─────────────────────────────────────────────────────────

	graphRepo := redisinfra.NewGraphRepository(redisClient)
	userRepo := redisinfra.NewUserRepository(redisClient, cipher)
	workoutLogRepo := redisinfra.NewWorkoutLogRepository(redisClient)
	workoutPlanRepo := redisinfra.NewWorkoutPlanRepository(redisClient)
	progressionRepo := redisinfra.NewProgressionHistoryRepository(redisClient)
	cotRepo := redisinfra.NewCoTTraceRepository(redisClient)
	vetoRepo := redisinfra.NewVetoEventRepository(redisClient)

	// ── MCP Server (the only data access path for agents) ────────────────────

	mcpServer := mcp.NewServer(
		graphRepo,
		userRepo,
		workoutLogRepo,
		workoutPlanRepo,
		nil, // exercise library repo — to be wired (static data source)
		progressionRepo,
		vetoRepo,
		cotRepo,
		nil, // conflict repo — to be wired
	)

	// ── Message Broker ───────────────────────────────────────────────────────

	broker := messaging.NewBroker(64)

	// ── SSE Broker ───────────────────────────────────────────────────────────

	sseBroker := sse.NewBroker()

	// ── Agents ───────────────────────────────────────────────────────────────
	// Each agent receives ONLY the ports it needs — never raw repositories.

	interviewAgent := interview.New(mcpServer, llmClient)
	twinAgent := twin.New(mcpServer, llmClient)
	moeGateAgent := moegate.New(mcpServer, llmClient)
	recoveryAgent := recovery.New(mcpServer)
	injuryAgent := injury.New(mcpServer)
	growthAgent := growth.New(mcpServer)
	metabolicAgent := metabolic.New(mcpServer)
	progressionAgent := progression.New(mcpServer)
	plannerAgent := planner.New(mcpServer)
	coachAgent := coach.New(mcpServer, llmClient)
	logAgent := logagent.New(mcpServer, llmClient)

	// ── Orchestrators ────────────────────────────────────────────────────────

	dailyLoop := orchestrator.NewDailyLoopOrchestrator(
		logAgent,
		twinAgent,
		moeGateAgent,
		recoveryAgent,
		injuryAgent,
		growthAgent,
		metabolicAgent,
		progressionAgent,
		plannerAgent,
		coachAgent,
		mcpServer,
		broker,
		sseBroker,
	)

	onboarding := orchestrator.NewOnboardingOrchestrator(
		interviewAgent,
		twinAgent,
		moeGateAgent,
		mcpServer,
		dailyLoop,
	)

	// Merge all state handlers into one machine
	allHandlers := make(map[orchestrator.State]orchestrator.StateHandler)
	for k, v := range onboarding.Handlers() {
		allHandlers[k] = v
	}
	for k, v := range dailyLoop.Handlers() {
		allHandlers[k] = v
	}
	_ = orchestrator.NewMachine(allHandlers)

	// ── HTTP Handlers ─────────────────────────────────────────────────────────

	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)

	authHandler := handlers.NewAuthHandler(jwtSvc)
	interviewHandler := handlers.NewInterviewHandler(interviewAgent, twinAgent)
	logHandler := handlers.NewWorkoutLogHandler(logAgent)
	workoutHandler := handlers.NewWorkoutHandler(mcpServer, sseBroker)
	graphHandler := handlers.NewGraphHandler(mcpServer)

	// Rate limit: 100 API calls per user per hour, tracked in Redis.
	rateLimitFn := func(userID string) (bool, error) {
		key := redisinfra.RateLimitKey(userID)
		count, err := redisClient.RDB().Incr(ctx, key).Result()
		if err != nil {
			return true, nil // allow on Redis error
		}
		if count == 1 {
			redisClient.RDB().Expire(ctx, key, time.Hour)
		}
		return count <= 100, nil
	}

	router := deliveryhttp.NewRouter(
		jwtSvc,
		rateLimitFn,
		authHandler,
		interviewHandler,
		logHandler,
		workoutHandler,
		graphHandler,
	)

	// ── HTTP Server ───────────────────────────────────────────────────────────

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("TRACE server listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// ── Graceful Shutdown ─────────────────────────────────────────────────────

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	log.Println("server stopped")

	_ = moeGateAgent
	_ = broker
}
