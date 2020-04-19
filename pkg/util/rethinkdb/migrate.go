package rethinkdb

import (
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/grants"

	rdb "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

func (r *rethinkDBSession) Migrate(adminPass string, desiredReplicas, desiredShards int32, allowAnonymous bool) error {
	// Setup DBs
	dbs, err := r.listDBs()
	if err != nil {
		return err
	}
	if contains(dbs, "test") {
		rdbLogger.Info("Deleting 'test' database")
		if err := r.deleteDB("test"); err != nil {
			return err
		}
	}
	if !contains(dbs, kvdiDB) {
		rdbLogger.Info("Creating new database", "Database.Name", kvdiDB)
		if err := r.createDB(kvdiDB); err != nil {
			return err
		}
	}

	// Setup Tables
	tables, err := r.listDBTables(kvdiDB)
	if err != nil {
		return err
	}
	for _, table := range allTables {
		if !contains(tables, table) {
			rdbLogger.Info("Creating new table", "Database.Name", kvdiDB, "Table.Name", table)
			if err := r.createTable(kvdiDB, table); err != nil {
				return err
			}
		}
		shards, replicas, err := r.getTableConfig(kvdiDB, table)
		if err != nil {
			return err
		}
		if replicas != desiredReplicas || shards != desiredShards {
			rdbLogger.Info("Configuring table sharding and replication", "Table.Name", table, "Replicas", desiredReplicas, "Shards", desiredShards)
			if cursor, err := rdb.DB(kvdiDB).Table(table).Reconfigure(rdb.ReconfigureOpts{
				Replicas: desiredReplicas,
				Shards:   desiredShards,
			}).Run(r.session); err != nil {
				return err
			} else if cursor.Err() != nil {
				return cursor.Err()
			}
		}
	}

	// Ensure an admin role
	if _, err := r.GetRole(adminRole); err != nil {
		if !errors.IsRoleNotFoundError(err) {
			return err
		}
		rdbLogger.Info("Creating new 'admin' role...")
		if err := r.CreateRole(&Role{
			Name:   adminRole,
			Grants: grants.All,
		}); err != nil {
			return err
		}
	}

	// Ensure a launch template role
	if _, err := r.GetRole(launchTemplateRole); err != nil {
		if !errors.IsRoleNotFoundError(err) {
			return err
		}
		rdbLogger.Info("Creating new 'launch-templates' role...")
		if err := r.CreateRole(&Role{
			Name:   launchTemplateRole,
			Grants: grants.LaunchTemplatesGrant,
		}); err != nil {
			return err
		}
	}

	// Ensure an admin user
	if user, err := r.GetUser(adminUser); err != nil {
		if !errors.IsUserNotFoundError(err) {
			return err
		}
		rdbLogger.Info("Creating new 'admin' user...")
		if err := r.CreateUser(&User{Name: adminUser, Password: adminPass, Roles: []*Role{{Name: adminRole}}}); err != nil {
			return err
		}
	} else if user.PasswordSalt == "" || !util.PasswordMatchesHash(adminPass, user.PasswordSalt) {
		rdbLogger.Info("Admin password salt in database doesn't match provided password, updating...")
		if err := r.SetUserPassword(user, adminPass); err != nil {
			return err
		}
	}

	// Ensure anonymous user status
	if allowAnonymous {
		if _, err := r.GetUser(anonymousUser); err != nil {
			if !errors.IsUserNotFoundError(err) {
				return err
			}
			rdbLogger.Info("Creating new 'anonymous' user...")
			if err := r.CreateUser(&User{Name: anonymousUser, Password: "", Roles: []*Role{{Name: launchTemplateRole}}}); err != nil {
				return err
			}
		}
	} else {
		if _, err := r.GetUser(anonymousUser); err == nil {
			rdbLogger.Info("Deleting 'anonymous' user...")
			if err := r.DeleteUser(&User{Name: anonymousUser}); err != nil {
				return err
			}
		} else if !errors.IsUserNotFoundError(err) {
			return err
		}
	}

	return nil
}
