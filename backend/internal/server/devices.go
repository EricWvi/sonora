package server

import (
	"net/http"

	"sonora/internal/admin"
	"sonora/internal/auth"
)

// Device management for native clients, adapted from asr
// (internal/server/devices.go) without the push connection.

func (s *server) listDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.store.ListDevices(r.Context(), auth.User(r.Context()).ID)
	respond(w, r, devices, err)
}

// revokeDevice signs a device out. A device may revoke itself, which is
// how the client signs out; other devices are revoked with a web session.
func (s *server) revokeDevice(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if d, isDevice := auth.Device(r.Context()); isDevice && d.ID != id {
		admin.WriteError(w, http.StatusForbidden, "forbidden", "只能在网页上吊销其他设备")
		return
	}
	respondEmpty(w, r, s.store.DeleteDevice(r.Context(), auth.User(r.Context()).ID, id))
}
