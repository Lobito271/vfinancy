package shipment

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/internal/domain/enums"
	"vfinancy/backend/internal/domain/repositories"
)

type fakeTx struct{}

func (fakeTx) WithinTransaction(ctx context.Context, fn repositories.TxRunner) error { return fn(ctx) }

type fakeRepo struct {
	byID map[uuid.UUID]*Shipment
	seq  int
}

func (r *fakeRepo) Create(_ context.Context, s *Shipment) error {
	r.byID[s.ID] = s
	return nil
}

func (r *fakeRepo) Update(_ context.Context, s *Shipment) error {
	if _, ok := r.byID[s.ID]; !ok {
		return repositories.ErrNotFound
	}
	r.byID[s.ID] = s
	return nil
}

func (r *fakeRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.byID[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(r.byID, id)
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id uuid.UUID) (*Shipment, error) {
	s, ok := r.byID[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return s, nil
}

func (r *fakeRepo) NextCode(_ context.Context) (string, error) {
	r.seq++
	return code4(r.seq), nil
}

func (r *fakeRepo) List(context.Context, ShipmentFilter) (repositories.Page[*Shipment], error) {
	return repositories.Page[*Shipment]{}, nil
}

func code4(n int) string {
	s := []byte("0000")
	for i := 3; n > 0 && i >= 0; i-- {
		s[i] = byte('0' + n%10)
		n /= 10
	}
	return string(s)
}

func TestCreateAssignsSequentialCode(t *testing.T) {
	svc := New(&fakeRepo{byID: map[uuid.UUID]*Shipment{}}, fakeTx{}, logger.New("error", "text", ""))
	a, err := svc.Create(context.Background(), CreateInput{Description: "Primer envío"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	b, err := svc.Create(context.Background(), CreateInput{Description: "Segundo envío"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if a.Code != "0001" || b.Code != "0002" {
		t.Fatalf("codes = %s, %s; want 0001, 0002", a.Code, b.Code)
	}
	if a.Status != enums.ShipmentStatusPending {
		t.Fatalf("status = %s, want pending", a.Status)
	}
	if err := a.Validate(true); err != nil {
		t.Fatalf("created shipment must validate: %v", err)
	}
}

func TestValidateRejectsMalformedCode(t *testing.T) {
	s := &Shipment{ID: uuid.New(), Code: "12", Status: enums.ShipmentStatusPending}
	if err := s.Validate(true); err == nil {
		t.Fatal("code = 12 must be rejected")
	}
	s.Code = "ABCD"
	if err := s.Validate(true); err == nil {
		t.Fatal("code = ABCD must be rejected")
	}
	s.Code = "0001"
	if err := s.Validate(true); err != nil {
		t.Fatalf("code = 0001 must pass: %v", err)
	}
}

func TestUpdateKeepsCodeAndChangesStatus(t *testing.T) {
	repo := &fakeRepo{byID: map[uuid.UUID]*Shipment{}}
	svc := New(repo, fakeTx{}, logger.New("error", "text", ""))
	created, err := svc.Create(context.Background(), CreateInput{Description: "Envío"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	status := enums.ShipmentStatusShipped
	upd, err := svc.Update(context.Background(), UpdateInput{ID: created.ID, Status: &status})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upd.Code != created.Code {
		t.Fatalf("code changed from %s to %s", created.Code, upd.Code)
	}
	if upd.Status != enums.ShipmentStatusShipped {
		t.Fatalf("status = %s, want shipped", upd.Status)
	}
	if upd.Description != "Envío" {
		t.Fatalf("description lost: %q", upd.Description)
	}
}
