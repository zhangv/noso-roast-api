package service

import (
	"hahacking/service/user"
)

type ServiceGroup struct {
	user.UserServiceImpl
}

var ServiceGroupApp = new(ServiceGroup)

func NewUserService() user.UserService {
	return &user.UserServiceImpl{}
}
