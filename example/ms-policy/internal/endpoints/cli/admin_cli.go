package cli

import (
	"context"
	"fmt"
	"policy/internal/domain/dto"
	"policy/internal/domain/entities"
	"policy/internal/domain/usecase"
	"syscall"
)

type AdminCmdType int

const (
	CreateRoleCmd AdminCmdType = iota + 1
	CreatePermissionCmd
	AssignUserCmd
	InitCmd
)

func (act AdminCmdType) String() string {
	return [...]string{"create-role", "create-permission", "assign-user", "init"}[act-1]
}

func (act AdminCmdType) Index() int {
	return int(act)
}

type AdminCli struct {
	converter Converter
	cases     usecase.AdminCases
}

func New(cases usecase.AdminCases, converter Converter) *AdminCli {
	return &AdminCli{cases: cases, converter: converter}
}

func (cli *AdminCli) RunCmd(ctx context.Context, cmdName string, args ...string) error {
	defer func() {
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()
	var err error
	switch cmdName {
	case InitCmd.String():
		err = cli.init(ctx)
	case CreateRoleCmd.String():
		err = cli.createRole(ctx, args...)
	case CreatePermissionCmd.String():
		err = cli.createPermission(ctx, args...)
	case AssignUserCmd.String():
		err = cli.assignUser(ctx, args...)
	default:
		err = fmt.Errorf("unknown admin command: %v", cmdName)
	}
	return err
}

func (cli *AdminCli) createRole(ctx context.Context, args ...string) error {
	data, err := cli.converter.ConvertRoleCreate(args...)
	if err != nil {
		return err
	}
	return cli.cases.CreateRole(ctx, data)
}

func (cli *AdminCli) createPermission(ctx context.Context, args ...string) error {
	data, err := cli.converter.ConvertPermissionCreate(args...)
	if err != nil {
		return err
	}
	return cli.cases.CreatePermission(ctx, data)
}

func (cli *AdminCli) assignUser(ctx context.Context, args ...string) error {
	data, err := cli.converter.ConvertUserAssign(args...)
	if err != nil {
		return err
	}
	return cli.cases.AssignUser(ctx, data)
}

func (cli *AdminCli) init(ctx context.Context) error {
	err := cli.cases.CreateRole(
		ctx,
		dto.NewRoleCreate(entities.UserRole.String(), entities.UserRoleName),
	)
	if err != nil {
		return err
	}
	return cli.cases.CreateRole(
		ctx,
		dto.NewRoleCreate(entities.AdminRole.String(), entities.AdminRoleName),
	)
}
