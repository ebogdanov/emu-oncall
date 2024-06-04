package grafana

const (
	callback = "callback"
)

type alertGroupResponse struct {
	Next       string       `json:"next,omitempty"`
	Previous   string       `json:"previous,omitempty"`
	Results    []AlertGroup `json:"results"`
	Message    string       `json:"message,omitempty"`
	MsgID      string       `json:"messageId,omitempty"`
	StatusCode int          `json:"statusCode,omitempty"`
	TraceID    string       `json:"traceID,omitempty"`
	Detail     string       `json:"detail,omitempty"`
}

type alertReceiveChannel struct {
	ID          string `json:"id"`
	Integration string `json:"integration"`
	VerbalName  string `json:"verbal_name"`
	Deleted     bool   `json:"deleted"`
}

type renderWeb struct {
	Title      string `json:"title"`
	Message    string `json:"message"`
	ImageURL   string `json:"image_url"`
	SourceLink string `json:"source_link"`
}

type acknowledgedByUser struct {
	ID       string `json:"pk"`
	Username string `json:"username"`
}

type relatedUser struct {
	Username string `json:"username"`
	ID       string `json:"pk"`
	Avatar   string `json:"avatar"`
}

type renderMarkdown struct {
	Title      string `json:"title"`
	Message    string `json:"message"`
	ImageURL   string `json:"image_url"`
	SourceLink string `json:"source_link"`
}

type Ok struct {
	Success bool `json:"success"`
}

type Err struct {
	AccessErrorID string `json:"accessErrorId,omitempty"`
	Message       string `json:"message"`
	Title         string `json:"title,omitempty"`
}

type User struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Login         string   `json:"login"`
	Email         string   `json:"email,omitempty"`
	IsAdmin       bool     `json:"isAdmin"`
	IsDisabled    bool     `json:"isDisabled"`
	LastSeenAt    string   `json:"lastSeenAt"`
	LastSeenAtAge string   `json:"lastSeenAtAge"`
	AuthLabels    []string `json:"authLabels"`
}

type ScheduleItem struct {
	ID          string   `json:"id"`
	TeamID      string   `json:"team_id"`
	Name        string   `json:"name"`
	TimeZone    string   `json:"time_zone"`
	Users       []string `json:"on_call_now"`
	Shifts      []string `json:"shifts"`
	transport   string
	callbackURL string
}

type scheduleEvent struct {
	Title        string
	Transport    string
	UserID       string
	Msg          string
	ScheduleName string
	URL          string
}

// "ok" or "alerting"
type formattedAlert struct {
	AlertUID              string `json:"alert_uid,omitempty"`
	Title                 string `json:"title,omitempty"`
	ImageURL              string `json:"image_url,omitempty"`
	State                 string `json:"state,omitempty"`
	LinkToUpstreamDetails string `json:"link_to_upstream_details"`
	Message               string `json:"message"`
}
