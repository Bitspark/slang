package elem

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// An Operation names one kind of effect a built-in performs through a capability.
// Operations, not transports, are what a profile implements and permits
// (decision 0012, after research 0007).
type Operation string

const (
	HTTPRequest   Operation = "http.request"
	InboundListen Operation = "inbound.listen"
	MailSend      Operation = "mail.send"
	MQTTPublish   Operation = "mqtt.publish"
	MQTTSubscribe Operation = "mqtt.subscribe"
	SQLQuery      Operation = "sql.query"
	SQLExecute    Operation = "sql.execute"
	FilesRead     Operation = "files.read"
	FilesReplace  Operation = "files.replace"
	FilesAppend   Operation = "files.append"
	ClockRead     Operation = "clock.read"
	ClockWait     Operation = "clock.wait"
	RandomRead    Operation = "random.read"
	StateRead     Operation = "state.read"
	StateWrite    Operation = "state.write"
	LogEmit       Operation = "log.emit"
)

// Operations lists every operation a built-in of this engine can require.
var Operations = []Operation{HTTPRequest, InboundListen, MailSend, MQTTPublish, MQTTSubscribe,
	SQLQuery, SQLExecute, FilesRead, FilesReplace, FilesAppend, ClockRead, ClockWait,
	RandomRead, StateRead, StateWrite, LogEmit}

// A Requirement is one operation a built-in performs, and where its target comes
// from: Properties fix it before the program starts, while a Dynamic target
// arrives in an input item and can be judged only when the effect happens.
type Requirement struct {
	Operation  Operation
	Properties []string
	Dynamic    bool
}

// A Profile is the rules an execution runs under: the operations its host
// implements and those an execution is permitted to use. Its version names
// exactly these rules, so changing them means a new version.
type Profile struct {
	Name        string
	Version     int
	Implemented []Operation
	Permitted   []Operation
}

// ID names the profile's immutable version, such as "hosted/1".
func (p Profile) ID() string {
	return fmt.Sprintf("%s/%d", p.Name, p.Version)
}

// LocalProfile runs programs with their user's permissions: this machine
// implements and permits every operation.
var LocalProfile = Profile{Name: "local", Version: 1, Implemented: Operations, Permitted: Operations}

// hostedOperations are what a hosted program can do: HTTP through its capability
// host, and inside its own container the clock, randomness, its own state and
// its log. The runner has no network, no file namespace of its own and no
// listening socket (decision 0006).
var hostedOperations = []Operation{HTTPRequest, ClockRead, ClockWait, RandomRead, StateRead, StateWrite, LogEmit}

// HostedProfile is the hosted runtime's.
var HostedProfile = Profile{Name: "hosted", Version: 1, Implemented: hostedOperations, Permitted: hostedOperations}

// PublicProfile is for untrusted visitors on a machine that implements
// everything: it permits computation, the clock, randomness and a program's own
// state, and denies every other effect.
var PublicProfile = Profile{Name: "public", Version: 1, Implemented: Operations,
	Permitted: []Operation{ClockRead, ClockWait, RandomRead, StateRead, StateWrite}}

// Profiles are the profiles a host can select by name.
var Profiles = []Profile{LocalProfile, HostedProfile, PublicProfile}

// ProfileNamed returns the profile called name.
func ProfileNamed(name string) (Profile, error) {
	var names []string
	for _, p := range Profiles {
		if p.Name == name {
			return p, nil
		}
		names = append(names, p.Name)
	}
	return Profile{}, fmt.Errorf("unknown profile %q: must be one of %s", name, strings.Join(names, ", "))
}

// A State says what a profile can do with a built-in ID. The order of precedence
// is removed, unavailable, denied, available; an ID that is no built-in is unknown.
type State string

const (
	Available   State = "available"
	Denied      State = "denied"
	Unavailable State = "unavailable"
	Removed     State = "removed"
	Unknown     State = "unknown"
)

// A Cause is one required operation that keeps a built-in from being available.
type Cause struct {
	Operation Operation
	State     State
}

// A Resolution is what a profile makes of one built-in, and why. Causes lists
// every operation standing in the way, even when one state takes precedence.
type Resolution struct {
	ID           uuid.UUID
	Name         string
	State        State
	Profile      string
	Requirements []Requirement
	Causes       []Cause
	RemovedIn    string
}

// Err explains why the built-in cannot be used, or is nil when it can.
func (r Resolution) Err() error {
	switch r.State {
	case Available:
		return nil
	case Unknown:
		return fmt.Errorf("unknown built-in %s", r.ID)
	case Removed:
		return fmt.Errorf("built-in %q was removed %s", r.Name, r.RemovedIn)
	}
	var causes []string
	for _, c := range r.Causes {
		verb := "is not implemented"
		if c.State == Denied {
			verb = "is not permitted"
		}
		causes = append(causes, fmt.Sprintf("%s %s", c.Operation, verb))
	}
	return fmt.Errorf("built-in %q is %s under profile %s: %s", r.Name, r.State, r.Profile, strings.Join(causes, ", "))
}

// Resolve decides one built-in's state under p from its declared requirements.
// It never constructs the built-in.
func (p Profile) Resolve(id uuid.UUID) Resolution {
	r := Resolution{ID: id, Profile: p.ID()}
	if removal, ok := removedBuiltins[id]; ok {
		r.Name, r.State, r.RemovedIn = removal.name, Removed, removal.removedIn
		return r
	}
	cfg, ok := cfgs[id]
	if !ok {
		r.State = Unknown
		return r
	}
	r.Name, r.Requirements, r.State = cfg.blueprint.Meta.Name, cfg.requires, Available
	for _, req := range cfg.requires {
		switch {
		case !contains(p.Implemented, req.Operation):
			r.Causes = append(r.Causes, Cause{req.Operation, Unavailable})
			r.State = Unavailable
		case !contains(p.Permitted, req.Operation):
			r.Causes = append(r.Causes, Cause{req.Operation, Denied})
			if r.State == Available {
				r.State = Denied
			}
		}
	}
	return r
}

// Catalog resolves every built-in this engine knows under p, removed ones
// included, sorted by name.
func (p Profile) Catalog() []Resolution {
	var entries []Resolution
	for id := range cfgs {
		entries = append(entries, p.Resolve(id))
	}
	for id := range removedBuiltins {
		entries = append(entries, p.Resolve(id))
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}

func contains(operations []Operation, op Operation) bool {
	for _, o := range operations {
		if o == op {
			return true
		}
	}
	return false
}

type removal struct{ name, removedIn string }

// removedBuiltins reserves the IDs of built-ins this engine no longer provides,
// so that no document can supply a definition under them (API design 5.6).
var removedBuiltins = map[uuid.UUID]removal{
	uuid.MustParse("199f14c3-3e25-4813-aaba-7ec7fa3d94e2"): {"semaphore P", "after v0.1.27"},
	uuid.MustParse("dc9b35a3-bd0e-4ca3-99df-4e2689ea5097"): {"semaphore V", "after v0.1.27"},
	uuid.MustParse("b8771c73-cddf-4eb1-a10c-bf78c2552efe"): {"memory get", "after v0.1.27"},
	uuid.MustParse("3be41b5b-5a43-4f94-a7ae-7f0bacc4ae77"): {"memory set", "after v0.1.27"},
	uuid.MustParse("b6cb78ca-bbfd-475e-a11f-3593ce295e3c"): {"Kafka subscribe", "after v0.1.27"},
	uuid.MustParse("362482c1-2021-4e5c-9463-b580a6c1967e"): {"Redis Get", "after v0.1.27"},
	uuid.MustParse("4b946e4a-e26b-45c7-9759-c60bd57d190d"): {"Redis HGet", "after v0.1.27"},
	uuid.MustParse("8d9e4c6e-20a2-44b1-8d51-ed98f4d3b4d8"): {"Redis HIncr", "after v0.1.27"},
	uuid.MustParse("a6b45f70-e20c-40a5-ac39-c00068d10c81"): {"Redis HSet", "after v0.1.27"},
	uuid.MustParse("8f8a095c-9274-4d39-96d9-3ef463659426"): {"Redis LPush", "after v0.1.27"},
	uuid.MustParse("cdbf3e0d-1ce0-4565-9df6-d0e829c730e5"): {"Redis Set", "after v0.1.27"},
	uuid.MustParse("eb3fd302-f6b0-4c2a-b353-ff0a01e49d09"): {"Redis Subscribe", "after v0.1.27"},
	uuid.MustParse("cf20bcec-2028-45b4-a00c-0ce348c381c4"): {"meta store", "after v0.1.27"},
	uuid.MustParse("13cbad40-da00-40d7-bdcd-981b14ec346b"): {"shell execute", "after v0.1.27"},
	uuid.MustParse("5b704038-9617-454a-b7a1-2091277cff69"): {"window", "after v0.1.27"},
}
