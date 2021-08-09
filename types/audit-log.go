package types

// taken from https://www.enterpriseready.io/features/audit-log/
/*
When an event is logged it should have details that provide enough information about the event to provide the necessary context of who, what, when and where etc. Specifically, the follow fields are critical to an audit log:

Actor - The username, uuid, API token name of the account taking the action.
Group - The group (aka organization, team, account) that the actor is a member of (needed to show admins the full history of their group).
Where - IP address, device ID, country.
When - The NTP synced server time when the event happened.
Target - the object or underlying resource that is being changed (the noun) as well as the fields that include a key value for the new state of the target.
Action - the way in which the object was changed (the verb).
Action Type - the corresponding C``R``U``D category.
Event Name - Common name for the event that can be used to filter down to similar events.
Description - A human readable description of the action taken, sometimes includes links to other pages within the application.
Optional information
Server server ids or names, server location.
Version version of the code that is sending the events.
Protocols ie http vs https.
Global Actor ID if a customer is using Single Sign On, it might be important to also include a Global UID if it differs from the application specific ID.
*/
type AuditLog struct {
	Actor       string            `auditdb:"index" json:"actor"` // The username, uuid, API token name of the account taking the action.
	ActorType   string            `auditdb:"index" json:"actor_type"`
	Group       string            `auditdb:"index" json:"group"` // The group (aka organization, team, account) that the actor is a member of (needed to show admins the full history of their group).
	Where       string            `auditdb:"index" json:"where"` // IP address, device ID, country.
	WhereType   string            `auditdb:"index" json:"where_type"`
	When        string            `json:"when"`                      // The NTP synced RFC3339 server time when the event happened.
	Target      string            `auditdb:"index" json:"target"`    // The object or underlying resource that is being changed (the noun) as well as the fields that include a key value for the new state of the target.
	TargetID    string            `auditdb:"index" json:"target_id"` // The ID (optional) of the target
	Action      string            `auditdb:"index" json:"action"`    // The way in which the object was changed (the verb).
	ActionType  string            `auditdb:"index" json:"action_type"`
	Name        string            `auditdb:"index" json:"name"` // Common name for the event that can be used to filter down to similar events.
	Description string            `json:"description"`          // A human readable description of the action taken, sometimes includes links to other pages within the application.
	Metadata    map[string]string `json:"metadata"`

	TS int64 `json:"-"` // timestamp, used only for sorting
}

func (a AuditLog) Indexes() (map[string]string, map[string]interface{}) {
	return getIndexes(a)
}
