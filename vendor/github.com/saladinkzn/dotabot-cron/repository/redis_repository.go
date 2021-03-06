package repository

import (
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

type RedisRepository struct {
	client *redis.Client
}

func CreateRedisRepository(client *redis.Client) SubscriptionRepository {
	return RedisRepository{
		client: client,
	}
}

func parseDotaAccountId(k string) string {
	return k[len("lastMatches.[") : len(k)-1]
}

func (this RedisRepository) GetLastKnownMatchId(subscription TelegramMatchSubscription) (result int64, err error) {
	// Supporting spring-data-redis style
	hash, err := this.client.HGetAll(makeTelegramChatKey(subscription.ChatId)).Result()
	if err != nil {
		return
	}

	for k, v := range hash {
		if !strings.HasPrefix(k, "lastMatches.[") {
			continue
		}

		substr := parseDotaAccountId(k)
		if subscription.DotaAccountId != substr {
			continue
		}

		result, err = strconv.ParseInt(v, 10, 64)
		return
	}

	result = -1
	return
}

func makeTelegramChatKey(chatId int64) string {
	return "telegramMatchSubscription:" + strconv.FormatInt(chatId, 10)
}

func makeAccountIdKey(dotaAccountId string) string {
	return "lastMatches.[" + dotaAccountId + "]"
}

func (this RedisRepository) SaveLastKnownMatchId(subscription TelegramMatchSubscription, matchId uint64) error {
	telegramChatKey := makeTelegramChatKey(subscription.ChatId)

	_, err := this.client.HSet(telegramChatKey, makeAccountIdKey(subscription.DotaAccountId), strconv.FormatUint(matchId, 10)).Result()
	return err
}

func (this RedisRepository) RemoveLastKnownMatchId(subscription TelegramMatchSubscription) error {
	telegramChatKey := makeTelegramChatKey(subscription.ChatId)

	_, err := this.client.HDel(telegramChatKey, makeAccountIdKey(subscription.DotaAccountId)).Result()
	return err
}

func (this RedisRepository) FindAll() (result []TelegramMatchSubscription, err error) {
	chatIds, err := this.client.Keys("telegramMatchSubscription:*").Result()
	if err != nil {
		return
	}

	for _, chatIdKey := range chatIds {
		chatId, err := strconv.ParseInt(chatIdKey[len("telegramMatchSubscription:"):len(chatIdKey)], 10, 64)
		if err != nil {
			return nil, err
		}

		hkeys, err := this.client.HKeys(chatIdKey).Result()
		if err != nil {
			return nil, err
		}

		for _, hkey := range hkeys {
			if !strings.HasPrefix(hkey, "lastMatches.") {
				continue
			}

			dotaAccountId := parseDotaAccountId(hkey)
			subscription := TelegramMatchSubscription{ChatId: chatId, DotaAccountId: dotaAccountId}
			result = append(result, subscription)
		}
	}

	return
}
