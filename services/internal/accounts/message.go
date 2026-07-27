package accounts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Message is a lead/contact submission from a store's public "contact" section. Unlike a Customer, a
// message has no login — anyone who fills out the storefront form can leave one, scoped to whichever
// store received it.
type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID  string             `bson:"tenantId" json:"-"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Body      string             `bson:"body" json:"body"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// maxMessagesListed caps how many submissions the dashboard fetches at once — a merchant inbox, not an
// unbounded export. messageBodyMax caps a submission's free text, same order of magnitude as a section
// field (a contact message is short prose, not a policy page).
const (
	maxMessagesListed = 200
	messageBodyMax    = 4000
)

// --- repository ---

func (r *Repository) insertMessage(ctx context.Context, m *Message) error {
	_, err := r.messages.InsertOne(ctx, m)
	return err
}

// messagesByTenant returns a store's submissions, newest first, capped at maxMessagesListed.
func (r *Repository) messagesByTenant(ctx context.Context, tenantID string) ([]Message, error) {
	cur, err := r.messages.Find(ctx, bson.M{"tenantId": tenantID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(maxMessagesListed))
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	var out []Message
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- service ---

// newMessage trims/caps and validates a contact-form submission, pure so it's unit-testable without a
// DB (same split as sanitizeLayout/sanitizePages elsewhere in this package).
func newMessage(tenantID, name, email, body string) (*Message, error) {
	name = capRunes(strings.TrimSpace(name), 120)
	email = strings.TrimSpace(strings.ToLower(email))
	body = capRunes(strings.TrimSpace(body), messageBodyMax)
	if name == "" || !strings.Contains(email, "@") || body == "" {
		return nil, fmt.Errorf("%w: name, a valid email and a message are required", ErrInvalidInput)
	}
	return &Message{
		ID:        primitive.NewObjectID(),
		TenantID:  tenantID,
		Name:      name,
		Email:     email,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// SubmitMessage records a contact-form submission for a store. tenantID is the store id taken straight
// from the (public) URL path — there's no auth on this endpoint to spoof a header in the first place.
func (s *Service) SubmitMessage(ctx context.Context, tenantID, name, email, body string) error {
	m, err := newMessage(tenantID, name, email, body)
	if err != nil {
		return err
	}
	return s.repo.insertMessage(ctx, m)
}

// Messages returns a store's contact submissions, scoped to its owner — same ownership-check pattern as
// UpdateStore/DeleteStore, so a merchant can't read another store's inbox by guessing an id.
func (s *Service) Messages(ctx context.Context, ownerID, storeID string) ([]Message, error) {
	store, err := s.repo.storeByID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store.OwnerID != ownerID {
		return nil, ErrStoreNotFound // don't reveal existence of stores the caller doesn't own
	}
	msgs, err := s.repo.messagesByTenant(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if msgs == nil {
		msgs = []Message{}
	}
	return msgs, nil
}
