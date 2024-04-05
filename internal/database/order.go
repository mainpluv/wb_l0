package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mainpluv/wb_l0/internal/model"
)

type OrderRepository interface {
	Create(model.Order) (*model.Order, error)
	GetAll() ([]model.Order, error)
	GetOne(uuid.UUID) (*model.Order, error)
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

	// в0ставка данных в табл delivery
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

	// вставка данных в табл payment
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

	// вставка данных в табл orders
	orderQuery := `
		INSERT INTO orders (order_uuid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shard_key, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err = tx.Exec(context.Background(), orderQuery,
		order.OrderUUID, order.TrackNumber, order.Entry, deliveryID, paymentID,
		order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return nil, err
	}

	// вставка данных в табл items
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
		// вставка данных в табл orders_items
		_, err = tx.Exec(context.Background(), "INSERT INTO orders_items (order_uuid, item_id) VALUES ($1, $2)", order.OrderUUID, itemID)
		if err != nil {
			return nil, err
		}
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

	// запрос для получения всех данных о заказах с информацией о доставке и оплате
	query := `
	SELECT 
	o.order_uuid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
	o.delivery_service, o.shard_key, o.sm_id, o.date_created, o.oof_shard, 
	d.id, d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
	p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
	p.bank, p.delivery_cost, p.goods_total, p.custom_fee, p.id
	FROM 
	orders o
	JOIN delivery d ON o.delivery_id = d.id
	JOIN payment p ON o.payment_id = p.id`

	rows, err := tx.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	rows.Close()

	for rows.Next() {
		var order model.Order
		var delivery model.Delivery
		var payment model.Payment

		// сканирование данных заказа, доставки и оплаты из строк результата запроса
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

		// установка значений доставки и оплаты для заказа
		order.Delivery = delivery
		order.Payment = payment

		// запрос для получения товаров в заказе
		iQuery := `
			SELECT i.id, i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, 
			i.size, i.total_price, i.nm_id, i.brand, i.status
			FROM items i
			INNER JOIN orders_items oi ON i.id = oi.item_id
			WHERE oi.order_uuid = $1`

		rowsItems, err := tx.Query(context.Background(), iQuery, order.OrderUUID)
		if err != nil {
			return nil, err
		}
		defer rowsItems.Close()

		var items []model.Item
		for rowsItems.Next() {
			var item model.Item
			err := rowsItems.Scan(
				&item.Id, &item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
				&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID,
				&item.Brand, &item.Status,
			)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}

		// установка товаров для текущего заказа
		order.Items = items

		// добавление текущего заказа в список заказов
		orders = append(orders, order)
	}

	return orders, nil
}
func (o *OrderRepo) GetOne(uuid uuid.UUID) (*model.Order, error) {
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

	var order model.Order
	var delivery model.Delivery
	var payment model.Payment

	// запрос для получения данных о заказе с указанным uuid
	query := `
    SELECT 
    o.order_uuid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
    o.delivery_service, o.shard_key, o.sm_id, o.date_created, o.oof_shard, 
    d.id, d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
    p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
    p.bank, p.delivery_cost, p.goods_total, p.custom_fee, p.id
    FROM 
    orders o
    JOIN delivery d ON o.delivery_id = d.id
    JOIN payment p ON o.payment_id = p.id
    WHERE o.order_uuid = $1`

	row := tx.QueryRow(context.Background(), query, uuid)
	err = row.Scan(
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

	// установка значений доставки и оплаты для заказа
	order.Delivery = delivery
	order.Payment = payment

	// запрос для получения товаров в заказе
	iQuery := `
    SELECT i.id, i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, 
    i.size, i.total_price, i.nm_id, i.brand, i.status
    FROM items i
    INNER JOIN orders_items oi ON i.id = oi.item_id
    WHERE oi.order_uuid = $1`

	rowsItems, err := tx.Query(context.Background(), iQuery, uuid)
	if err != nil {
		return nil, err
	}
	defer rowsItems.Close()

	var items []model.Item
	for rowsItems.Next() {
		var item model.Item
		err := rowsItems.Scan(
			&item.Id, &item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID,
			&item.Brand, &item.Status,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	// установка товаров для текущего заказа
	order.Items = items

	return &order, nil
}
