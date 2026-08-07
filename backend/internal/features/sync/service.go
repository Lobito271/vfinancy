package sync

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"vfinancy/backend/internal/domain/repositories"
)

// Service drives the local side of replication: it pushes local changes
// to the sync server and pulls remote changes back, resolving conflicts
// with last-writer-wins and recording every conflict in sync_conflicts.
type Service struct {
	repo     Repository
	client   *HTTPClient
	log      *slog.Logger
	name     string
	platform string
}

// NewService returns a worker bound to a repository and a server client.
func NewService(repo Repository, client *HTTPClient, log *slog.Logger, name, platform string) *Service {
	return &Service{repo: repo, client: client, log: log, name: name, platform: platform}
}

// RunOnce performs one full replication exchange. It is safe to call
// repeatedly from a background loop: it registers the device on first
// contact, pushes pending changes, applies the server response with LWW,
// advances the per-table cursors and prunes consumed tombstones.
func (s *Service) RunOnce(ctx context.Context) error {
	device, err := s.ensureDevice(ctx)
	if err != nil {
		return err
	}

	cursors, err := s.repo.GetCursors(ctx, device.ID)
	if err != nil {
		return err
	}

	req := &Request{Cursors: cursors}
	for _, meta := range SyncedTables {
		cur := cursors[meta.Name]
		rows, _, err := s.repo.RowsChangedSince(ctx, meta, cur)
		if err != nil {
			return err
		}
		req.Rows = append(req.Rows, rows...)
		tombs, err := s.repo.TombstonesSince(ctx, meta.Name, cur)
		if err != nil {
			return err
		}
		req.Tombstones = append(req.Tombstones, tombs...)
	}

	resp, err := s.client.Sync(ctx, device.Token, req)
	if err != nil {
		return err
	}

	conflicts := s.applyResponse(ctx, device.ID, resp)

	if err := s.repo.UpdateCursors(ctx, device.ID, resp.Cursors); err != nil {
		return err
	}
	for table, wm := range resp.Cursors {
		if _, err := s.repo.PurgeTombstones(ctx, table, wm); err != nil {
			s.log.Warn("sync: tombstone purge failed", "table", table, "error", err)
		}
	}
	if resp.ServerTime > 0 {
		if err := s.repo.TouchDeviceSeen(ctx, device.ID, resp.ServerTime); err != nil {
			s.log.Warn("sync: touch device failed", "error", err)
		}
	}
	for _, c := range conflicts {
		if err := s.repo.LogConflict(ctx, c); err != nil {
			s.log.Warn("sync: persist conflict failed", "table", c.TableName, "record", c.RecordID, "error", err)
		}
	}
	return nil
}

// applyResponse applies the rows and tombstones returned by the server
// with LWW and collects the conflicts that were resolved.
func (s *Service) applyResponse(ctx context.Context, deviceID string, resp *Response) []*Conflict {
	out := make([]*Conflict, 0)

	for _, ch := range resp.Rows {
		meta := LookupTable(ch.TableName)
		if meta == nil {
			s.log.Warn("sync: ignoring unknown table", "table", ch.TableName)
			continue
		}
		applied, localTime, err := s.repo.ApplyRow(ctx, meta, ch.Payload, ch.UpdatedAt)
		if err != nil {
			s.log.Warn("sync: apply row failed", "table", ch.TableName, "record", ch.RecordID, "error", err)
			continue
		}
		if !applied {
			out = append(out, newConflict(deviceID, ch.TableName, ch.RecordID, "UPDATE",
				msToTime(localTime), msToTime(ch.UpdatedAt), ResolutionLocalWon, "local row is newer"))
		}
	}

	for _, tb := range resp.Tombstones {
		meta := LookupTable(tb.TableName)
		if meta == nil {
			continue
		}
		applied, localTime, err := s.repo.ApplyTombstone(ctx, meta, tb.RecordID, tb.UpdatedAt)
		if err != nil {
			s.log.Warn("sync: apply tombstone failed", "table", tb.TableName, "record", tb.RecordID, "error", err)
			continue
		}
		if !applied {
			out = append(out, newConflict(deviceID, tb.TableName, tb.RecordID, "DELETE",
				msToTime(localTime), msToTime(tb.UpdatedAt), ResolutionLocalWon, "local row is newer than the remote delete"))
		}
	}

	for _, r := range resp.Results {
		if r.Status != StatusConflict {
			continue
		}
		msg := r.Message
		if msg == "" {
			msg = "server kept the newer version"
		}
		out = append(out, newConflict(deviceID, r.TableName, r.RecordID, "UPDATE",
			nil, nil, ResolutionRemoteWon, msg))
	}

	return out
}

// ensureDevice returns the local device, registering it with the server
// on first contact. Until registration succeeds the device id is
// unknown, so no cursors exist yet; local writes still proceed because
// row replication does not depend on the device row.
func (s *Service) ensureDevice(ctx context.Context) (*Device, error) {
	d, err := s.repo.GetLocalDevice(ctx)
	if err == nil {
		return d, nil
	}
	if !errors.Is(err, repositories.ErrNotFound) {
		return nil, err
	}
	companyID, err := s.repo.FirstCompanyID(ctx)
	if err != nil {
		return nil, err
	}
	reg, err := s.client.Register(ctx, &RegisterRequest{
		CompanyID: companyID,
		Name:      s.name,
		Platform:  s.platform,
	})
	if err != nil {
		return nil, err
	}
	d = &Device{
		ID:        reg.DeviceID,
		CompanyID: companyID,
		Name:      s.name,
		Platform:  s.platform,
		Token:     reg.Token,
		IsLocal:   true,
		IsActive:  true,
	}
	if err := s.repo.RegisterDevice(ctx, d); err != nil {
		return nil, err
	}
	s.log.Info("sync: device registered", "device_id", d.ID)
	return d, nil
}

func newConflict(deviceID, table, recordID, operation string, local, remote *time.Time, resolution, message string) *Conflict {
	return &Conflict{
		ID:              newID(),
		DeviceID:        strPtr(deviceID),
		TableName:       table,
		RecordID:        recordID,
		Operation:       operation,
		LocalUpdatedAt:  local,
		RemoteUpdatedAt: remote,
		Resolution:      resolution,
		Message:         message,
		CreatedAt:       time.Now().UTC(),
	}
}

func msToTime(ms int64) *time.Time {
	if ms == 0 {
		return nil
	}
	t := time.UnixMilli(ms).UTC()
	return &t
}

func strPtr(s string) *string {
	return &s
}
