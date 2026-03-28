package cache

import "fmt"

func StockKey(goodsID int64) string {
	return fmt.Sprintf("sk:stock:%d", goodsID)
}

func SoldOutKey(goodsID int64) string {
	return fmt.Sprintf("sk:sold_out:%d", goodsID)
}

func UsersKey(goodsID int64) string {
	return fmt.Sprintf("sk:users:%d", goodsID)
}

func IdempotentTokenKey(userID, goodsID int64) string {
	return fmt.Sprintf("sk:token:%d:%d", userID, goodsID)
}

func ResultKey(userID, goodsID int64) string {
	return fmt.Sprintf("sk:result:%d:%d", userID, goodsID)
}

func URLTokenKey(userID, goodsID int64) string {
	return fmt.Sprintf("sk:url_token:%d:%d", userID, goodsID)
}
