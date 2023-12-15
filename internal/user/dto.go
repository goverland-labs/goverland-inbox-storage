package user

type CreateSessionRequest struct {
	Address        *string `json:"address"`
	GuestSessionID *string `json:"guest_session_id"`
	DeviceUUID     string  `json:"device_uuid"`
	DeviceName     string  `json:"device_name"`
	Role           Role    `json:"role"`
}

type ProfileInfo struct {
	User         *User     `json:"user"`
	LastSessions []Session `json:"last_sessions"`
}
