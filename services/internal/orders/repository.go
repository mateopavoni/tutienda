package orders

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrOrderNotFound is returned when an order id has no match.
var ErrOrderNotFound = errors.New("order not found")

// Repository persists orders in MongoDB.
type Repository struct {
	orders *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{orders: db.Collection("orders")}
}

func (r *Repository) insert(ctx context.Context, o *Order) (primitive.ObjectID, error) {
	out, err := r.orders.InsertOne(ctx, o)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return out.InsertedID.(primitive.ObjectID), nil
}

func (r *Repository) setStatus(ctx context.Context, id primitive.ObjectID, status OrderStatus, payment PaymentStatus, reason string) error {
	_, err := r.orders.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status, "paymentStatus": payment, "failureReason": reason}},
	)
	return err
}

// listByCustomer returns a customer's orders within their store, newest first. Scoped by both tenant and
// customer so one shopper can never read another's (or another store's) orders.
func (r *Repository) listByCustomer(ctx context.Context, tenant, customerID string) ([]Order, error) {
	cur, err := r.orders.Find(ctx,
		bson.M{"tenantId": tenant, "customerId": customerID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	var out []Order
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) get(ctx context.Context, tenant string, id primitive.ObjectID) (*Order, error) {
	var o Order
	err := r.orders.FindOne(ctx, bson.M{"_id": id, "tenantId": tenant}).Decode(&o)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}
