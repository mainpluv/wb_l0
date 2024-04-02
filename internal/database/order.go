package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mainpluv/wb_l0/internal/model"
)

type OrderRepository interface {
	Create(model.Order) (*model.Order, error)
	GetAll() ([]model.Order, error)
}

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{
		pool: pool,
	}
}

func (o *OrderRepo) Create(order model.Order) (*model.Order, error) {
	tx, err := o.pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	// Вставка данных в таблицу delivery
	deliveryQuery := `
		INSERT INTO delivery (name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	var deliveryID int
	err = tx.QueryRow(context.Background(), deliveryQuery,
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region,
		order.Delivery.Email).Scan(&deliveryID)
	if err != nil {
		return nil, err
	}

	// Вставка данных в таблицу payment
	paymentQuery := `
		INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`
	var paymentID int
	err = tx.QueryRow(context.Background(), paymentQuery,
		order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt,
		order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal,
		order.Payment.CustomFee).Scan(&paymentID)
	if err != nil {
		return nil, err
	}

	// Вставка данных в таблицу items
	for _, item := range order.Items {
		itemQuery := `
			INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id`
		var itemID int
		err := tx.QueryRow(context.Background(), itemQuery,
			item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand,
			item.Status).Scan(&itemID)
		if err != nil {
			return nil, err
		}
		// Вставка данных в таблицу orders_items
		_, err = tx.Exec(context.Background(), "INSERT INTO orders_items (order_uuid, item_id) VALUES ($1, $2)", order.OrderUUID, itemID)
		if err != nil {
			return nil, err
		}
	}

	// Вставка данных в таблицу orders
	orderQuery := `
		INSERT INTO orders (track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err = tx.Exec(context.Background(), orderQuery,
		order.TrackNumber, order.Entry, deliveryID, paymentID,
		order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *OrderRepo) GetAll() ([]model.Order, error) {
	tx, err := o.pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	var orders []model.Order

	// Запрос для получения всех данных о заказах с информацией о доставке и оплате
	query := `
		SELECT *
		FROM orders o
		JOIN delivery d ON o.delivery_id = d.id
		JOIN payment p ON o.payment_id = p.id`

	rows, err := tx.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		var delivery model.Delivery
		var payment model.Payment

		// Сканирование данных заказа, доставки и оплаты из строк результата запроса
		err := rows.Scan(
			&order.OrderUUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated,
			&order.OofShard, &delivery.Id, &delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
			&delivery.Address, &delivery.Region, &delivery.Email, &payment.Transaction, &payment.RequestID,
			&payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank,
			&payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee, &payment.Id,
		)
		if err != nil {
			return nil, err
		}

		// Установка значений доставки и оплаты для заказа
		order.Delivery = delivery
		order.Payment = payment

		// Запрос для получения товаров в заказе
		itemsQuery := `
			SELECT i.id, i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, 
			i.size, i.total_price, i.nm_id, i.brand, i.status
			FROM items i
			INNER JOIN orders_items oi ON i.id = oi.item_id
			WHERE oi.order_uuid = $1`

		rows, err := tx.Query(context.Background(), itemsQuery, order.OrderUUID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []model.Item
		for rows.Next() {
			var item model.Item
			err := rows.Scan(
				&item.Id, &item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
				&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID,
				&item.Brand, &item.Status,
			)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}

		// Установка товаров для текущего заказа
		order.Items = items

		// Добавление текущего заказа в список заказов
		orders = append(orders, order)
	}

	return orders, nil
}
