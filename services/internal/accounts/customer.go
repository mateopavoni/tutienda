package accounts

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
)

// ErrCustomerNotFound is returned when no customer matches an id/email within a store.
var ErrCustomerNotFound = errors.New("customer not found")

// Customer is a storefront shopper of one store. Unlike a Merchant (who runs the SaaS account), a
// customer belongs to a single tenant — the same email can be a separate customer in another store, so
// identity is scoped by {TenantID, Email}. The password is only ever stored hashed.
type Customer struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID     string             `bson:"tenantId" json:"-"`
	Email        string             `bson:"email" json:"email"`
	Name         string             `bson:"name,omitempty" json:"name,omitempty"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}

// --- repository ---

func (r *Repository) insertCustomer(ctx context.Context, c *Customer) error {
	_, err := r.customers.InsertOne(ctx, c)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *Repository) customerByEmail(ctx context.Context, tenantID, email string) (*Customer, error) {
	var c Customer
	err := r.customers.FindOne(ctx, bson.M{"tenantId": tenantID, "email": email}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrCustomerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) customerByID(ctx context.Context, tenantID, id string) (*Customer, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrCustomerNotFound
	}
	var c Customer
	err = r.customers.FindOne(ctx, bson.M{"_id": oid, "tenantId": tenantID}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrCustomerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// --- service ---

// CustomerSignup registers a shopper in one store and returns a customer-scoped (buyer) token bound to
// that tenant. The tenant comes from the gateway-resolved slug, never the client.
func (s *Service) CustomerSignup(ctx context.Context, tenantID, email, password, name string) (string, *Customer, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.Contains(email, "@") || len(password) < 8 {
		return "", nil, fmt.Errorf("%w: email required and password >= 8 chars", ErrInvalidInput)
	}
	hash, err := authx.HashPassword(password)
	if err != nil {
		return "", nil, err
	}
	c := &Customer{
		ID:           primitive.NewObjectID(),
		TenantID:     tenantID,
		Email:        email,
		Name:         capRunes(strings.TrimSpace(name), 120),
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.insertCustomer(ctx, c); err != nil {
		return "", nil, err // ErrDuplicate ⇒ 409 (already registered in this store)
	}
	token, err := s.issuer.IssueCustomer(c.ID.Hex(), tenantID)
	if err != nil {
		return "", nil, err
	}
	return token, c, nil
}

// CustomerLogin verifies a shopper's credentials within a store and returns a buyer token.
func (s *Service) CustomerLogin(ctx context.Context, tenantID, email, password string) (string, *Customer, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	c, err := s.repo.customerByEmail(ctx, tenantID, email)
	if errors.Is(err, ErrCustomerNotFound) {
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}
	if !authx.CheckPassword(c.PasswordHash, password) {
		return "", nil, ErrInvalidCredentials
	}
	token, err := s.issuer.IssueCustomer(c.ID.Hex(), tenantID)
	if err != nil {
		return "", nil, err
	}
	return token, c, nil
}

// Customer returns a shopper's profile, scoped to their store.
func (s *Service) Customer(ctx context.Context, tenantID, id string) (*Customer, error) {
	return s.repo.customerByID(ctx, tenantID, id)
}
