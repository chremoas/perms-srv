package handler

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/chremoas/services-common/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	sq "github.com/Masterminds/squirrel"

	permSrv "github.com/chremoas/perms-srv/proto"
)

type permissionsHandler struct {
	db sq.StatementBuilderType
	*zap.Logger
	namespace string
}

var (
	errCannotAdd     = errors.New("you cannot add to the server_admins group")
	errCannotDelete  = errors.New("you cannot delete from the server_admins group")
	errCannotRemove  = errors.New("you cannot remove users from the server_admins group")
	errNoPermissions = errors.New("user has no permissions")
)

func NewPermissionsHandler(config *config.Configuration, log *zap.Logger) (permSrv.PermissionsHandler, error) {
	var (
		sugar = log.Sugar()
		err   error
	)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.roledb"),
	)

	ldb, err := sqlx.Connect(viper.GetString("database.driver"), dsn)
	if err != nil {
		return nil, err
	}

	err = ldb.Ping()
	if err != nil {
		return nil, err
	}

	dbCache := sq.NewStmtCache(ldb)
	db := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(dbCache)

	// Ensure required permissions exist in the database
	var (
		id                    int
		permissionName        = "server_admins"
		permissionDescription = "Server Admins"
	)
	err = db.Select("id").
		From("permissions").
		Where(sq.Eq{"name": "server_admins"}).
		Where(sq.Eq{"namespace": config.Namespace}).
		QueryRow().Scan(&id)

	switch err {
	case nil:
		sugar.Infof("%s found", permissionName)
	case sql.ErrNoRows:
		sugar.Infof("%s NOT found, creating", permissionName)
		err = db.Insert("permissions").
			Columns("namespace", "name", "description").
			Values(config.Namespace, permissionName, permissionDescription).
			Suffix("RETURNING \"id\"").
			QueryRow().Scan(&id)
		if err != nil {
			return nil, err
		}
	default:
		sugar.Error(err)
		return nil, err
	}

	var adminCount int
	err = db.Select("count(*)").
		From("permission_membership").
		Where(sq.Eq{"permission": id}).
		Where(sq.Eq{"namespace": config.Namespace}).
		QueryRow().Scan(&adminCount)
	if err != nil {
		return nil, err
	}

	if adminCount == 0 {
		fmt.Println("No admins defined, please edit the config file and run chremoas-ctl reconfigure")
	}

	return &permissionsHandler{db: db, Logger: log, namespace: config.Namespace}, nil
}

func (h *permissionsHandler) Perform(_ context.Context, request *permSrv.PermissionsRequest, response *permSrv.PerformResponse) error {
	var (
		err   error
		sugar = h.Sugar()
	)

	_, err = h.db.Select("*").
		From("permissions").
		Join("permission_membership ON permissions.id = permission_membership.permission").
		Where(sq.Eq{"permissions.name": "server_admins"}).
		Where(sq.Eq{"permissions.namespace": h.namespace}).
		Where(sq.Eq{"permission_membership.user_id": request.User}).
		Query()

	if err == nil {
		// Doesn't matter what other permissions you have. If you are a server_admin you are god.
		response.CanPerform = true
		return nil
	} else if err != sql.ErrNoRows {
		// We don't wait to return an error if there are no rows, that just means not an admin
		sugar.Error(err)
		return err
	}

	for _, permission := range request.PermissionsList {
		_, err = h.db.Select("*").
			From("permissions").
			Join("permission_membership ON permissions.id = permission_membership.permission").
			Where(sq.Eq{"permissions.name": permission}).
			Where(sq.Eq{"permissions.namespace": h.namespace}).
			Where(sq.Eq{"permission_membership.user_id": request.User}).
			Query()

		if err == nil {
			// We found a permission this user is authorized for in the list
			response.CanPerform = true
			return nil
		} else if err != sql.ErrNoRows {
			// We don't wait to return an error if there are no rows, that just means not authorized
			sugar.Error(err)
			return err
		}
	}

	response.CanPerform = false
	return nil
}

func (h *permissionsHandler) AddPermission(_ context.Context, request *permSrv.Permission, response *permSrv.Permission) error {
	var (
		sugar = h.Sugar()
	)

	if request.Name == "server_admins" {
		return errCannotAdd
	}

	_, err := h.db.Insert("permissions").
		Columns("namespace", "name", "description").
		Values(h.namespace, request.Name, request.Description).
		Query()

	if err != nil {
		sugar.Error(err)
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) AddPermissionUser(_ context.Context, request *permSrv.PermissionUser, response *permSrv.PermissionUser) error {
	var (
		err   error
		id    int
		sugar = h.Sugar()
	)

	if request.Permission == "server_admins" {
		return errCannotAdd
	}

	err = h.db.Select("id").
		From("permissions").
		Where(sq.Eq{"name": request.Permission}).
		Where(sq.Eq{"namespace": h.namespace}).
		QueryRow().Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(fmt.Sprintf("no such permission: `%s`", request.Permission))
		}
		sugar.Error(err)
		return err
	}

	_, err = h.db.Insert("permission_membership").
		Columns("namespace", "permission", "user_id").
		Values(h.namespace, id, request.User).
		Query()
	if err != nil {
		sugar.Error(err)
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) RemovePermission(_ context.Context, request *permSrv.Permission, response *permSrv.Permission) error {
	var (
		err   error
		id    int
		sugar = h.Sugar()
	)

	if request.Name == "server_admins" {
		return errCannotDelete
	}

	err = h.db.Select("id").
		From("permissions").
		Where(sq.Eq{"name": request.Name}).
		Where(sq.Eq{"namespace": h.namespace}).
		QueryRow().Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(fmt.Sprintf("no such permission: `%s`", request.Name))
		}
		sugar.Error(err)
		return err
	}

	// Right now we just wack it. Will add a flag for that in the future.
	_, err = h.db.Delete("permission_membership").
		Where(sq.Eq{"permission": id}).Query()
	if err != nil {
		sugar.Error(err)
		return err
	}

	_, err = h.db.Delete("permissions").
		Where(sq.Eq{"id": id}).Query()
	if err != nil {
		sugar.Error(err)
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) RemovePermissionUser(_ context.Context, request *permSrv.PermissionUser, response *permSrv.PermissionUser) error {
	var (
		err   error
		id    int
		sugar = h.Sugar()
	)

	if request.Permission == "server_admins" {
		return errCannotRemove
	}

	err = h.db.Select("id").
		From("permissions").
		Where(sq.Eq{"name": request.Permission}).
		Where(sq.Eq{"namespace": h.namespace}).
		QueryRow().Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(fmt.Sprintf("no such permission: `%s`", request.Permission))
		}
		sugar.Error(err)
		return err
	}

	_, err = h.db.Delete("permission_membership").
		Where(sq.Eq{"permission": id}).
		Where(sq.Eq{"user_id": request.User}).
		Query()
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(fmt.Sprintf("no such user: `%s`", request.User))
		}
		sugar.Error(err)
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) ListPermissions(_ context.Context, _ *permSrv.NilRequest, response *permSrv.PermissionsResponse) error {
	var (
		name, description string
		sugar             = h.Sugar()
	)

	rows, err := h.db.Select("name", "description").
		From("permissions").
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		sugar.Error(err)
		return err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			sugar.Error(err)
		}
	}()

	for rows.Next() {
		err = rows.Scan(&name, &description)
		if err != nil {
			sugar.Error(err)
			continue
		}

		response.PermissionsList = append(response.PermissionsList,
			&permSrv.Permission{Name: name, Description: description})
	}

	if len(response.PermissionsList) == 0 {
		return errNoPermissions
	}

	return nil
}

func (h *permissionsHandler) ListPermissionUsers(_ context.Context, request *permSrv.UsersRequest, response *permSrv.UsersResponse) error {
	var (
		err   error
		sugar = h.Sugar()
	)

	rows, err := h.db.Select("user_id").
		From("permission_membership").
		Join("permissions ON permission_membership.permission = permissions.id").
		Where(sq.Eq{"permissions.name": request.Permission}).
		Where(sq.Eq{"permissions.namespace": h.namespace}).
		Query()
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(fmt.Sprintf("no such permission: `%s`", request.Permission))
		}
		sugar.Error(err)
		return err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			sugar.Error(err)
		}
	}()

	var userID int
	for rows.Next() {
		err = rows.Scan(&userID)
		if err != nil {
			sugar.Error(err)
			continue
		}
		response.UserList = append(response.UserList, fmt.Sprintf("%d", userID))
	}

	return nil
}

func (h *permissionsHandler) ListUserPermissions(_ context.Context, request *permSrv.PermissionUser, response *permSrv.PermissionsResponse) error {
	var (
		name, description string
		sugar             = h.Sugar()
	)

	rows, err := h.db.Select("permissions.name", "permissions.description").
		From("permissions").
		Join("permission_membership ON permission_membership.permission = permissions.id").
		Where(sq.Eq{"permission_membership.user_id": request.User[3 : len(request.User)-1]}).
		Where(sq.Eq{"permissions.namespace": h.namespace}).
		Query()
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(fmt.Sprintf("no such user: `%s`", request.User))
		}
		sugar.Error(err)
		return err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			sugar.Error(err)
		}
	}()

	for rows.Next() {
		err = rows.Scan(&name, &description)
		if err != nil {
			sugar.Error(err)
			continue
		}

		response.PermissionsList = append(response.PermissionsList,
			&permSrv.Permission{Name: name, Description: description})
	}

	if len(response.PermissionsList) == 0 {
		return errNoPermissions
	}

	return nil
}
