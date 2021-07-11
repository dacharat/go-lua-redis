package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dacharat/go-lua-redis/config"
	redisClient "github.com/go-redis/redis/v8"
)

var (
	RedisClient     *redisClient.Client
	Script          *redisClient.Script
	SortOrderScript *redisClient.Script
)

func NewRedis() {
	fmt.Println("Redis Host: ", config.RedisHost)
	RedisClient = redisClient.NewClient(&redisClient.Options{
		Addr:     config.RedisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := RedisClient.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}

	Script = createScript()
	SortOrderScript = createSortOrderScript()
}

func createScript() *redisClient.Script {
	return redisClient.NewScript(`
		local goodsSurplus
		local flag
		local existUserIds    = tostring(KEYS[1])
		local memberUid       = tonumber(ARGV[1])
		local goodsSurplusKey = tostring(KEYS[2])
		local hasBuy = redis.call("sIsMember", existUserIds, memberUid)

		if hasBuy ~= 0 then
		  return 0
		end
		

		goodsSurplus =  redis.call("GET", goodsSurplusKey)
		if goodsSurplus == false then
		  return 0
		end
		
		- There are no remaining items to grab
		goodsSurplus = tonumber(goodsSurplus)
		if goodsSurplus <= 0 then
		  return 0
		end
		
		flag = redis.call("SADD", existUserIds, memberUid)
		flag = redis.call("DECR", goodsSurplusKey)
		
		return 1
	`)
}

func evalScript(client *redisClient.Client, userId string, wg *sync.WaitGroup) {
	defer wg.Done()
	script := createScript()
	sha, err := script.Load(client.Context(), client).Result()
	if err != nil {
		log.Fatalln(err)
	}
	ret := client.EvalSha(client.Context(), sha, []string{
		"hadBuyUids",
		"goodsSurplus",
	}, userId)
	if result, err := ret.Result(); err != nil {
		log.Fatalf("Execute Redis fail: %v", err.Error())
	} else {
		fmt.Println("")
		fmt.Printf("userid: %s, result: %d", userId, result)
	}
}

func createSortOrderScript() *redisClient.Script {
	return redisClient.NewScript(`
	if tonumber(redis.call('ZADD', KEYS[1], ARGV[1], ARGV[2])) == 1 then
		return redis.call('INCR', KEYS[2])  
	else 
		return redis.call('GET', KEYS[2]) 
	end
	`)
}

func EvalSortScript(userId string, value []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	sha, err := SortOrderScript.Load(RedisClient.Context(), RedisClient).Result()
	if err != nil {
		log.Fatalln(err)
	}
	orderByTimeKey := fmt.Sprintf("{User:%s}:OrderByTime", userId)
	orderCountKey := fmt.Sprintf("{User:%s}:OrderCount", userId)
	ret := RedisClient.EvalSha(RedisClient.Context(), sha, []string{
		orderByTimeKey,
		orderCountKey,
	}, time.Now().Nanosecond(), value)
	if result, err := ret.Result(); err != nil {
		log.Fatalf("Execute Redis fail: %v", err.Error())
	} else {
		fmt.Println("")
		fmt.Printf("userid: %s, result: %d", userId, result)
	}
}
