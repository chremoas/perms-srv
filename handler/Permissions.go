package handler

import (
	"errors"
	"fmt"
	permsrv "github.com/chremoas/perms-srv/proto"
	redis "github.com/chremoas/services-common/redis"
	//"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	"strings"
)

type permissionsHandler struct {
	//Client client.Client
	Redis *redis.Client
}

func NewPermissionsHandler() permsrv.PermissionsHandler {
	redisClient := redis.Init("localhost:6379", "", 0, "perms-srv")

	_, err := redisClient.Client.Ping().Result()
	if err != nil {
		panic(err)
	}

	exists, err := redisClient.Client.Exists(redisClient.KeyName("permission:description:server_admins")).Result()

	if err != nil {
		fmt.Println(err)
	}

	if exists == 0 {
		fmt.Println("Permissions not setup, please edit the config file and run chremoas-ctl reconfigure")
	}

	admins, err := redisClient.Client.LRange(redisClient.KeyName("permission:members:server_admins"), 0, -1).Result()
	fmt.Printf("admins: %+v\n", admins)

	if len(admins) == 0 {
		fmt.Println("No admins defined, please edit the config file and run chremoas-ctl reconfigure")
	}

	return &permissionsHandler{Redis: redisClient}
}

func (h *permissionsHandler) Perform(ctx context.Context, request *permsrv.PermissionsRequest, response *permsrv.PerformResponse) error {
	return errors.New("Not Implemented")
}

func (h *permissionsHandler) AddPermission(ctx context.Context, request *permsrv.Permission, response *permsrv.Permission) error {
	permName := h.Redis.KeyName(fmt.Sprintf("permission:description:%s", request.Name))

	if request.Name == "server_admins" {
		return errors.New("You cannot add the server_admins group.")
	}

	exists, err := h.Redis.Client.Exists(permName).Result()

	if err != nil {
		fmt.Println(err)
	}

	if exists == 1 {
		return fmt.Errorf("Permission group `%s` already exists.", request.Name)
	}

	foo, err := h.Redis.Client.Set(permName, request.Description, 0).Result()
	fmt.Printf("Set: %+v\n", foo)

	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) AddPermissionUser(ctx context.Context, request *permsrv.PermissionUser, response *permsrv.PermissionUser) error {
	permName := h.Redis.KeyName(fmt.Sprintf("permission:members:%s", request.Permission))
	permDesc := h.Redis.KeyName(fmt.Sprintf("permission:description:%s", request.Permission))

	if request.Permission == "server_admins" {
		return errors.New("You cannot add users to the server_admins group.")
	}

	exists, err := h.Redis.Client.Exists(permDesc).Result()

	if err != nil {
		fmt.Println(err)
	}

	if exists == 0 {
		return fmt.Errorf("Permission group `%s` doesn't exists.", request.Permission)
	}

	foo, err := h.Redis.Client.SAdd(permName, request.User).Result()
	fmt.Printf("SAdd: %+v\n", foo)

	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) RemovePermission(ctx context.Context, request *permsrv.Permission, response *permsrv.Permission) error {
	permName := h.Redis.KeyName(fmt.Sprintf("permission:description:%s", request.Name))
	permMembers := h.Redis.KeyName(fmt.Sprintf("permission:members:%s", request.Name))

	if request.Name == "server_admins" {
		return errors.New("You cannot delete the server_admins group.")
	}

	exists, err := h.Redis.Client.Exists(permName).Result()

	if err != nil {
		fmt.Println(err)
	}

	if exists == 0 {
		return fmt.Errorf("Permission group `%s` doesn't exists.", request.Name)
	}

	members, err := h.Redis.Client.SMembers(permMembers).Result()

	fmt.Printf("Members: %+v\n", members)
	fmt.Printf("Member Count: %d\n", len(members))

	if len(members) > 0 {
		return fmt.Errorf("Permission group `%s` not empty.", request.Name)
	}

	foo, err := h.Redis.Client.Del(permName).Result()
	fmt.Printf("Del: %+v\n", foo)

	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) RemovePermissionUser(ctx context.Context, request *permsrv.PermissionUser, response *permsrv.PermissionUser) error {
	permName := h.Redis.KeyName(fmt.Sprintf("permission:members:%s", request.Permission))
	permDesc := h.Redis.KeyName(fmt.Sprintf("permission:description:%s", request.Permission))

	if request.Permission == "server_admins" {
		return errors.New("You cannot remove users from the server_admins group.")
	}

	exists, err := h.Redis.Client.Exists(permDesc).Result()

	if err != nil {
		fmt.Println(err)
	}

	if exists == 0 {
		return fmt.Errorf("Permission group `%s` doesn't exists.", request.Permission)
	}

	isMember, err := h.Redis.Client.SIsMember(permName, request.User).Result()
	if !isMember {
		return fmt.Errorf("`%s` not a member of group '%s'", request.User, request.Permission)
	}

	foo, err := h.Redis.Client.SRem(permName, request.User).Result()
	fmt.Printf("SRem: %+v\n", foo)

	if err != nil {
		return err
	}

	response = request
	return nil
}

func (h *permissionsHandler) ListPermissions(ctx context.Context, request *permsrv.NilRequest, response *permsrv.PermissionsResponse) error {
	perms, err := h.Redis.Client.Keys(h.Redis.KeyName("permission:description:*")).Result()

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
	permName := h.Redis.KeyName(fmt.Sprintf("permission:members:%s", request.Permission))

	users, err := h.Redis.Client.SMembers(permName).Result()

	if err != nil {
		return err
	}

	fmt.Printf("users: %+v\n", users)
	for user := range users {
		fmt.Printf("user: %+v\n", users[user])
		userlist = append(userlist, users[user])
	}

	response.UserList = userlist
	return nil
}

func (h *permissionsHandler) ListUserPermissions(ctx context.Context, request *permsrv.PermissionUser, response *permsrv.PermissionsResponse) error {
	perms, err := h.Redis.Client.Keys(h.Redis.KeyName("permission:members:*")).Result()

	if err != nil {
		return err
	}

	// This is expensive but shouldn't really matter as it won't be used all that much. -brian
	for perm := range perms {
		permName := strings.Split(perms[perm], ":")

		permDesc := h.Redis.KeyName(fmt.Sprintf("permission:description:%s", permName[len(permName)-1]))
		permDescription, err := h.Redis.Client.Get(permDesc).Result()

		if err != nil {
			return err
		}

		isMember, err := h.Redis.Client.SIsMember(perms[perm], request.User).Result()

		if err != nil {
			return err
		}

		if !isMember {
			response.PermissionsList = append(response.PermissionsList,
				&permsrv.Permission{Name: permName[len(permName)-1], Description: permDescription})
		}
	}

	return nil
}
