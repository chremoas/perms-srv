package handler

import (
	"errors"
	"fmt"
	permsrv "github.com/chremoas/perms-srv/proto"
	redis "github.com/chremoas/services-common/redis"
	common "github.com/chremoas/services-common/command"
	"golang.org/x/net/context"
	"strings"
	"github.com/chremoas/services-common/config"
)

type permissionsHandler struct {
	//Client client.Client
	Redis *redis.Client
}

func NewPermissionsHandler(config *config.Configuration) permsrv.PermissionsHandler {
	//addr := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
	redisClient := redis.Init("Dunno")

	_, err := redisClient.Client.Ping().Result()
	if err != nil {
		panic(err)
	}

	exists, err := redisClient.Client.Exists(redisClient.KeyName("description:server_admins")).Result()

	if err != nil {
		fmt.Println(err)
	}

	if exists == 0 {
		fmt.Println("Permissions not setup, please edit the config file and run chremoas-ctl reconfigure")
	}

	admins, err := redisClient.Client.SMembers(redisClient.KeyName("members:server_admins")).Result()

	if len(admins) == 0 {
		fmt.Println("No admins defined, please edit the config file and run chremoas-ctl reconfigure")
	}

	return &permissionsHandler{Redis: redisClient}
}

func (h *permissionsHandler) Perform(ctx context.Context, request *permsrv.PermissionsRequest, response *permsrv.PerformResponse) error {
	serverAdmins := h.Redis.KeyName("members:server_admins")
	isServerAdmin, err := h.Redis.Client.SIsMember(serverAdmins, request.User).Result()

	if err != nil {
		return err
	}

	// Doesn't matter what other permissions you have. If you are a server_admin you are god.
	if isServerAdmin {
		response.CanPerform = true
		return nil
	}

	for perm := range request.PermissionsList {
		permName := h.Redis.KeyName(fmt.Sprintf("members:%s", request.PermissionsList[perm]))
		isMember, err := h.Redis.Client.SIsMember(permName, request.User).Result()

		if err != nil {
			return err
		}

		if isMember {
			response.CanPerform = true
			return nil
		}
	}

	response.CanPerform = false
	return nil
}

func (h *permissionsHandler) AddPermission(ctx context.Context, request *permsrv.Permission, response *permsrv.Permission) error {
	permName := h.Redis.KeyName(fmt.Sprintf("description:%s", request.Name))

	if request.Name == "server_admins" {
		return errors.New("You cannot add the server_admins group.")
	}

	exists, err := h.Redis.Client.Exists(permName).Result()

	if err != nil {
		return err
	}

	if exists == 1 {
		return fmt.Errorf("Permission group `%s` already exists.", request.Name)
	}

	_, err = h.Redis.Client.Set(permName, request.Description, 0).Result()

	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) AddPermissionUser(ctx context.Context, request *permsrv.PermissionUser, response *permsrv.PermissionUser) error {
	permName := h.Redis.KeyName(fmt.Sprintf("members:%s", request.Permission))
	permDesc := h.Redis.KeyName(fmt.Sprintf("description:%s", request.Permission))

	if request.Permission == "server_admins" {
		return errors.New("You cannot add users to the server_admins group.")
	}

	exists, err := h.Redis.Client.Exists(permDesc).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Permission group `%s` doesn't exists.", request.Permission)
	}

	_, err = h.Redis.Client.SAdd(permName, request.User).Result()

	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) RemovePermission(ctx context.Context, request *permsrv.Permission, response *permsrv.Permission) error {
	permName := h.Redis.KeyName(fmt.Sprintf("description:%s", request.Name))
	permMembers := h.Redis.KeyName(fmt.Sprintf("members:%s", request.Name))

	if request.Name == "server_admins" {
		return errors.New("You cannot delete the server_admins group.")
	}

	exists, err := h.Redis.Client.Exists(permName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Permission group `%s` doesn't exists.", request.Name)
	}

	members, err := h.Redis.Client.SMembers(permMembers).Result()

	if len(members) > 0 {
		return fmt.Errorf("Permission group `%s` not empty.", request.Name)
	}

	_, err = h.Redis.Client.Del(permName).Result()

	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) RemovePermissionUser(ctx context.Context, request *permsrv.PermissionUser, response *permsrv.PermissionUser) error {
	permName := h.Redis.KeyName(fmt.Sprintf("members:%s", request.Permission))
	permDesc := h.Redis.KeyName(fmt.Sprintf("description:%s", request.Permission))

	if request.Permission == "server_admins" {
		return errors.New("You cannot remove users from the server_admins group.")
	}

	exists, err := h.Redis.Client.Exists(permDesc).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Permission group `%s` doesn't exists.", request.Permission)
	}

	isMember, err := h.Redis.Client.SIsMember(permName, request.User).Result()
	if !isMember {
		return fmt.Errorf("`%s` not a member of group '%s'", request.User, request.Permission)
	}

	_, err = h.Redis.Client.SRem(permName, request.User).Result()

	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) ListPermissions(ctx context.Context, request *permsrv.NilRequest, response *permsrv.PermissionsResponse) error {
	perms, err := h.Redis.Client.Keys(h.Redis.KeyName("description:*")).Result()

	if err != nil {
		return err
	}

	for perm := range perms {
		permDescription, err := h.Redis.Client.Get(perms[perm]).Result()

		if err != nil {
			return err
		}

		permName := strings.Split(perms[perm], ":")

		response.PermissionsList = append(response.PermissionsList,
			&permsrv.Permission{Name: permName[len(permName)-1], Description: permDescription})
	}

	return nil
}

func (h *permissionsHandler) ListPermissionUsers(ctx context.Context, request *permsrv.UsersRequest, response *permsrv.UsersResponse) error {
	var userlist []string
	permName := h.Redis.KeyName(fmt.Sprintf("members:%s", request.Permission))

	users, err := h.Redis.Client.SMembers(permName).Result()

	if err != nil {
		return err
	}

	for user := range users {
		userlist = append(userlist, users[user])
	}

	response.UserList = userlist
	return nil
}

func (h *permissionsHandler) ListUserPermissions(ctx context.Context, request *permsrv.PermissionUser, response *permsrv.PermissionsResponse) error {
	perms, err := h.Redis.Client.Keys(h.Redis.KeyName("members:*")).Result()

	if err != nil {
		return err
	}

	// This is expensive but shouldn't really matter as it won't be used all that much. -brian
	for perm := range perms {
		permName := strings.Split(perms[perm], ":")

		permDesc := h.Redis.KeyName(fmt.Sprintf("description:%s", permName[len(permName)-1]))
		permDescription, err := h.Redis.Client.Get(permDesc).Result()

		if err != nil {
			return err
		}

		userId := common.ExtractUserId(request.User)
		isMember, err := h.Redis.Client.SIsMember(perms[perm], userId).Result()

		if err != nil {
			return err
		}

		if isMember {
			response.PermissionsList = append(response.PermissionsList,
				&permsrv.Permission{Name: permName[len(permName)-1], Description: permDescription})
		}
	}

	return nil
}
