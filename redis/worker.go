package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	helper "github.com/mrNobody95/basicGo"
	"github.com/mrNobody95/redisGoWork"
	"log"
	"os"
	"os/signal"
	"time"
)

type Context struct {
	customerID int64
}

var Namespace = helper.GetStringEnv("REDIS_NAMESPACE")

var workerPool = &redis.Pool{
	Dial: func() (redis.Conn, error) {
		return redis.Dial(helper.GetStringEnv("REDIS_CONNECTION"),
			fmt.Sprintf("%s:%s", helper.GetStringEnv("REDIS_HOST"), helper.GetStringEnv("REDIS_PORT")))
	},
	MaxIdle:         helper.GetIntEnv("REDIS_MAX_IDLE_CONNECTIONS"),
	MaxActive:       helper.GetIntEnv("REDIS_MAX_ACTIVE_CONNECTIONS"),
	IdleTimeout:     helper.GetDurationEnv("REDIS_IDLE_TIMEOUT"),
	Wait:            helper.GetBoolEnv("REDIS_WAIT_CONNECTIONS"),
	MaxConnLifetime: helper.GetDurationEnv("REDIS_MAX_LIFETIME"),
}

var PeriodicWorkIds map[uint64]uint64
var WorkNames []string
var counterMap map[string]int

var pool = work.NewWorkerPool(Context{}, 20, Namespace, workerPool)
var Enqueuer = work.NewEnqueuer(Namespace, workerPool)
var started = false

type JobFunc func(job *work.Job) error

func InitRedisWorker() {
	PeriodicWorkIds = make(map[uint64]uint64)
	counterMap = make(map[string]int)
	WorkNames = make([]string, 0)
	pool.Middleware((*Context).Log)
}

func StartRedis() {
	pool.Start()

	started = true

	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan
	fmt.Println("Kill redis")

	// Stop the pool
	drainer := work.Drainer{
		Pool:             pool,
		Namespace:        Namespace,
		WorkerPoolConfig: workerPool,
		PeriodicWorkIds:  PeriodicWorkIds,
		WorkNames:        []string{AutoScheduledCommentCheckLabel},
	}
	drainer.Drain()
}

func DefinePeriodicJob(job PeriodicJob) {
	if !started {
		pool.JobWithOptions(job.Label, work.JobOptions{Priority: job.Priority, MaxFails: job.MaxFails}, job.Job)
		id := pool.PeriodicallyDurationEnqueue(job.Duration, job.Label, job.Args)
		PeriodicWorkIds[id] = id
		if !job.NotDrain {
			WorkNames = append(WorkNames, job.Label)
		}
	} else {
		log.Fatal("periodic job must be defined before redis service started")
	}
}

func DefineTimeBaseJob(job TimeBaseJob) error {
	_, err := Enqueuer.RunAt(job.Name, job.Time, job.Args)
	return err
}

func DefineDurationBaseJob(job DurationBaseJob) error {
	_, err := Enqueuer.EnqueueUniqueIn(job.Name, job.Seconds, job.Args)
	return err
}

type PeriodicJob struct {
	Duration time.Duration
	Args     map[string]interface{}
	Label    string
	Priority uint
	MaxFails uint
	Job      JobFunc
	NotDrain bool
}

type TimeBaseJob struct {
	Time time.Time
	Name string
	Args map[string]interface{}
}

type DurationBaseJob struct {
	Seconds int64
	Name    string
	Args    map[string]interface{}
}
