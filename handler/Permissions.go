package handler

import (
	"errors"
	"fmt"
	permsrv "github.com/chremoas/perms-srv/proto"
	redis "github.com/chremoas/services-common/redis"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
)

type permissionsHandler struct {
	Client      client.Client
	RedisClient *redis.Client
}

func NewPermissionsHandler() permsrv.PermissionsHandler {
	redisClient := redis.Init("localhost:6379", "", 0, "perms-srv")
	permissions := make(map[string][]int)

	_, err := redisClient.Ping()
	if err != nil {
		panic(err)
	}

	err = redisClient.Get("permissions", &permissions)
	if err == redis.Nil {
		fmt.Println("Permissions not set, please edit the config file and run chremoas-ctl reconfigure")
	} else if err != nil {
		fmt.Println(err)
	}

	return &permissionsHandler{RedisClient: redisClient}
}

func (h *permissionsHandler) Perform(ctx context.Context, request *permsrv.PermissionsRequest, response *permsrv.PerformResponse) error {
	return errors.New("Not Implemented")
}

func (h *permissionsHandler) AddPermission(ctx context.Context, request *permsrv.Permission, response *permsrv.Permission) error {
	permissions, err := h.getPermissions()

	if err != nil {
		return err
	}

	for perm := range permissions {
		if request.Name == perm {
			return fmt.Errorf("Permission group `%s` already exists.", perm)
		}
	}
	permissions[request.Name] = request.Description

	err = h.RedisClient.Set("permissions", permissions)
	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) RemovePermission(ctx context.Context, request *permsrv.Permission, response *permsrv.Permission) error {
	permissions, err := h.getPermissions()

	if err != nil {
		return err
	}

	if request.Name == "server_admins" {
		return errors.New("You cannot delete the server_admins group.")
	}

	// Maybe check it exists before we try to delete so we can return a useful message?
	delete(permissions, request.Name)

	err = h.RedisClient.Set("permissions", permissions)
	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) ListPermissions(ctx context.Context, request *permsrv.NilRequest, response *permsrv.PermissionsResponse) error {
	permissions, err := h.getPermissions()

	if err != nil {
		return err
	}

	for perm := range permissions {
		response.PermissionsList = append(response.PermissionsList,
			&permsrv.Permission{Name: perm, Description: permissions[perm]})
	}

	return nil
}

func (h *permissionsHandler) ListPermissionUsers(ctx context.Context, request *permsrv.UsersRequest, response *permsrv.UsersResponse) error {
	var userlist []int64
	users, err := h.getPermissionUsers(request.Permission)

	if err != nil {
		return err
	}

	for user := range users {
		userlist = append(userlist, users[user])
	}

	response.UserList = userlist
	return nil
}

func (h *permissionsHandler) ListUserPermissions(ctx context.Context, request *permsrv.PermissionsRequest, response *permsrv.PermissionsResponse) error {
	return errors.New("Not Implemented")
}

func (h *permissionsHandler) getPermissions() (map[string]string, error) {
	permissions := make(map[string]string)
	err := h.RedisClient.Get("permissions", &permissions)

	if err == redis.Nil {
		return nil, errors.New("Server Admin group not set, please edit the config file and run chremoas-ctl reconfigure")
	} else if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (h *permissionsHandler) getPermissionUsers(permission string) ([]int64, error) {
	var users []int64
	err := h.RedisClient.Get(permission, &users)

	if err == redis.Nil {
		return nil, fmt.Errorf("Permission group `%s` doesn't exist.", permission)
	} else if err != nil {
		return nil, err
	}

	return users, nil
}
