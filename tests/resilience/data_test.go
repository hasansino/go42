//go:build resilience

package resilience

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go42-dev/go42/internal/database/mysql"
	"github.com/go42-dev/go42/internal/database/pgsql"
)

const (
	concurrentOperationCount = 8
	connectionFlapCount      = 2
	latencyToxicName         = "read_latency"
)

func TestDataDependenciesRecoverFromLatency(t *testing.T) {
	for _, factory := range dataDependencyFactories() {
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			resetProxy(t, factory.proxy)
			client, err := factory.open(t.Context(), startupTestTimeout)
			if err != nil {
				t.Fatalf("initialize dependency: %v", err)
			}
			defer shutdownClient(t, client)
			assertOperationSucceeds(t, client.operation)

			addToxic(t, factory.proxy.Name, toxicConfig{
				Name:     latencyToxicName,
				Type:     "latency",
				Stream:   "downstream",
				Toxicity: 1,
				Attributes: map[string]any{
					"latency": 750,
					"jitter":  50,
				},
			})

			operationCtx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			started := time.Now()
			err = client.operation(operationCtx)
			cancel()
			if err == nil {
				t.Fatal("dependency operation succeeded despite latency exceeding its deadline")
			}
			if elapsed := time.Since(started); elapsed > time.Second {
				t.Errorf("dependency operation took %s to honor its timeout", elapsed)
			}

			removeToxic(t, factory.proxy.Name, latencyToxicName)
			assertOperationSucceeds(t, client.operation)
		})
	}
}

func TestDataDependenciesRecoverFromRepeatedConcurrentOutages(t *testing.T) {
	for _, factory := range dataDependencyFactories() {
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			resetProxy(t, factory.proxy)
			client, err := factory.open(t.Context(), startupTestTimeout)
			if err != nil {
				t.Fatalf("initialize dependency: %v", err)
			}
			defer shutdownClient(t, client)

			for cycle := 1; cycle <= connectionFlapCount; cycle++ {
				setProxyEnabled(t, factory.proxy.Name, false)
				errs := runConcurrentOperations(t.Context(), client.operation, concurrentOperationCount)
				for index, err := range errs {
					if err == nil {
						t.Errorf("cycle %d operation %d succeeded during outage", cycle, index)
					}
				}

				setProxyEnabled(t, factory.proxy.Name, true)
				assertOperationSucceeds(t, client.operation)
				errs = runConcurrentOperations(t.Context(), client.operation, concurrentOperationCount)
				for index, err := range errs {
					if err != nil {
						t.Errorf("cycle %d operation %d failed after recovery: %v", cycle, index, err)
					}
				}
			}
		})
	}
}

func TestDataDependenciesShutdownWhileDisconnected(t *testing.T) {
	for _, factory := range dataDependencyFactories() {
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			resetProxy(t, factory.proxy)
			client, err := factory.open(t.Context(), startupTestTimeout)
			if err != nil {
				t.Fatalf("initialize dependency: %v", err)
			}
			setProxyEnabled(t, factory.proxy.Name, false)

			shutdownCtx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()
			if err := client.shutdown(shutdownCtx); err != nil {
				t.Errorf("shutdown disconnected dependency: %v", err)
			}
		})
	}
}

func TestPostgresMasterAndSlaveFailuresAreIndependent(t *testing.T) {
	masterProxy := proxyConfig{
		Name:     postgresProxyName,
		Listen:   "0.0.0.0:15432",
		Upstream: "pgsql:5432",
		Enabled:  true,
	}
	slaveProxy := proxyConfig{
		Name:     "postgres_slave",
		Listen:   "0.0.0.0:15433",
		Upstream: "pgsql:5432",
		Enabled:  true,
	}
	resetProxy(t, masterProxy)
	resetProxy(t, slaveProxy)
	masterDSN := fmt.Sprintf(
		"postgres://user:qwerty@%s/go42?sslmode=disable",
		envOrDefault(postgresAddressEnv, defaultPostgresAddress),
	)
	slaveDSN := "postgres://user:qwerty@127.0.0.1:15433/go42?sslmode=disable"
	database, err := pgsql.Open(
		t.Context(),
		masterDSN,
		slaveDSN,
		pgsql.WithConnectRetryTimeout(startupTestTimeout),
		pgsql.WithConnectRetryBackoff(50*time.Millisecond, 200*time.Millisecond),
		pgsql.WithMaxIdleConns(1),
	)
	if err != nil {
		t.Fatalf("initialize PostgreSQL master/slave: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = database.Shutdown(ctx)
	}()

	setProxyEnabled(t, slaveProxy.Name, false)
	assertEventuallyFails(t, "observe PostgreSQL slave outage", func() error {
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()
		return database.Ping(ctx)
	})
	if err := database.Master().WithContext(t.Context()).Exec("select 1").Error; err != nil {
		t.Errorf("PostgreSQL master failed while slave was unavailable: %v", err)
	}
	setProxyEnabled(t, slaveProxy.Name, true)
	assertPingSucceeds(t, database.Ping)

	setProxyEnabled(t, masterProxy.Name, false)
	assertEventuallyFails(t, "observe PostgreSQL master outage", func() error {
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()
		return database.Ping(ctx)
	})
	if err := database.Slave().WithContext(t.Context()).Exec("select 1").Error; err != nil {
		t.Errorf("PostgreSQL slave failed while master was unavailable: %v", err)
	}
	setProxyEnabled(t, masterProxy.Name, true)
	assertPingSucceeds(t, database.Ping)
}

func TestMySQLMasterAndSlaveFailuresAreIndependent(t *testing.T) {
	masterProxy := proxyConfig{
		Name:     mysqlProxyName,
		Listen:   "0.0.0.0:13306",
		Upstream: "mysql:3306",
		Enabled:  true,
	}
	slaveProxy := proxyConfig{
		Name:     "mysql_slave",
		Listen:   "0.0.0.0:13307",
		Upstream: "mysql:3306",
		Enabled:  true,
	}
	resetProxy(t, masterProxy)
	resetProxy(t, slaveProxy)
	masterDSN := fmt.Sprintf(
		"user:qwerty@tcp(%s)/go42?parseTime=true",
		envOrDefault(mysqlAddressEnv, defaultMySQLAddress),
	)
	slaveDSN := "user:qwerty@tcp(127.0.0.1:13307)/go42?parseTime=true"
	database, err := mysql.Open(
		t.Context(),
		masterDSN,
		slaveDSN,
		mysql.WithConnectRetryTimeout(startupTestTimeout),
		mysql.WithConnectRetryBackoff(50*time.Millisecond, 200*time.Millisecond),
		mysql.WithMaxIdleConns(1),
	)
	if err != nil {
		t.Fatalf("initialize MySQL master/slave: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = database.Shutdown(ctx)
	}()

	setProxyEnabled(t, slaveProxy.Name, false)
	assertEventuallyFails(t, "observe MySQL slave outage", func() error {
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()
		return database.Ping(ctx)
	})
	if err := database.Master().WithContext(t.Context()).Exec("select 1").Error; err != nil {
		t.Errorf("MySQL master failed while slave was unavailable: %v", err)
	}
	setProxyEnabled(t, slaveProxy.Name, true)
	assertPingSucceeds(t, database.Ping)

	setProxyEnabled(t, masterProxy.Name, false)
	assertEventuallyFails(t, "observe MySQL master outage", func() error {
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()
		return database.Ping(ctx)
	})
	if err := database.Slave().WithContext(t.Context()).Exec("select 1").Error; err != nil {
		t.Errorf("MySQL slave failed while master was unavailable: %v", err)
	}
	setProxyEnabled(t, masterProxy.Name, true)
	assertPingSucceeds(t, database.Ping)
}

func runConcurrentOperations(
	ctx context.Context,
	operation func(context.Context) error,
	count int,
) []error {
	errs := make([]error, count)
	var waitGroup sync.WaitGroup
	waitGroup.Add(count)
	for index := range count {
		go func() {
			defer waitGroup.Done()
			operationCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			defer cancel()
			errs[index] = operation(operationCtx)
		}()
	}
	waitGroup.Wait()
	return errs
}
