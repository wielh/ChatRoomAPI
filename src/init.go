package src

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"sync"

	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type config struct {
	Server struct {
		Host              string `yaml:"host"`
		Port              int    `yaml:"port"`
		SessionEncryptKey string `yaml:"session_encrypt_key"`
	} `yaml:"server"`
	Database struct {
		Host       string `yaml:"host"`
		DBUser     string `yaml:"user"`
		DBPassword string `yaml:"password"`
		DBName     string `yaml:"name"`
		Port       int32  `yaml:"port"`
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
}

type allConfigs struct {
	YamlConfig config
	DB         *gorm.DB
	Redis      *redis.Client
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
		err = GlobalConfig.yamlInit()
		if err != nil {
			return
		}

		err = GlobalConfig.postgreInit()
		if err != nil {
			return
		}

		err = GlobalConfig.redisInit()
		if err != nil {
			return
		}
	})
	return err
}

func (a *allConfigs) yamlInit() error {
	file, err := os.Open("src/config.yaml")
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
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", d.Host, d.DBUser, d.DBPassword, d.DBName, d.Port) // pg only
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
	return nil
}

func (a *allConfigs) NewTransection() *gorm.DB {
	return GlobalConfig.DB.Begin()
}
