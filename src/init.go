package src

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"sync"

	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type config struct {
	Server struct {
		Host    string `yaml:"host"`
		Port    int    `yaml:"port"`
		Timeout int    `yaml:"timeout_second"`
		Session struct {
			SecretKey string `yaml:"secret_key"`
			Age       int    `yaml:"age_second"`
			HttpOnly  bool   `yaml:"http_only"`
			Secure    bool   `yaml:"secure"`
		} `yaml:"session"`
		RateLimitConfig struct {
			All struct {
				Second     int `yaml:"second"`
				MaxRequest int `yaml:"max_request"`
			} `yaml:"all"`
			IP struct {
				Second     int `yaml:"second"`
				MaxRequest int `yaml:"max_request"`
			} `yaml:"ip"`
			Repeat struct {
				Second     int `yaml:"second"`
				MaxRequest int `yaml:"max_request"`
			} `yaml:"repeat"`
		} `yaml:"rate_limit"`
	} `yaml:"server"`
	Database struct {
		Host       string `yaml:"host"`
		DBUser     string `yaml:"user"`
		DBPassword string `yaml:"password"`
		DBName     string `yaml:"name"`
		Port       int32  `yaml:"port"`
		SSLMode    string `yaml:"sslmode"`
	} `yaml:"database"`
	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`
	Redis struct {
		Address      string `yaml:"address"`
		Password     string `yaml:"password"`
		DBNumber     int    `yaml:"db"`
		PoolSize     int    `yaml:"max_connection"`
		MinIdleConns int    `yaml:"min_connection"`
	} `yaml:"redis"`
	Tempo struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
}

type allConfigs struct {
	YamlConfig   config
	DB           *gorm.DB
	Redis        *redis.Client
	RedisSession redisStore.Store
}

var GlobalConfig allConfigs
var once sync.Once

func init() {
	err := newGlobalConfig()
	if err != nil {
		fmt.Println("error while app init")
		panic(err)
	}
}

func newGlobalConfig() error {
	var err error
	once.Do(func() {
		GlobalConfig = allConfigs{}
		fmt.Println("load yaml file as config ...")
		err = GlobalConfig.yamlInit()
		if err != nil {
			return
		}

		fmt.Println("pg connection init...")
		err = GlobalConfig.postgreInit()
		if err != nil {
			return
		}

		fmt.Println("redis connection init...")
		err = GlobalConfig.redisInit()
		if err != nil {
			return
		}

		fmt.Println("opentelemetry init...")
		err = GlobalConfig.tempoInit()
		if err != nil {
			return
		}

		fmt.Println("Init done")
	})
	return err
}

func (a *allConfigs) yamlInit() error {
	file, err := os.Open("config.yaml")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
		return err
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&a.YamlConfig); err != nil {
		log.Fatalf("Error decoding YAML: %v", err)
		return err
	}
	return err
}

func (a *allConfigs) postgreInit() error {
	d := a.YamlConfig.Database
	sslMode := ""
	if d.SSLMode != "" {
		sslMode += fmt.Sprintf(" sslmode=%s", d.SSLMode)
	}

	dbName := ""
	if d.DBName != "" {
		dbName += fmt.Sprintf(" dbname=%s", d.DBName)
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s %s port=%d %s",
		d.Host, d.DBUser, d.DBPassword, dbName, d.Port, sslMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to the database", dsn, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	a.DB = db
	return nil
}

func (a *allConfigs) redisInit() error {
	r := a.YamlConfig.Redis
	s := a.YamlConfig.Server.Session
	rdb := redis.NewClient(&redis.Options{
		Addr:         r.Address,
		Password:     r.Password,
		DB:           r.DBNumber,
		PoolSize:     r.PoolSize,
		MinIdleConns: r.MinIdleConns,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
		return err
	}
	a.Redis = rdb

	store, err := redisStore.NewStore(r.PoolSize, "tcp", r.Address, r.Password, []byte(s.SecretKey))
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
		return err
	}
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   s.Age,
		HttpOnly: s.HttpOnly,
		Secure:   s.Secure,
	})

	a.RedisSession = store
	return nil
}

func (a *allConfigs) tempoInit() error {
	ctx := context.Background()
	tempoConfig := GlobalConfig.YamlConfig.Tempo
	tempoURL := fmt.Sprintf("%s:%d", tempoConfig.Host, tempoConfig.Port)

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(tempoURL),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
		return err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("gin-practice"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	testTracr := tp.Tracer("test")
	_, span := testTracr.Start(ctx, "init")
	time.Sleep(100 * time.Millisecond) // Simulate some work
	defer span.End()
	return nil
}

func (a *allConfigs) NewTransection() *gorm.DB {
	return GlobalConfig.DB.Begin()
}
