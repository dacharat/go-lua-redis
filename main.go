package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/dacharat/go-lua-redis/config"
	"github.com/dacharat/go-lua-redis/redis"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Order struct {
	OrderID string `json:"orderId"`
}

func main() {
	fmt.Println("Start Server!!")
	config.SetConfig()
	redis.NewRedis()
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// user := User{Name: "jack", Age: 20}
	// value, err := json.Marshal(user)
	// if err != nil {
	// 	panic(err)
	// }

	// err = redis.RedisClient.Set(ctx, "name", value, 5*time.Minute).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// result, err := redis.RedisClient.Get(ctx, "name").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(result)

	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(100000)
	order := Order{OrderID: strconv.Itoa(id)}

	value, err := json.Marshal(order)
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	redis.EvalSortScript("1234", value, &wg)

	results, err := redis.RedisClient.Keys(ctx, "*").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(results)

	orders, err := redis.RedisClient.ZRange(ctx, "{User:1234}:OrderByTime", 0, -1).Result()
	// orders, err := redis.RedisClient.ZRangeWithScores(ctx, "{User:1234}:OrderByTime", 0, -1).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(orders)

	orders2, err := redis.RedisClient.ZRevRange(ctx, "{User:1234}:OrderByTime", 0, -1).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(orders2)
}
