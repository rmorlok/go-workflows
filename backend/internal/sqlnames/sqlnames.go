package sqlnames

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"text/template"
	"time"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Config struct {
	QuoteLeft           string
	QuoteRight          string
	MaxIdentifierLength int
	PrefixIndexes       bool
}

type Names struct {
	TablePrefix string

	Instances     string
	PendingEvents string
	History       string
	Activities    string
	Attributes    string

	IdxInstancesInstanceIDExecutionID             string
	IdxInstancesIDExecutionID                     string
	IdxInstancesLockedUntilCompletedAt            string
	IdxInstancesLockedUntilCompletedAtQueue       string
	IdxInstancesParentInstanceIDParentExecutionID string

	IdxPendingEventsInIDExID                                      string
	IdxPendingEventsInIDExIDVisibleAtScheduleEventID              string
	IdxPendingEventsInstanceIDExecutionIDVisibleAtScheduleEventID string

	IdxHistoryInstanceIDExecutionID         string
	IdxHistoryInstanceIDExecutionIDSequence string
	IdxHistoryInstanceSequence              string

	IdxActivitiesIDWorker                              string
	IdxActivitiesInstanceIDExecutionIDActivityIDWorker string
	IdxActivitiesInstanceIDExecutionIDWorker           string
	IdxActivitiesInstanceIDExecutionIDWorkerQueue      string
	IdxActivitiesLockedUntil                           string
	IdxActivitiesLockedUntilQueue                      string

	IdxAttributesInstanceIDExecutionIDEventID string
	IdxAttributesEventID                      string

	migrationsTable string
}

var validIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func New(tablePrefix string, config Config) (Names, error) {
	n := Names{
		TablePrefix:     tablePrefix,
		migrationsTable: prefixedTableName(tablePrefix, "schema_migrations"),
	}

	tables := map[string]*string{
		"instances":      &n.Instances,
		"pending_events": &n.PendingEvents,
		"history":        &n.History,
		"activities":     &n.Activities,
		"attributes":     &n.Attributes,
	}
	for base, target := range tables {
		raw := prefixedTableName(tablePrefix, base)
		if err := validateIdentifier(raw, config.MaxIdentifierLength); err != nil {
			return Names{}, err
		}
		*target = quoteIdentifier(raw, config.QuoteLeft, config.QuoteRight)
	}
	if err := validateIdentifier(n.migrationsTable, config.MaxIdentifierLength); err != nil {
		return Names{}, err
	}

	indexes := map[string]*string{
		"idx_instances_instance_id_execution_id":                                   &n.IdxInstancesInstanceIDExecutionID,
		"idx_instances_id_execution_id":                                            &n.IdxInstancesIDExecutionID,
		"idx_instances_locked_until_completed_at":                                  &n.IdxInstancesLockedUntilCompletedAt,
		"idx_instances_locked_until_completed_at_queue":                            &n.IdxInstancesLockedUntilCompletedAtQueue,
		"idx_instances_parent_instance_id_parent_execution_id":                     &n.IdxInstancesParentInstanceIDParentExecutionID,
		"idx_pending_events_inid_exid":                                             &n.IdxPendingEventsInIDExID,
		"idx_pending_events_inid_exid_visible_at_schedule_event_id":                &n.IdxPendingEventsInIDExIDVisibleAtScheduleEventID,
		"idx_pending_events_instance_id_execution_id_visible_at_schedule_event_id": &n.IdxPendingEventsInstanceIDExecutionIDVisibleAtScheduleEventID,
		"idx_history_instance_id_execution_id":                                     &n.IdxHistoryInstanceIDExecutionID,
		"idx_history_instance_id_execution_id_sequence_id":                         &n.IdxHistoryInstanceIDExecutionIDSequence,
		"idx_history_instance_sequence_id":                                         &n.IdxHistoryInstanceSequence,
		"idx_activities_id_worker":                                                 &n.IdxActivitiesIDWorker,
		"idx_activities_instance_id_execution_id_activity_id_worker":               &n.IdxActivitiesInstanceIDExecutionIDActivityIDWorker,
		"idx_activities_instance_id_execution_id_worker":                           &n.IdxActivitiesInstanceIDExecutionIDWorker,
		"idx_activities_instance_id_execution_id_worker_queue":                     &n.IdxActivitiesInstanceIDExecutionIDWorkerQueue,
		"idx_activities_locked_until":                                              &n.IdxActivitiesLockedUntil,
		"idx_activities_locked_until_queue":                                        &n.IdxActivitiesLockedUntilQueue,
		"idx_attributes_instance_id_execution_id_event_id":                         &n.IdxAttributesInstanceIDExecutionIDEventID,
		"idx_attributes_event_id":                                                  &n.IdxAttributesEventID,
	}
	for base, target := range indexes {
		raw := base
		maxIdentifierLength := 0
		if tablePrefix != "" && config.PrefixIndexes {
			raw = prefixedIndexName(tablePrefix, base, config.MaxIdentifierLength)
			maxIdentifierLength = config.MaxIdentifierLength
		}
		if err := validateIdentifier(raw, maxIdentifierLength); err != nil {
			return Names{}, err
		}
		*target = quoteIdentifier(raw, config.QuoteLeft, config.QuoteRight)
	}

	return n, nil
}

func (n Names) MigrationsTable(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if n.TablePrefix == "" {
		return ""
	}
	return n.migrationsTable
}

func (n Names) MigrationSource(fsys fs.FS, path string) (source.Driver, error) {
	return iofs.New(renderFS{
		fsys: fsys,
		data: n.templateData(),
	}, path)
}

func (n Names) templateData() map[string]string {
	return map[string]string{
		"Instances":     n.Instances,
		"PendingEvents": n.PendingEvents,
		"History":       n.History,
		"Activities":    n.Activities,
		"Attributes":    n.Attributes,

		"IdxInstancesInstanceIDExecutionID":             n.IdxInstancesInstanceIDExecutionID,
		"IdxInstancesIDExecutionID":                     n.IdxInstancesIDExecutionID,
		"IdxInstancesLockedUntilCompletedAt":            n.IdxInstancesLockedUntilCompletedAt,
		"IdxInstancesLockedUntilCompletedAtQueue":       n.IdxInstancesLockedUntilCompletedAtQueue,
		"IdxInstancesParentInstanceIDParentExecutionID": n.IdxInstancesParentInstanceIDParentExecutionID,

		"IdxPendingEventsInIDExID":                                      n.IdxPendingEventsInIDExID,
		"IdxPendingEventsInIDExIDVisibleAtScheduleEventID":              n.IdxPendingEventsInIDExIDVisibleAtScheduleEventID,
		"IdxPendingEventsInstanceIDExecutionIDVisibleAtScheduleEventID": n.IdxPendingEventsInstanceIDExecutionIDVisibleAtScheduleEventID,

		"IdxHistoryInstanceIDExecutionID":         n.IdxHistoryInstanceIDExecutionID,
		"IdxHistoryInstanceIDExecutionIDSequence": n.IdxHistoryInstanceIDExecutionIDSequence,
		"IdxHistoryInstanceSequence":              n.IdxHistoryInstanceSequence,

		"IdxActivitiesIDWorker":                              n.IdxActivitiesIDWorker,
		"IdxActivitiesInstanceIDExecutionIDActivityIDWorker": n.IdxActivitiesInstanceIDExecutionIDActivityIDWorker,
		"IdxActivitiesInstanceIDExecutionIDWorker":           n.IdxActivitiesInstanceIDExecutionIDWorker,
		"IdxActivitiesInstanceIDExecutionIDWorkerQueue":      n.IdxActivitiesInstanceIDExecutionIDWorkerQueue,
		"IdxActivitiesLockedUntil":                           n.IdxActivitiesLockedUntil,
		"IdxActivitiesLockedUntilQueue":                      n.IdxActivitiesLockedUntilQueue,

		"IdxAttributesInstanceIDExecutionIDEventID": n.IdxAttributesInstanceIDExecutionIDEventID,
		"IdxAttributesEventID":                      n.IdxAttributesEventID,
	}
}

func prefixedTableName(prefix, base string) string {
	return prefix + base
}

func prefixedIndexName(prefix, base string, maxIdentifierLength int) string {
	name := prefix + base
	if maxIdentifierLength <= 0 || len(name) <= maxIdentifierLength {
		return name
	}

	sum := sha1.Sum([]byte(name))
	hash := hex.EncodeToString(sum[:])[:8]
	keep := maxIdentifierLength - len(hash) - 1
	return name[:keep] + "_" + hash
}

func validateIdentifier(name string, maxIdentifierLength int) error {
	if !validIdentifier.MatchString(name) {
		return fmt.Errorf("invalid SQL identifier %q: only letters, digits, and underscores are supported, and the first character must be a letter or underscore", name)
	}
	if maxIdentifierLength > 0 && len(name) > maxIdentifierLength {
		return fmt.Errorf("invalid SQL identifier %q: length %d exceeds maximum %d", name, len(name), maxIdentifierLength)
	}
	return nil
}

func quoteIdentifier(name, left, right string) string {
	return left + name + right
}

type renderFS struct {
	fsys fs.FS
	data map[string]string
}

func (r renderFS) Open(name string) (fs.File, error) {
	f, err := r.fsys.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if info.IsDir() {
		return f, nil
	}

	body, err := io.ReadAll(f)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(name).Option("missingkey=error").Parse(string(body))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.data); err != nil {
		return nil, err
	}

	return &renderedFile{
		Reader: bytes.NewReader(buf.Bytes()),
		info: renderedFileInfo{
			FileInfo: info,
			size:     int64(buf.Len()),
		},
	}, nil
}

type renderedFile struct {
	*bytes.Reader
	info fs.FileInfo
}

func (f *renderedFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *renderedFile) Close() error {
	return nil
}

type renderedFileInfo struct {
	fs.FileInfo
	size int64
}

func (i renderedFileInfo) Size() int64 {
	return i.size
}

func (i renderedFileInfo) ModTime() time.Time {
	return i.FileInfo.ModTime()
}
