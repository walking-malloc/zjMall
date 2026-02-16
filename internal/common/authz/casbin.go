package authz

import (
	"log"
	"path/filepath"

	"github.com/casbin/casbin/v2"
)

var Enforcer *casbin.Enforcer

func InitCasbin() error {
	modelPath := filepath.Join("configs", "casbin_model.conf")
	policyPath := filepath.Join("configs", "casbin_policy.csv")

	e, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		return err
	}

	// 可选：预加载策略到内存
	if err := e.LoadPolicy(); err != nil {
		return err
	}

	Enforcer = e
	log.Println("✅ Casbin RBAC 初始化成功")
	return nil
}
