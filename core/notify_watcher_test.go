package core_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/store"
)

// newNotifyTestAppPair creates two app instances that share the same data dir
// and the same isolated Postgres database pair, so that a change in one app
// is visible to the other (the fork uses Postgres instead of per-dir SQLite files).
func newNotifyTestAppPair(t *testing.T, dataDir string) (app1, app2 *core.BaseApp, cleanup func()) {
	t.Helper()

	suffix := strings.ToLower(security.PseudorandomString(8))
	dataDB := "pb_test_" + suffix + "_data_db"
	auxDB := "pb_test_" + suffix + "_auxiliary_db"

	app1 = core.NewBaseApp(core.BaseAppConfig{
		DataDir:        dataDir,
		PostgresDataDB: dataDB,
		PostgresAuxDB:  auxDB,
	})
	app2 = core.NewBaseApp(core.BaseAppConfig{
		DataDir:        dataDir,
		PostgresDataDB: dataDB,
		PostgresAuxDB:  auxDB,
	})

	cleanup = func() {
		app1.ResetBootstrapState()
		app2.ResetBootstrapState()
		exec.Command("sh", "-c", fmt.Sprintf("PGPASSWORD=admin dropdb --if-exists -h 127.0.0.1 -U postgres %s", dataDB)).Run()
		exec.Command("sh", "-c", fmt.Sprintf("PGPASSWORD=admin dropdb --if-exists -h 127.0.0.1 -U postgres %s", auxDB)).Run()
	}

	return app1, app2, cleanup
}

func TestNotifyWatcher_SettingsUpdate(t *testing.T) {
	t.Parallel()

	testEvents := store.New[core.App, int](nil)

	tmpDir, err := os.MkdirTemp("", "pb_notify_test*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app1, app2, cleanup := newNotifyTestAppPair(t, tmpDir)
	defer cleanup()

	if err := app1.Bootstrap(); err != nil {
		t.Fatal(err)
	}

	if err := app2.Bootstrap(); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{}, 1)

	app1.OnSettingsReload().BindFunc(func(e *core.SettingsReloadEvent) error {
		testEvents.SetFunc(app1, func(old int) int {
			return old + 1
		})
		return e.Next()
	})

	app2.OnSettingsReload().BindFunc(func(e *core.SettingsReloadEvent) error {
		err := e.Next()
		testEvents.SetFunc(app2, func(old int) int {
			return old + 1
		})
		// signal only after the settings were actually reloaded (e.Next)
		select {
		case done <- struct{}{}:
		default:
		}
		return err
	})

	// updating app1 settings should trigger a reload in app2
	app1.Settings().SuperuserIPs = []string{"127.0.0.1"}

	// The fsnotify watcher can miss the quick create+remove notify event on some
	// platforms (notably macOS kqueue), so retry the save a few times.
	reloaded := false
	for attempt := 0; attempt < 10; attempt++ {
		if err := app1.Save(app1.Settings()); err != nil {
			t.Fatal(err)
		}

		select {
		case <-done:
			reloaded = true
		case <-time.After(250 * time.Millisecond):
			// retry
		}

		if reloaded {
			break
		}
	}

	if !reloaded {
		t.Fatal("app2 reload event timeout")
	}

	if app1Total := testEvents.Get(app1); app1Total < 1 {
		t.Fatalf("Expected at least 1 app1 event, got %d", app1Total)
	}

	if app2Total := testEvents.Get(app2); app2Total < 1 {
		t.Fatalf("Expected at least 1 app2 event, got %d", app2Total)
	}

	app2SuperuserIPs := app2.Settings().SuperuserIPs
	if len(app2SuperuserIPs) != 1 || app2SuperuserIPs[0] != "127.0.0.1" {
		t.Fatalf("Expected exactly 127.0.0.1 superuser IP in app2 settings event, got %v", app2SuperuserIPs)
	}
}

func TestNotifyWatcher_CollectionsUpdate(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "pb_notify_test*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app1, app2, cleanup := newNotifyTestAppPair(t, tmpDir)
	defer cleanup()

	if err := app1.Bootstrap(); err != nil {
		t.Fatal(err)
	}

	if err := app2.Bootstrap(); err != nil {
		t.Fatal(err)
	}

	testQueries := store.New[string, []string](nil)
	// Note: the Postgres fork uses a single connection pool (ConcurrentDB() == NonconcurrentDB()),
	// so all queries are tracked through ConcurrentDB() only.
	app2.ConcurrentDB().(*dbx.DB).QueryLogFunc = func(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
		testQueries.SetFunc("concurrent", func(old []string) []string {
			return append(old, sql)
		})
	}
	app2.ConcurrentDB().(*dbx.DB).ExecLogFunc = func(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
		testQueries.SetFunc("concurrent", func(old []string) []string {
			return append(old, sql)
		})
	}

	// create/update/delete app1 collections should trigger a reload in app2.
	// The fsnotify watcher can miss notify events on some platforms (macOS kqueue),
	// so retry until a reload query is observed.
	deadline := time.After(10 * time.Second)
	for len(testQueries.Get("concurrent")) == 0 {
		dummyCollection := core.NewBaseCollection("test")
		if err := app1.Save(dummyCollection); err != nil {
			t.Fatal(err)
		}
		dummyCollection.Fields.Add(&core.TextField{Name: "test"})
		if err := app1.Save(dummyCollection); err != nil {
			t.Fatal(err)
		}
		if err := app1.Delete(dummyCollection); err != nil {
			t.Fatal(err)
		}

		select {
		case <-deadline:
			t.Fatal("reload query not observed")
		case <-time.After(100 * time.Millisecond):
		}
	}

	concurrentQueries := testQueries.Get("concurrent")

	expectedQuery := `SELECT {{_collections}}.* FROM "_collections" ORDER BY "rowid" ASC`
	if concurrentQueries[0] != expectedQuery {
		t.Fatalf("Expected query\n%s\ngot\n%s", expectedQuery, concurrentQueries[0])
	}
}
