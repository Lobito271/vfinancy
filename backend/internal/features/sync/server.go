package sync

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// Server is the cloud mirror side of replication. It runs inside
// backend/cmd/syncserver backed by PostgreSQL and applies pushed rows
// and tombstones with LWW, then answers the caller with every change
// newer than the caller's cursors.
type Server struct {
	repo Repository
	log  *slog.Logger
}

// NewServer returns a sync server bound to a repository.
func NewServer(repo Repository, log *slog.Logger) *Server {
	return &Server{repo: repo, log: log}
}

// Routes returns the HTTP handler for the sync API, protected by the
// shared API key.
func (s *Server) Routes(apiKey string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(pathRegister, s.handleRegister)
	mux.HandleFunc(pathSync, s.handleSync)
	return withAPIKey(mux, apiKey)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var in RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if in.CompanyID == "" {
		writeError(w, http.StatusBadRequest, "company_id is required")
		return
	}
	device := &Device{
		ID:        newID(),
		CompanyID: in.CompanyID,
		Name:      in.Name,
		Platform:  in.Platform,
		Token:     generateToken(),
		IsActive:  true,
	}
	if device.Name == "" {
		device.Name = "desktop"
	}
	if device.Platform == "" {
		device.Platform = "desktop"
	}
	if err := s.repo.RegisterDevice(r.Context(), device); err != nil {
		s.log.Error("sync: register device failed", "error", err)
		writeError(w, http.StatusInternalServerError, "register failed")
		return
	}
	writeJSON(w, http.StatusCreated, &RegisterResponse{DeviceID: device.ID, Token: device.Token})
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token := r.Header.Get("X-Device-Token")
	if token == "" {
		writeError(w, http.StatusUnauthorized, "device token required")
		return
	}
	device, err := s.repo.GetDeviceByToken(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unknown device token")
		return
	}

	var in Request
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if in.Cursors == nil {
		in.Cursors = map[string]int64{}
	}

	results := make([]Result, 0, len(in.Rows)+len(in.Tombstones))
	conflicts := s.applyPushed(r.Context(), device.ID, &in, &results)

	resp := s.collect(r.Context(), in.Cursors, &results)

	for _, c := range conflicts {
		if err := s.repo.LogConflict(r.Context(), c); err != nil {
			s.log.Warn("sync: persist server conflict failed", "error", err)
		}
	}
	if err := s.repo.TouchDeviceSeen(r.Context(), device.ID, resp.ServerTime); err != nil {
		s.log.Warn("sync: touch device failed", "error", err)
	}

	writeJSON(w, http.StatusOK, resp)
}

// applyPushed applies the caller's rows and tombstones with LWW and
// records any conflicts that resolved against the cloud copy.
func (s *Server) applyPushed(ctx context.Context, deviceID string, in *Request, results *[]Result) []*Conflict {
	conflicts := make([]*Conflict, 0)

	for _, ch := range in.Rows {
		meta := LookupTable(ch.TableName)
		if meta == nil {
			*results = append(*results, Result{TableName: ch.TableName, RecordID: ch.RecordID, Status: StatusFailed, Message: "unknown table"})
			continue
		}
		applied, localTime, err := s.repo.ApplyRow(ctx, meta, ch.Payload, ch.UpdatedAt)
		if err != nil {
			s.log.Warn("sync: server apply row failed", "table", ch.TableName, "record", ch.RecordID, "error", err)
			*results = append(*results, Result{TableName: ch.TableName, RecordID: ch.RecordID, Status: StatusFailed, Message: err.Error()})
			continue
		}
		if !applied {
			*results = append(*results, Result{TableName: ch.TableName, RecordID: ch.RecordID, Status: StatusConflict, Message: "a newer version exists on the server"})
			conflicts = append(conflicts, newConflict(deviceID, ch.TableName, ch.RecordID, "UPDATE",
				msToTime(localTime), msToTime(ch.UpdatedAt), ResolutionLocalWon, "server row is newer; pushed change discarded"))
		} else {
			*results = append(*results, Result{TableName: ch.TableName, RecordID: ch.RecordID, Status: StatusApplied})
		}
	}

	for _, tb := range in.Tombstones {
		meta := LookupTable(tb.TableName)
		if meta == nil {
			*results = append(*results, Result{TableName: tb.TableName, RecordID: tb.RecordID, Status: StatusFailed, Message: "unknown table"})
			continue
		}
		applied, localTime, err := s.repo.ApplyTombstone(ctx, meta, tb.RecordID, tb.UpdatedAt)
		if err != nil {
			s.log.Warn("sync: server apply tombstone failed", "table", tb.TableName, "record", tb.RecordID, "error", err)
			*results = append(*results, Result{TableName: tb.TableName, RecordID: tb.RecordID, Status: StatusFailed, Message: err.Error()})
			continue
		}
		if !applied {
			*results = append(*results, Result{TableName: tb.TableName, RecordID: tb.RecordID, Status: StatusConflict, Message: "a newer version exists on the server"})
			conflicts = append(conflicts, newConflict(deviceID, tb.TableName, tb.RecordID, "DELETE",
				msToTime(localTime), msToTime(tb.UpdatedAt), ResolutionLocalWon, "server row is newer; delete discarded"))
			continue
		}
		// The deletion was accepted: register the tombstone globally so
		// other devices learn about it even though the row is gone.
		if err := s.repo.SetTombstone(ctx, &Tombstone{TableName: tb.TableName, RecordID: tb.RecordID, UpdatedAt: tb.UpdatedAt}); err != nil {
			s.log.Warn("sync: set tombstone failed", "table", tb.TableName, "record", tb.RecordID, "error", err)
		}
		*results = append(*results, Result{TableName: tb.TableName, RecordID: tb.RecordID, Status: StatusApplied})
	}

	return conflicts
}

// collect gathers every change newer than the caller's cursors and
// computes the new watermark per table.
func (s *Server) collect(ctx context.Context, cursors map[string]int64, results *[]Result) *Response {
	resp := &Response{
		Results:    *results,
		Rows:       make([]Change, 0),
		Tombstones: make([]Tombstone, 0),
		Cursors:    make(map[string]int64, len(SyncedTables)),
		ServerTime: time.Now().UTC().UnixMilli(),
	}
	for _, meta := range SyncedTables {
		after := cursors[meta.Name]
		rows, wm, err := s.repo.RowsChangedSince(ctx, meta, after)
		if err != nil {
			s.log.Warn("sync: collect rows failed", "table", meta.Name, "error", err)
			continue
		}
		tombs, err := s.repo.TombstonesSince(ctx, meta.Name, after)
		if err != nil {
			s.log.Warn("sync: collect tombstones failed", "table", meta.Name, "error", err)
			continue
		}
		for _, t := range tombs {
			if t.UpdatedAt > wm {
				wm = t.UpdatedAt
			}
		}
		resp.Rows = append(resp.Rows, rows...)
		resp.Tombstones = append(resp.Tombstones, tombs...)
		resp.Cursors[meta.Name] = wm
	}
	return resp
}

func withAPIKey(next http.Handler, apiKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey != "" && r.Header.Get("X-API-Key") != apiKey {
			writeError(w, http.StatusUnauthorized, "invalid or missing X-API-Key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return newID() + newID()
	}
	return hex.EncodeToString(b)
}
