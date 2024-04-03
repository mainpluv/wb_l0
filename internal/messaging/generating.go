package messaging

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/mainpluv/wb_l0/internal/model"
)

func GenerateOrder() model.Order {
	tracknb := "WBILM" + genStr(7)
	order := model.Order{
		OrderUUID:         uuid.UUID{},
		TrackNumber:       tracknb,
		Entry:             "WBIL",
		Delivery:          genDelivery(),
		Payment:           genPayment(),
		Items:             genItems(tracknb),
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        genStr(4),
		DeliveryService:   "meest",
		ShardKey:          genStr(1),
		SmID:              genInt(1, 99),
		DateCreated:       time.Now(),
		OofShard:          genStr(1),
	}
	return order
}

func genStr(n int) string {
	arr := make([]byte, n)
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	source := rand.NewSource(time.Now().UnixNano()) // создаем источник случайных чисел
	r := rand.New(source)
	for i := range arr {
		arr[i] = letters[r.Intn(len(letters))]
	}
	return string(arr)
}

func genInt(a, b int) int {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return r.Intn(a-b) + b
}
func genNumbersStr(n int) string {
	var res string
	for i := 0; i < n; i++ {
		res += strconv.Itoa(genInt(0, 9))
	}
	return res
}

func genDelivery() model.Delivery {
	delivery := model.Delivery{
		Id:      genInt(1, 100),
		Name:    genStr(5) + " " + genStr(6),
		Phone:   "+" + genNumbersStr(11),
		Zip:     genNumbersStr(7),
		City:    genStr(10),
		Address: genStr(10) + " " + genNumbersStr(3),
		Region:  genStr(7),
		Email:   genStr(6) + "@gmail.com",
	}
	return delivery
}

func genPayment() model.Payment {
	amount := genInt(10, 100000)
	deliveryCost := genInt(0, amount-1)
	goodsTotal := amount - deliveryCost
	payment := model.Payment{
		Transaction:  uuid.New(),
		RequestID:    "",
		Currency:     genCurrency(),
		Provider:     "wbpay",
		Amount:       amount,
		PaymentDt:    time.Now().UnixNano(),
		Bank:         genBank(),
		DeliveryCost: deliveryCost,
		GoodsTotal:   goodsTotal,
		CustomFee:    0,
	}
	return payment
}

func genCurrency() string {
	var currencies = []string{"RUB", "USD", "EUR"}
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return currencies[r.Intn(len(currencies))]
}

func genBank() string {
	var banks = []string{"sber", "tinkoff", "alpha"}
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return banks[r.Intn(len(banks))]
}

func genItems(tracknb string) []model.Item {
	var items = []model.Item{}
	for i := 0; i < genInt(1, 3); i++ {
		price := genInt(1, 100000)
		sale := genInt(0, 99)
		totalprice := price - price*sale/100
		item := model.Item{
			ChrtID:      genInt(1000000, 9999999),
			TrackNumber: tracknb,
			Price:       price,
			Rid:         genRID(),
			Name:        genStr(8),
			Sale:        genInt(0, 99),
			Size:        strconv.Itoa(genInt(0, 5)),
			TotalPrice:  totalprice,
			NmID:        genInt(1000000, 9999999),
			Brand:       genStr(6),
			Status:      genInt(100, 200),
		}
		items = append(items, item)
	}
	return items
}

func genRID() string {
	arr := make([]byte, 21)
	const letters = "abcdefghijklmnopqrstuvwxyz1234567890"
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	for i := range arr {
		arr[i] = letters[r.Intn(len(letters))]
	}
	return string(arr)
}
